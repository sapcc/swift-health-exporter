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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// diskUsageTask implements the collector.collectorTask interface.
type diskUsageTask struct {
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	capacityBytes         *promhelper.TypedDesc
	freeBytes             *promhelper.TypedDesc
	usedBytes             *promhelper.TypedDesc
	fractionalUsage       *promhelper.TypedDesc
	fractionalUsageByDisk *promhelper.TypedDesc
}

func newDiskUsageTask(pathToReconExecutable string, hostTimeout int, ctxTimeout time.Duration) task {
	return &diskUsageTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		capacityBytes: promhelper.NewGaugeTypedDesc(
			"swift_cluster_storage_capacity_bytes",
			"Capacity storage bytes as reported by the swift-recon tool.", nil),
		freeBytes: promhelper.NewGaugeTypedDesc(
			"swift_cluster_storage_free_bytes",
			"Free storage bytes as reported by the swift-recon tool.", nil),
		usedBytes: promhelper.NewGaugeTypedDesc(
			"swift_cluster_storage_used_bytes",
			"Used storage bytes as reported by the swift-recon tool.", nil),
		fractionalUsage: promhelper.NewGaugeTypedDesc(
			// In order to be consistent with the legacy system, the metric
			// name uses the word percent instead of fractional.
			"swift_cluster_storage_used_percent",
			"Fractional usage as reported by the swift-recon tool.", nil),
		fractionalUsageByDisk: promhelper.NewGaugeTypedDesc(
			// In order to be consistent with the legacy system, the metric
			// name uses the word percent instead of fractional.
			"swift_cluster_storage_used_percent_by_disk",
			"Fractional usage of a disk as reported by the swift-recon tool.", []string{"storage_ip", "disk"}),
	}
}

// describeMetrics implements the task interface.
func (t *diskUsageTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.capacityBytes.Describe(ch)
	t.freeBytes.Describe(ch)
	t.usedBytes.Describe(ch)
	t.fractionalUsage.Describe(ch)
	t.fractionalUsageByDisk.Describe(ch)
}

var specialCharRx = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// collectMetrics implements the task interface.
func (t *diskUsageTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	exitCode := 0
	cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), "--diskusage", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
	if err == nil {
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
				exitCode = 1
				outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
				logg.Error("swift recon: %s: %s: %s: %s",
					cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
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
				device := specialCharRx.ReplaceAllLiteralString(disk.Device, "")
				diskUsageRatio := float64(disk.Used) / float64(disk.Size)
				ch <- t.fractionalUsageByDisk.MustNewConstMetric(diskUsageRatio, hostname, device)
			}
		}

		usageRatio := float64(totalUsed) / float64(totalSize)
		if totalSize == 0 {
			usageRatio = 1.0
		}

		ch <- t.fractionalUsage.MustNewConstMetric(usageRatio)
		ch <- t.usedBytes.MustNewConstMetric(float64(totalUsed))
		ch <- t.freeBytes.MustNewConstMetric(float64(totalFree))
		ch <- t.capacityBytes.MustNewConstMetric(float64(totalSize))
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}
