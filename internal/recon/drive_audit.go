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

// driveAuditTask implements the collector.collectorTask interface.
type driveAuditTask struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	auditErrors *promhelper.TypedDesc
}

func newDriveAuditTask(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &driveAuditTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		auditErrors: promhelper.NewGaugeTypedDesc(
			"swift_cluster_drives_audit_errors",
			"Drive audit errors reported by the swift-recon tool.", []string{"storage_ip"}),
	}
}

// describeMetrics implements the task interface.
func (t *driveAuditTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.auditErrors.Describe(ch)
}

// collectMetrics implements the task interface.
func (t *driveAuditTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	exitCode := 0
	cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), "--driveaudit", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var data struct {
				DriveAuditErrors int64 `json:"drive_audit_errors"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				exitCode = 1
				outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
				logg.Error("swift recon: %s: %s: %s: %s",
					cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
				continue // to next host
			}

			ch <- t.auditErrors.MustNewConstMetric(float64(data.DriveAuditErrors), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}
