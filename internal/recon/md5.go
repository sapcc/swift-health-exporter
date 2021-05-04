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
	"fmt"
	"regexp"
	"strconv"
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
			"Errors encountered while checking hosts for md5sum(s) as reported by the swift-recon tool.", []string{"kind"}),
		matched: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_matched",
			"Matched hosts for md5sum(s) reported by the swift-recon tool.", []string{"kind"}),
		notMatched: promhelper.NewGaugeTypedDesc(
			"swift_cluster_md5_not_matched",
			"Not matched hosts for md5sum(s) reported by the swift-recon tool.", []string{"kind"}),
	}
}

// describeMetrics implements the task interface.
func (t *md5Task) describeMetrics(ch chan<- *prometheus.Desc) {
	t.all.Describe(ch)
	t.errors.Describe(ch)
	t.matched.Describe(ch)
	t.notMatched.Describe(ch)
}

// Match group ref:
//  <1: kind> <2: output in case of error> <3: matched hosts> <4: total hosts> <5: errors>
var md5OutputRx = regexp.MustCompile(
	`(?m)^.* Checking ([\.a-zA-Z0-9_]+) md5sums?\s*((?:->\s\S*\s.*\s*)*)?([0-9]+)/([0-9]+) hosts matched, ([0-9]+) error.*$`)

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
		matchList = md5OutputRx.FindAllSubmatch(out, -1)
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
		if len(match[2]) > 0 {
			exitCode = 1
		}
		var totalHosts, errsEncountered float64
		matchedHosts, err := strconv.ParseFloat(string(match[3]), 64)
		if err == nil {
			totalHosts, err = strconv.ParseFloat(string(match[4]), 64)
			if err == nil {
				errsEncountered, err = strconv.ParseFloat(string(match[5]), 64)
				if err == nil {
					kind := strings.ReplaceAll(string(match[1]), ".", "")
					notMatchedHosts := totalHosts - matchedHosts
					all := matchedHosts + notMatchedHosts + errsEncountered

					ch <- t.all.MustNewConstMetric(all, kind)
					ch <- t.errors.MustNewConstMetric(errsEncountered, kind)
					ch <- t.matched.MustNewConstMetric(matchedHosts, kind)
					ch <- t.notMatched.MustNewConstMetric(notMatchedHosts, kind)
				}
			}
		}
		if err != nil {
			exitCode = 1
			logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		}
	}
}
