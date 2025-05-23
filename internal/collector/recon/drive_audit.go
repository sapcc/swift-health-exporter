// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package recon

import (
	"context"
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

// UpdateMetrics implements the collector.Task interface.
func (t *DriveAuditTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	q := util.CmdArgsToStr(t.cmdArgs)
	queries := map[string]int{q: 0}
	e := &collector.TaskError{
		Cmd:     "swift-recon",
		CmdArgs: t.cmdArgs,
	}

	outputPerHost, err := getSwiftReconOutputPerHost(ctx, t.opts.CtxTimeout, t.opts.PathToExecutable, t.cmdArgs...)
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
