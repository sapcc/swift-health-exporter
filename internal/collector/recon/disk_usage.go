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
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// DiskUsageTask implements the collector.Task interface.
type DiskUsageTask struct {
	opts    *TaskOpts
	cmdArgs []string

	specialCharRe *regexp.Regexp

	capacityBytes         prometheus.Gauge
	freeBytes             prometheus.Gauge
	usedBytes             prometheus.Gauge
	fractionalUsage       prometheus.Gauge
	fractionalUsageByDisk *prometheus.GaugeVec
}

// NewDiskUsageTask returns a collector.Task for DiskUsageTask.
func NewDiskUsageTask(opts *TaskOpts) collector.Task {
	return &DiskUsageTask{
		opts:          opts,
		cmdArgs:       []string{fmt.Sprintf("--timeout=%d", opts.HostTimeout), "--diskusage", "--verbose"},
		specialCharRe: regexp.MustCompile(`[^a-zA-Z0-9]+`),
		capacityBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_cluster_storage_capacity_bytes",
				Help: "Capacity storage bytes as reported by the swift-recon tool.",
			}),
		freeBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_cluster_storage_free_bytes",
				Help: "Free storage bytes as reported by the swift-recon tool.",
			}),
		usedBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_cluster_storage_used_bytes",
				Help: "Used storage bytes as reported by the swift-recon tool.",
			}),
		fractionalUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				// In order to be consistent with the legacy system, the metric
				// name uses the word percent instead of fractional.
				Name: "swift_cluster_storage_used_percent",
				Help: "Fractional usage as reported by the swift-recon tool.",
			}),
		fractionalUsageByDisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				// In order to be consistent with the legacy system, the metric
				// name uses the word percent instead of fractional.
				Name: "swift_cluster_storage_used_percent_by_disk",
				Help: "Fractional usage of a disk as reported by the swift-recon tool.",
			}, []string{"storage_ip", "disk"}),
	}
}

// Name implements the collector.Task interface.
func (t *DiskUsageTask) Name() string {
	return "recon-diskusage"
}

// DescribeMetrics implements the collector.Task interface.
func (t *DiskUsageTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.capacityBytes.Describe(ch)
	t.freeBytes.Describe(ch)
	t.usedBytes.Describe(ch)
	t.fractionalUsage.Describe(ch)
	t.fractionalUsageByDisk.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *DiskUsageTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.capacityBytes.Collect(ch)
	t.freeBytes.Collect(ch)
	t.usedBytes.Collect(ch)
	t.fractionalUsage.Collect(ch)
	t.fractionalUsageByDisk.Collect(ch)
}

// UpdateMetrics implements the collector.Task interface.
func (t *DiskUsageTask) UpdateMetrics() (map[string]int, error) {
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

	var totalFree, totalUsed, totalSize flexibleUint64
	for hostname, dataBytes := range outputPerHost {
		var disksData []struct {
			Device  string         `json:"device"`
			Avail   flexibleUint64 `json:"avail"`
			Mounted bool           `json:"mounted"`
			Used    flexibleUint64 `json:"used"`
			Size    flexibleUint64 `json:"size"`
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

		for _, disk := range disksData {
			if !(disk.Mounted) {
				continue // to next disk
			}

			totalFree += disk.Avail
			totalUsed += disk.Used
			totalSize += disk.Size

			// submit metrics by disk (only fractional usage, which is the
			// most useful for alerting)
			device := t.specialCharRe.ReplaceAllLiteralString(disk.Device, "")
			diskUsageRatio := float64(disk.Used) / float64(disk.Size)
			t.fractionalUsageByDisk.With(prometheus.Labels{
				"storage_ip": hostname,
				"disk":       device,
			}).Set(float64(diskUsageRatio))
		}
	}

	usageRatio := float64(totalUsed) / float64(totalSize)
	if totalSize == 0 {
		usageRatio = 1.0
	}

	t.fractionalUsage.Set(usageRatio)
	t.usedBytes.Set(float64(totalUsed))
	t.freeBytes.Set(float64(totalFree))
	t.capacityBytes.Set(float64(totalSize))

	return queries, nil
}
