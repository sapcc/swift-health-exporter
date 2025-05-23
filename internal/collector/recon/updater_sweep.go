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

// UpdaterSweepTask implements the collector.Task interface.
type UpdaterSweepTask struct {
	opts    *TaskOpts
	cmdArgs []string

	containerTime *prometheus.GaugeVec
	objectTime    *prometheus.GaugeVec
}

// NewUpdaterSweepTask returns a collector.Task for UpdaterSweepTask.
func NewUpdaterSweepTask(opts *TaskOpts) collector.Task {
	return &UpdaterSweepTask{
		opts: opts,
		// <server-type> gets substituted in UpdateMetrics().
		cmdArgs: []string{
			fmt.Sprintf("--timeout=%d", opts.HostTimeout), "<server-type>",
			"--updater", "--verbose",
		},
		containerTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_updater_sweep_time",
				Help: "Container updater sweep time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		objectTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_objects_updater_sweep_time",
				Help: "Object updater sweep time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
	}
}

// Name implements the collector.Task interface.
func (t *UpdaterSweepTask) Name() string {
	return "recon-updater-sweep-time"
}

// DescribeMetrics implements the collector.Task interface.
func (t *UpdaterSweepTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.containerTime.Describe(ch)
	t.objectTime.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *UpdaterSweepTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.containerTime.Collect(ch)
	t.objectTime.Collect(ch)
}

// UpdateMetrics implements the collector.Task interface.
func (t *UpdaterSweepTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	queries := make(map[string]int)
	serverTypes := []string{"container", "object"}
	for _, server := range serverTypes {
		cmdArgs := t.cmdArgs
		cmdArgs[1] = server
		q := util.CmdArgsToStr(cmdArgs)
		queries[q] = 0
		e := &collector.TaskError{
			Cmd:     "swift-recon",
			CmdArgs: cmdArgs,
		}

		outputPerHost, err := getSwiftReconOutputPerHost(ctx, t.opts.CtxTimeout, t.opts.PathToExecutable, cmdArgs...)
		if err != nil {
			queries[q] = 1
			e.Inner = err
			return queries, e
		}

		for hostname, dataBytes := range outputPerHost {
			var data struct {
				ContainerUpdaterSweepTime float64 `json:"container_updater_sweep"`
				ObjectUpdaterSweepTime    float64 `json:"object_updater_sweep"`
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

			val := data.ContainerUpdaterSweepTime
			gaugeVec := t.containerTime
			if server == "object" {
				val = data.ObjectUpdaterSweepTime
				gaugeVec = t.objectTime
			}

			gaugeVec.With(prometheus.Labels{"storage_ip": hostname}).Set(val)
		}
	}

	return queries, nil
}
