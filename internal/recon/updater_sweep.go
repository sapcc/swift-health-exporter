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

// updaterSweepTask implements the collector.collectorTask interface.
type updaterSweepTask struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	containerTime *promhelper.TypedDesc
	objectTime    *promhelper.TypedDesc
}

func newUpdaterSweepTask(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &updaterSweepTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		containerTime: promhelper.NewGaugeTypedDesc(
			"swift_cluster_containers_updater_sweep_time",
			"Container updater sweep time reported by the swift-recon tool.", []string{"storage_ip"}),
		objectTime: promhelper.NewGaugeTypedDesc(
			"swift_cluster_objects_updater_sweep_time",
			"Object updater sweep time reported by the swift-recon tool.", []string{"storage_ip"}),
	}
}

// describeMetrics implements the task interface.
func (t *updaterSweepTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.containerTime.Describe(ch)
	t.objectTime.Describe(ch)
}

// collectMetrics implements the task interface.
func (t *updaterSweepTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	serverTypes := []string{"container", "object"}
	for _, server := range serverTypes {
		exitCode := 0
		cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), server, "--updater", "--verbose"}
		outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
		if err == nil {
			for hostname, dataBytes := range outputPerHost {
				var data struct {
					ContainerUpdaterSweepTime float64 `json:"container_updater_sweep"`
					ObjectUpdaterSweepTime    float64 `json:"object_updater_sweep"`
				}
				err := json.Unmarshal(dataBytes, &data)
				if err != nil {
					exitCode = 1
					outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
					logg.Error("swift recon: %s: %s: %s: %s",
						cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
					continue // to next host
				}

				val := data.ContainerUpdaterSweepTime
				desc := t.containerTime
				if server == "object" {
					val = data.ObjectUpdaterSweepTime
					desc = t.objectTime
				}

				ch <- desc.MustNewConstMetric(val, hostname)
			}
		} else {
			exitCode = 1
			logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		}

		ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
	}
}
