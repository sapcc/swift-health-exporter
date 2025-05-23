// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package recon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// ReplicationTask implements the collector.Task interface.
type ReplicationTask struct {
	opts    *TaskOpts
	cmdArgs []string

	accountReplicationAge        *prometheus.GaugeVec
	accountReplicationDuration   *prometheus.GaugeVec
	containerReplicationAge      *prometheus.GaugeVec
	containerReplicationDuration *prometheus.GaugeVec
	objectReplicationAge         *prometheus.GaugeVec
	objectReplicationDuration    *prometheus.GaugeVec
}

// NewReplicationTask returns a collector.Task for ReplicationTask.
func NewReplicationTask(opts *TaskOpts) collector.Task {
	return &ReplicationTask{
		opts: opts,
		// <server-type> gets substituted in UpdateMetrics().
		cmdArgs: []string{
			fmt.Sprintf("--timeout=%d", opts.HostTimeout), "<server-type>",
			"--replication", "--verbose",
		},
		accountReplicationAge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_accounts_replication_age",
				Help: "Account replication age reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		accountReplicationDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_accounts_replication_duration",
				Help: "Account replication duration reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerReplicationAge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_replication_age",
				Help: "Container replication age reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerReplicationDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_replication_duration",
				Help: "Container replication duration reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		objectReplicationAge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_objects_replication_age",
				Help: "Object replication age reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		objectReplicationDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_objects_replication_duration",
				Help: "Object replication duration reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
	}
}

// Name implements the collector.Task interface.
func (t *ReplicationTask) Name() string {
	return "recon-replication"
}

// DescribeMetrics implements the collector.Task interface.
func (t *ReplicationTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.accountReplicationAge.Describe(ch)
	t.accountReplicationDuration.Describe(ch)
	t.containerReplicationAge.Describe(ch)
	t.containerReplicationDuration.Describe(ch)
	t.objectReplicationAge.Describe(ch)
	t.objectReplicationDuration.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *ReplicationTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.accountReplicationAge.Collect(ch)
	t.accountReplicationDuration.Collect(ch)
	t.containerReplicationAge.Collect(ch)
	t.containerReplicationDuration.Collect(ch)
	t.objectReplicationAge.Collect(ch)
	t.objectReplicationDuration.Collect(ch)
}

// UpdateMetrics implements the collector.Task interface.
func (t *ReplicationTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	queries := make(map[string]int)
	serverTypes := []string{"account", "container", "object"}
	for _, server := range serverTypes {
		var ageTypedDesc, durTypedDesc *prometheus.GaugeVec
		switch server {
		case "account":
			ageTypedDesc = t.accountReplicationAge
			durTypedDesc = t.accountReplicationDuration
		case "container":
			ageTypedDesc = t.containerReplicationAge
			durTypedDesc = t.containerReplicationDuration
		case "object":
			ageTypedDesc = t.objectReplicationAge
			durTypedDesc = t.objectReplicationDuration
		}

		cmdArgs := t.cmdArgs
		cmdArgs[1] = server
		q := util.CmdArgsToStr(cmdArgs)
		queries[q] = 0
		e := &collector.TaskError{
			Cmd:     "swift-recon",
			CmdArgs: cmdArgs,
		}

		currentTime := float64(time.Now().Unix())
		outputPerHost, err := getSwiftReconOutputPerHost(ctx, t.opts.CtxTimeout, t.opts.PathToExecutable, cmdArgs...)
		if err != nil {
			queries[q] = 1
			e.Inner = err
			return queries, e
		}

		for hostname, dataBytes := range outputPerHost {
			var data struct {
				ReplicationLast flexibleFloat64 `json:"replication_last"`
				ReplicationTime flexibleFloat64 `json:"replication_time"`
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

			l := prometheus.Labels{"storage_ip": hostname}
			if data.ReplicationLast > 0 {
				if IsTest {
					currentTime = float64(timeNow().Second())
				}
				tDiff := currentTime - float64(data.ReplicationLast)
				ageTypedDesc.With(l).Set(tDiff)
			}
			durTypedDesc.With(l).Set(float64(data.ReplicationTime))
		}
	}

	return queries, nil
}
