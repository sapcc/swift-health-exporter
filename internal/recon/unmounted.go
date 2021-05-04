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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// unmountedTask implements the collector.collectorTask interface.
type unmountedTask struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	unmountedDrives *promhelper.TypedDesc
}

func newUnmountedTask(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &unmountedTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		unmountedDrives: promhelper.NewGaugeTypedDesc(
			"swift_cluster_drives_unmounted",
			"Unmounted drives reported by the swift-recon tool.", []string{"storage_ip"}),
	}
}

// describeMetrics implements the task interface.
func (t *unmountedTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.unmountedDrives.Describe(ch)
}

// collectMetrics implements the task interface.
func (t *unmountedTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	exitCode := 0
	cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), "--unmounted", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var disksData []struct {
				Device string `json:"device"`
			}
			err := json.Unmarshal(dataBytes, &disksData)
			if err != nil {
				exitCode = 1
				outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
				logg.Error("swift recon: %s: %s: %s: %s",
					cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
				continue // to next host
			}

			ch <- t.unmountedDrives.MustNewConstMetric(float64(len(disksData)), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}
