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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// md5Task implements the collector.collectorTask interface.
type md5Task struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	all        *promhelper.TypedDesc
	errors     *promhelper.TypedDesc
	matched    *promhelper.TypedDesc
	notMatched *promhelper.TypedDesc
}

func newMD5Task(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &md5Task{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		all: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_all",
			"Sum of matched-, not matched hosts, and errors encountered while check hosts for md5sum(s) as reported by the swift-recon tool.", []string{"kind"}),
		errors: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_errors",
			"Errors encountered while checking hosts for md5sum(s) as reported by the swift-recon tool.", []string{"storage_ip", "kind"}),
		matched: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_matched",
			"Matched hosts for md5sum(s) reported by the swift-recon tool.", []string{"storage_ip", "kind"}),
		notMatched: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_not_matched",
			"Not matched hosts for md5sum(s) reported by the swift-recon tool.", []string{"storage_ip", "kind"}),
	}
}

// describeMetrics implements the task interface.
func (t *md5Task) describeMetrics(ch chan<- *prometheus.Desc) {
	t.all.Describe(ch)
	t.errors.Describe(ch)
	t.matched.Describe(ch)
	t.notMatched.Describe(ch)
}

// md5OutputBlockRx extracts the entire output block for a specific kind from
// the aggregate md5 recon output.
//
// Match group ref:
//  <1: kind> <2: output block>
//
// Example of an output block for ring:
//   [<time-stamp>] Checking ring md5sums
//   -> On disk object.ring.gz md5sum: 123456
//   -> http://10.0.0.1:6000/recon/ringmd5: <urlopen error [Errno 111] ECONNREFUSED>
//   -> http://10.0.0.2:6000/recon/ringmd5: {'/path/to/account.ring.gz': '123456', '/path/to/container.ring.gz': '123456', '/path/to/object.ring.gz': '123456'}
//   -> http://10.0.0.2:6000/recon/ringmd5 matches.
// 1/2 hosts matched, 1 error[s] while checking hosts.
var md5OutputBlockRx = regexp.MustCompile(
	`(?m)^.* Checking ([\.a-zA-Z0-9_]+) md5sums?\s*((?:(?:->|!!)\s\S*\s.*\s*)*)\s*[0-9]+/[0-9]+ hosts matched, [0-9]+ error.*$`)

// collectMetrics implements the task interface.
func (t *md5Task) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	exitCode := 0
	cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), "--md5"}
	defer func() {
		ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
	}()

	var matchList [][][]byte
	out, err := runCommandWithTimeout(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		matchList = md5OutputBlockRx.FindAllSubmatch(out, -1)
		if len(matchList) == 0 {
			err = fmt.Errorf("command did not return any usable output:\n%s", string(out))
		}
	}
	if err != nil {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		return
	}

	for _, match := range matchList {
		kind := string(match[1])
		outputBlock := match[2]
		outputPerHost, err := splitOutputPerHost(outputBlock, cmdArgs)
		if err != nil {
			exitCode = 1
			logg.Error("swift recon: %s: %s: output follows:\n%s", cmdArgsToStr(cmdArgs), err.Error(), string(outputBlock))
			continue // to next output block
		}

		// allHostsMatch and errHosts are used to submit zero value metrics so
		// that we can use aggregation rules in case that MD5 of all hosts for
		// a specific kind matches.
		allHostsMatch := true
		// We use a map for errHosts to maintain a record of unique hostnames
		// that had errors since there could be multiple errors per host.
		errHosts := make(map[string]bool)
		var all float64
		for hostname, dataBytes := range outputPerHost {
			// Host output can be in the following formats:
			//   1. {'/path/to/<server-type>.ring.gz': '<md5-sum>', ...}
			//   2. "matches."
			//   3. "(/path/to/<server-type>.ring.gz => <md5-sum>) doesn't match on disk md5sum"
			//   4. "some error message"
			if json.Valid(dataBytes) {
				// We have output scenario 1 which we can skip.
				continue
			}

			str := string(dataBytes)
			switch {
			case strings.HasSuffix(str, "matches."):
				ch <- t.matched.MustNewConstMetric(1, hostname, kind)
				all++
			case strings.Contains(str, `doesn"t match`): // func splitOutputPerHost() changes ' -> "
				allHostsMatch = false
				ch <- t.notMatched.MustNewConstMetric(1, hostname, kind)
				all++
			default:
				errHosts[hostname] = true
				exitCode = 1
				logg.Error("swift recon: %s: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, kind, str)
			}
		}
		for hostname := range errHosts {
			ch <- t.errors.MustNewConstMetric(1, hostname, kind)
			all++
		}
		ch <- t.all.MustNewConstMetric(all, kind)

		if allHostsMatch {
			ch <- t.notMatched.MustNewConstMetric(0, "", kind)
		}
		if len(errHosts) == 0 {
			ch <- t.errors.MustNewConstMetric(0, "", kind)
		}
	}
}
