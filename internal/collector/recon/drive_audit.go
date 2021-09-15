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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// DriveAuditTask implements the collector.Task interface.
type DriveAuditTask struct {
	opts    *TaskOpts
	cmdArgs []string

	auditErrors *prometheus.GaugeVec
}

// NewDriveAuditTask returns a collector.Task for DriveAuditTask.
func NewDriveAuditTask(opts *TaskOpts) collector.Task {
	return &DriveAuditTask{
		opts:    opts,
		cmdArgs: []string{fmt.Sprintf("--timeout=%d", opts.HostTimeout), "--driveaudit", "--verbose"},
		auditErrors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_drives_audit_errors",
				Help: "Drive audit errors reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
	}
}

// Name implements the collector.Task interface.
func (t *DriveAuditTask) Name() string {
	return "recon-driveaudit"
}

// DescribeMetrics implements the collector.Task interface.
func (t *DriveAuditTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.auditErrors.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *DriveAuditTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.auditErrors.Collect(ch)
}

// Measure implements the collector.Task interface.
func (t *DriveAuditTask) Measure() (map[string]int, error) {
	q := util.CmdArgsToStr(t.cmdArgs)
	queries := map[string]int{q: 0}
	e := &collector.TaskError{
		Cmd:     "swift-recon",
		CmdArgs: t.cmdArgs,
	}

	outputPerHost, err := getSwiftReconOutputPerHost(t.opts.CtxTimeout, t.opts.PathToExecutable, t.cmdArgs...)
	if err != nil {
		queries[q] = 1
		e.Inner = err
		return queries, e
	}

	for hostname, dataBytes := range outputPerHost {
		var data struct {
			DriveAuditErrors int64 `json:"drive_audit_errors"`
		}
		err := json.Unmarshal(dataBytes, &data)
		if err != nil {
			queries[q] = 1
			e.Inner = err
			e.Hostname = hostname
			e.CmdOutput = string(dataBytes)
			logg.Info(e.Error())
			continue // to next host
		}

		t.auditErrors.With(prometheus.Labels{
			"storage_ip": hostname,
		}).Set(float64(data.DriveAuditErrors))
	}

	return queries, nil
}
