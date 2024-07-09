// Copyright 2019 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package recon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// MD5Task implements the collector.Task interface.
type MD5Task struct {
	opts    *TaskOpts
	cmdArgs []string

	all        *prometheus.GaugeVec
	errors     *prometheus.GaugeVec
	matched    *prometheus.GaugeVec
	notMatched *prometheus.GaugeVec
}

// NewMD5Task returns a collector.Task for MD5Task.
func NewMD5Task(opts *TaskOpts) collector.Task {
	return &MD5Task{
		opts:    opts,
		cmdArgs: []string{fmt.Sprintf("--timeout=%d", opts.HostTimeout), "--md5", "--verbose"},
		all: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_md5_all",
				Help: "Sum of matched-, not matched, and errored hosts while checking md5sum(s) as reported by the swift-recon tool.",
			}, []string{"kind"}),
		errors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_md5_errors",
				Help: "Error encountered while checking host for md5sum(s) as reported by the swift-recon tool.",
			}, []string{"storage_ip", "kind"}),
		matched: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_md5_matched",
				Help: "Matched host for md5sum(s) reported by the swift-recon tool.",
			}, []string{"storage_ip", "kind"}),
		notMatched: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_md5_not_matched",
				Help: "Not matched host for md5sum(s) reported by the swift-recon tool.",
			}, []string{"storage_ip", "kind"}),
	}
}

// Name implements the collector.Task interface.
func (t *MD5Task) Name() string {
	return "recon-md5"
}

// DescribeMetrics implements the collector.Task interface.
func (t *MD5Task) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.all.Describe(ch)
	t.errors.Describe(ch)
	t.matched.Describe(ch)
	t.notMatched.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *MD5Task) CollectMetrics(ch chan<- prometheus.Metric) {
	t.all.Collect(ch)
	t.errors.Collect(ch)
	t.matched.Collect(ch)
	t.notMatched.Collect(ch)
}

// md5OutputBlockRx extracts the entire output block for a specific kind from
// the aggregate md5 recon output.
//
// Match group ref:
//
//	<1: kind> <2: output block>
//
// Example of an output block for ring:
//
//	[<time-stamp>] Checking ring md5sums
//	-> On disk object.ring.gz md5sum: 123456
//	-> http://10.0.0.1:6000/recon/ringmd5: <urlopen error [Errno 111] ECONNREFUSED>
//	-> http://10.0.0.2:6000/recon/ringmd5: {'/path/to/account.ring.gz': '123456', '/path/to/container.ring.gz': '123456', '/path/to/object.ring.gz': '123456'}
//	-> http://10.0.0.2:6000/recon/ringmd5 matches.
//
// 1/2 hosts matched, 1 error[s] while checking hosts.
var md5OutputBlockRx = regexp.MustCompile(
	`(?m)^.* Checking ([\.a-zA-Z0-9_]+) md5sums?\s*((?:(?:->|!!).*\n)*)\s*[0-9]+/[0-9]+ hosts matched, [0-9]+ error.*$`)

// UpdateMetrics implements the collector.Task interface.
func (t *MD5Task) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	q := util.CmdArgsToStr(t.cmdArgs)
	queries := map[string]int{q: 0}
	e := &collector.TaskError{
		Cmd:     "swift-recon",
		CmdArgs: t.cmdArgs,
	}

	var matchList [][][]byte
	out, err := util.RunCommandWithTimeout(ctx, t.opts.CtxTimeout, t.opts.PathToExecutable, t.cmdArgs...)
	if err == nil {
		matchList = md5OutputBlockRx.FindAllSubmatch(out, -1)
		if len(matchList) == 0 {
			e.CmdOutput = string(out)
			err = errors.New("command did not return any usable output")
		}
	}
	if err != nil {
		queries[q] = 1
		e.Inner = err
		return queries, e
	}

	for _, match := range matchList {
		kind := string(match[1])
		outputBlock := match[2]
		outputPerHost, err := splitOutputPerHost(outputBlock, t.cmdArgs)
		if err != nil {
			queries[q] = 1
			e.Inner = err
			e.CmdOutput = string(outputBlock)
			logg.Info(e.Error())
			continue // to next output block
		}

		// processedErrHost is a map that contains a list of hosts for which
		// the error metric has already been submitted. We use this map to
		// avoid submitting duplicate metrics since there could be multiple
		// errors per host.
		processedErrHost := make(map[string]bool)
		var all float64
		for hostname, dataBytes := range outputPerHost {
			// Host output can be in the following formats:
			//   1. {"/path/to/<server-type>.ring.gz": "<md5-sum>", ...}
			//   2. "matches."
			//   3. "(/path/to/<server-type>.ring.gz => <md5-sum>) doesn't match on disk md5sum"
			//   4. "some error message"
			if json.Valid(dataBytes) {
				// We have output scenario 1 which we can skip.
				continue
			}

			str := string(dataBytes)
			var matched, notMatched, errored float64
			switch {
			case strings.HasSuffix(str, "matches."):
				matched = 1
				all++
			case strings.Contains(str, `doesn"t match`): // func splitOutputPerHost() changes ' -> "
				notMatched = 1
				all++
			default:
				if processedErrHost[hostname] {
					continue // to next host
				}
				queries[q] = 1
				errored = 1
				all++
				processedErrHost[hostname] = true
			}

			l := prometheus.Labels{"storage_ip": hostname, "kind": kind}
			t.matched.With(l).Set(matched)
			t.notMatched.With(l).Set(notMatched)
			t.errors.With(l).Set(errored)
		}
		t.all.With(prometheus.Labels{"kind": kind}).Set(all)
	}

	return queries, nil
}
