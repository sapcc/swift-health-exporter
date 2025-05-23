// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package recon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"

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
func (t *DiskUsageTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
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

	var totalFree, totalUsed, totalSize flexibleFloat64
	for hostname, dataBytes := range outputPerHost {
		var disksData []struct {
			Device  string          `json:"device"`
			Avail   flexibleFloat64 `json:"avail"`
			Mounted bool            `json:"mounted"`
			Used    flexibleFloat64 `json:"used"`
			Size    flexibleFloat64 `json:"size"`
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
			}).Set(diskUsageRatio)
		}
	}

	if rawCapStr := os.Getenv("SWIFT_CLUSTER_RAW_CAPACITY_BYTES"); rawCapStr != "" {
		rawCap, err := strconv.ParseFloat(rawCapStr, 64)
		if err != nil {
			logg.Error("could not parse 'SWIFT_CLUSTER_RAW_CAPACITY_BYTES' value: %s", err.Error())
		} else {
			totalSize = flexibleFloat64(rawCap)
		}
	}

	usageRatio := float64(totalUsed) / float64(totalSize)
	// The usageRatio value can be greater than 1:
	// 1. if the manually given total capacity (SWIFT_CLUSTER_RAW_CAPACITY_BYTES) is less than
	//    the total capacity reported by 'swift-recon' tool
	// 2. and the usage is greater than the manually given total capacity.
	if totalSize == 0 || usageRatio > 1 {
		usageRatio = 1.0
	}

	t.fractionalUsage.Set(usageRatio)
	t.usedBytes.Set(float64(totalUsed))
	t.freeBytes.Set(float64(totalFree))
	t.capacityBytes.Set(float64(totalSize))

	return queries, nil
}
