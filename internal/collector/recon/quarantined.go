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
	"context"
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// QuarantinedTask implements the collector.Task interface.
type QuarantinedTask struct {
	opts    *TaskOpts
	cmdArgs []string

	accounts   *prometheus.GaugeVec
	containers *prometheus.GaugeVec
	objects    *prometheus.GaugeVec
}

// NewQuarantinedTask returns a collector.Task for QurantinedTask.
func NewQuarantinedTask(opts *TaskOpts) collector.Task {
	return &QuarantinedTask{
		opts:    opts,
		cmdArgs: []string{fmt.Sprintf("--timeout=%d", opts.HostTimeout), "--quarantined", "--verbose"},
		accounts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_accounts_quarantined",
				Help: "Quarantined accounts reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_quarantined",
				Help: "Quarantined containers reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		objects: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_objects_quarantined",
				Help: "Quarantined objects reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
	}
}

// Name implements the collector.Task interface.
func (t *QuarantinedTask) Name() string {
	return "recon-quarantined"
}

// DescribeMetrics implements the collector.Task interface.
func (t *QuarantinedTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.accounts.Describe(ch)
	t.containers.Describe(ch)
	t.objects.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *QuarantinedTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.accounts.Collect(ch)
	t.containers.Collect(ch)
	t.objects.Collect(ch)
}

// UpdateMetrics implements the collector.Task interface.
func (t *QuarantinedTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
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
			Objects    int64 `json:"objects"`
			Accounts   int64 `json:"accounts"`
			Containers int64 `json:"containers"`
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
		t.accounts.With(l).Set(float64(data.Accounts))
		t.containers.With(l).Set(float64(data.Containers))
		t.objects.With(l).Set(float64(data.Objects))
	}

	return queries, nil
}
