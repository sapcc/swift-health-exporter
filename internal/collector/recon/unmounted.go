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

// UnmountedTask implements the collector.Task interface.
type UnmountedTask struct {
	opts    *TaskOpts
	cmdArgs []string

	unmountedDrives *prometheus.GaugeVec
}

// NewUnmountedTask returns a collector.Task for UnmountedTask.
func NewUnmountedTask(opts *TaskOpts) collector.Task {
	return &UnmountedTask{
		opts:    opts,
		cmdArgs: []string{fmt.Sprintf("--timeout=%d", opts.HostTimeout), "--unmounted", "--verbose"},
		unmountedDrives: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_drives_unmounted",
				Help: "Unmounted drives reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
	}
}

// Name implements the collector.Task interface.
func (t *UnmountedTask) Name() string {
	return "recon-unmounted"
}

// DescribeMetrics implements the collector.Task interface.
func (t *UnmountedTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.unmountedDrives.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *UnmountedTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.unmountedDrives.Collect(ch)
}

// Measure implements the collector.Task interface.
func (t *UnmountedTask) Measure() (map[string]int, error) {
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
		var disksData []struct {
			Device string `json:"device"`
		}
		err := json.Unmarshal(dataBytes, &disksData)
		if err != nil {
			queries[q] = 1
			e.Inner = err
			e.Hostname = hostname
			e.CmdOutput = string(dataBytes)
			logg.Info(e.Error())
			continue // to next host
		}

		t.unmountedDrives.With(prometheus.Labels{"storage_ip": hostname}).
			Set(float64(len(disksData)))
	}

	return queries, nil
}
