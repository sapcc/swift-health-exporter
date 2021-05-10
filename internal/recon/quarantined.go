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

// quarantinedTask implements the collector.collectorTask interface.
type quarantinedTask struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	accounts   *promhelper.TypedDesc
	containers *promhelper.TypedDesc
	objects    *promhelper.TypedDesc
}

func newQuarantinedTask(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &quarantinedTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		accounts: promhelper.NewGaugeTypedDesc(
			"swift_cluster_accounts_quarantined",
			"Quarantined accounts reported by the swift-recon tool.", []string{"storage_ip"}),
		containers: promhelper.NewGaugeTypedDesc(
			"swift_cluster_containers_quarantined",
			"Quarantined containers reported by the swift-recon tool.", []string{"storage_ip"}),
		objects: promhelper.NewGaugeTypedDesc(
			"swift_cluster_objects_quarantined",
			"Quarantined objects reported by the swift-recon tool.", []string{"storage_ip"}),
	}
}

// describeMetrics implements the task interface.
func (t *quarantinedTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.accounts.Describe(ch)
	t.containers.Describe(ch)
	t.objects.Describe(ch)
}

// collectMetrics implements the task interface.
func (t *quarantinedTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	exitCode := 0
	cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), "--quarantined", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var data struct {
				Objects    int64 `json:"objects"`
				Accounts   int64 `json:"accounts"`
				Containers int64 `json:"containers"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				exitCode = 1
				outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
				logg.Error("swift recon: %s: %s: %s: %s",
					cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
				continue // to next host
			}

			ch <- t.accounts.MustNewConstMetric(float64(data.Accounts), hostname)
			ch <- t.containers.MustNewConstMetric(float64(data.Containers), hostname)
			ch <- t.objects.MustNewConstMetric(float64(data.Objects), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}
