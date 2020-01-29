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

package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

// ReconCollector implements the prometheus.Collector interface.
type ReconCollector struct {
	taskExitCode typedDesc
	tasks        []collectorTask
}

// NewReconCollector creates a new ReconCollector.
func NewReconCollector(pathToExecutable string) *ReconCollector {
	return &ReconCollector{
		taskExitCode: typedDesc{
			desc: prometheus.NewDesc("swift_recon_task_exit_code",
				"The exit code for a Swift Recon query execution.",
				[]string{"query"}, nil),
			valueType: prometheus.GaugeValue,
		},
		tasks: []collectorTask{
			newReconDiskUsageTask(pathToExecutable),
			newReconDriveAuditTask(pathToExecutable),
			newReconMD5Task(pathToExecutable),
			newReconQuarantinedTask(pathToExecutable),
			newReconReplicationTask(pathToExecutable),
			newReconUnmountedTask(pathToExecutable),
			newReconUpdaterSweepTask(pathToExecutable),
		},
	}
}

// Describe implements the prometheus.Collector interface.
func (c *ReconCollector) Describe(ch chan<- *prometheus.Desc) {
	c.taskExitCode.describe(ch)

	for _, t := range c.tasks {
		t.describeMetrics(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *ReconCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.tasks))
	for _, t := range c.tasks {
		go func(t collectorTask) {
			t.collectMetrics(ch, c.taskExitCode)
			wg.Done()
		}(t)
	}
	wg.Wait()
}

///////////////////////////////////////////////////////////////////////////////
// Recon collector tasks.

// reconDiskUsageTask implements the collector.collectorTask interface.
type reconDiskUsageTask struct {
	pathToReconExecutable string
	specialCharRx         *regexp.Regexp

	capacityBytes         typedDesc
	freeBytes             typedDesc
	usedBytes             typedDesc
	fractionalUsage       typedDesc
	fractionalUsageByDisk typedDesc
}

func newReconDiskUsageTask(pathToReconExecutable string) collectorTask {
	return &reconDiskUsageTask{
		pathToReconExecutable: pathToReconExecutable,
		specialCharRx:         regexp.MustCompile(`[^a-zA-Z0-9]+`),
		capacityBytes: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_storage_capacity_bytes",
				"Capacity storage bytes as reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		freeBytes: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_storage_free_bytes",
				"Free storage bytes as reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		usedBytes: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_storage_used_bytes",
				"Used storage bytes as reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		fractionalUsage: typedDesc{
			// In order to be consistent with the legacy system, the metric
			// name uses the word percent instead of fractional.
			desc: prometheus.NewDesc("swift_cluster_storage_used_percent",
				"Fractional usage as reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		fractionalUsageByDisk: typedDesc{
			// In order to be consistent with the legacy system, the metric
			// name uses the word percent instead of fractional.
			desc: prometheus.NewDesc("swift_cluster_storage_used_percent_by_disk",
				"Fractional usage of a disk as reported by the swift-recon tool.",
				[]string{"storage_ip", "disk"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconDiskUsageTask implements the collector.collectorTask interface.
func (t *reconDiskUsageTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.capacityBytes.describe(ch)
	t.freeBytes.describe(ch)
	t.usedBytes.describe(ch)
	t.fractionalUsage.describe(ch)
	t.fractionalUsageByDisk.describe(ch)
}

// reconDiskUsageTask implements the collector.collectorTask interface.
func (t *reconDiskUsageTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArgs := []string{"--diskusage", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var disksData []struct {
				Device  string `json:"device"`
				Avail   int64  `json:"avail"`
				Mounted bool   `json:"mounted"`
				Used    int64  `json:"used"`
				Size    int64  `json:"size"`
			}
			err := json.Unmarshal(dataBytes, &disksData)
			if err != nil {
				exitCode = 1
				logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
				continue // to next host
			}

			var totalFree, totalUsed, totalSize int64
			for _, disk := range disksData {
				if !(disk.Mounted) {
					continue // to next disk
				}

				totalFree += disk.Avail
				totalUsed += disk.Used
				totalSize += disk.Size

				// submit metrics by disk (only fractional usage, which is the
				// most useful for alerting)
				device := t.specialCharRx.ReplaceAllLiteralString(disk.Device, "")
				diskUsageRatio := float64(disk.Used) / float64(disk.Size)
				ch <- t.fractionalUsageByDisk.mustNewConstMetric(diskUsageRatio, hostname, device)
			}

			usageRatio := float64(totalUsed) / float64(totalSize)
			if totalSize == 0 {
				usageRatio = 1.0
			}
			ch <- t.fractionalUsage.mustNewConstMetric(usageRatio, hostname)
			ch <- t.usedBytes.mustNewConstMetric(float64(totalUsed), hostname)
			ch <- t.freeBytes.mustNewConstMetric(float64(totalFree), hostname)
			ch <- t.capacityBytes.mustNewConstMetric(float64(totalSize), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}

// reconMD5Task implements the collector.collectorTask interface.
type reconMD5Task struct {
	pathToReconExecutable string
	md5OutputRx           *regexp.Regexp

	all        typedDesc
	errors     typedDesc
	matched    typedDesc
	notMatched typedDesc
}

func newReconMD5Task(pathToReconExecutable string) collectorTask {
	return &reconMD5Task{
		pathToReconExecutable: pathToReconExecutable,
		md5OutputRx: regexp.MustCompile(
			`(?m)^.* Checking ([\.a-zA-Z0-9_]+) md5sum(?:s)?\s*([0-9]+)/([0-9]+) hosts matched, ([0-9]+) error.*$`),
		all: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_md5_all",
				"Sum of matched-, not matched hosts, and errors encountered while check hosts for md5sum(s) as reported by the swift-recon tool.",
				[]string{"kind"}, nil),
			valueType: prometheus.GaugeValue,
		},
		errors: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_md5_errors",
				"Errors encountered while checking hosts for md5sum(s) as reported by the swift-recon tool.",
				[]string{"kind"}, nil),
			valueType: prometheus.GaugeValue,
		},
		matched: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_md5_matched",
				"Matched hosts for md5sum(s) reported by the swift-recon tool.",
				[]string{"kind"}, nil),
			valueType: prometheus.GaugeValue,
		},
		notMatched: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_md5_not_matched",
				"Not matched hosts for md5sum(s) reported by the swift-recon tool.",
				[]string{"kind"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconMD5Task implements the collector.collectorTask interface.
func (t *reconMD5Task) describeMetrics(ch chan<- *prometheus.Desc) {
	t.all.describe(ch)
	t.errors.describe(ch)
	t.matched.describe(ch)
	t.notMatched.describe(ch)
}

// reconMD5Task implements the collector.collectorTask interface.
func (t *reconMD5Task) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArg := "--md5"
	out, err := runCommandWithTimeout(4*time.Second, t.pathToReconExecutable, cmdArg)
	if err == nil {
		matchList := t.md5OutputRx.FindAllSubmatch(out, -1)
		if len(matchList) > 0 {
			for _, match := range matchList {
				var totalHosts, errsEncountered float64
				matchedHosts, err := strconv.ParseFloat(string(match[2]), 64)
				if err == nil {
					totalHosts, err = strconv.ParseFloat(string(match[3]), 64)
					if err == nil {
						errsEncountered, err = strconv.ParseFloat(string(match[4]), 64)
						if err == nil {
							kind := strings.ReplaceAll(string(match[1]), ".", "")
							notMatchedHosts := totalHosts - matchedHosts
							all := matchedHosts + notMatchedHosts + errsEncountered
							ch <- t.all.mustNewConstMetric(all, kind)
							ch <- t.errors.mustNewConstMetric(errsEncountered, kind)
							ch <- t.matched.mustNewConstMetric(matchedHosts, kind)
							ch <- t.notMatched.mustNewConstMetric(notMatchedHosts, kind)
						}
					}
				}
				if err != nil {
					exitCode = 1
					logg.Error("swift recon: %s: %s", cmdArg, err.Error())
				}
			}
		} else {
			err = errors.New("command did not return any usable output")
		}
	}
	if err != nil {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArg, err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArg)
}

// reconUpdaterSweepTask implements the collector.collectorTask interface.
type reconUpdaterSweepTask struct {
	pathToReconExecutable string

	containerTime typedDesc
	objectTime    typedDesc
}

func newReconUpdaterSweepTask(pathToReconExecutable string) collectorTask {
	return &reconUpdaterSweepTask{
		pathToReconExecutable: pathToReconExecutable,
		containerTime: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_containers_updater_sweep_time",
				"Container updater sweep time reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectTime: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_objects_updater_sweep_time",
				"Object updater sweep time reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconUpdaterSweepTask implements the collector.collectorTask interface.
func (t *reconUpdaterSweepTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.containerTime.describe(ch)
	t.objectTime.describe(ch)
}

// reconUpdaterSweepTask implements the collector.collectorTask interface.
func (t *reconUpdaterSweepTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	serverTypes := []string{"container", "object"}
	for _, server := range serverTypes {
		exitCode := 0
		cmdArgs := []string{server, "--updater", "--verbose"}
		outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
		if err == nil {
			for hostname, dataBytes := range outputPerHost {
				var data struct {
					ContainerUpdaterSweepTime float64 `json:"container_updater_sweep"`
					ObjectUpdaterSweepTime    float64 `json:"object_updater_sweep"`
				}
				err := json.Unmarshal(dataBytes, &data)
				if err != nil {
					exitCode = 1
					logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
					continue // to next host
				}

				val := data.ContainerUpdaterSweepTime
				desc := t.containerTime
				if server == "object" {
					val = data.ObjectUpdaterSweepTime
					desc = t.objectTime
				}

				ch <- desc.mustNewConstMetric(val, hostname)
			}
		} else {
			exitCode = 1
			logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		}

		ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
	}
}

// reconReplicationTask implements the collector.collectorTask interface.
type reconReplicationTask struct {
	pathToReconExecutable string

	accountReplicationAge        typedDesc
	accountReplicationDuration   typedDesc
	containerReplicationAge      typedDesc
	containerReplicationDuration typedDesc
	objectReplicationAge         typedDesc
	objectReplicationDuration    typedDesc
}

func newReconReplicationTask(pathToReconExecutable string) collectorTask {
	return &reconReplicationTask{
		pathToReconExecutable: pathToReconExecutable,
		accountReplicationAge: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_accounts_replication_age",
				"Account replication age reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		accountReplicationDuration: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_accounts_replication_duration",
				"Account replication duration reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerReplicationAge: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_containers_replication_age",
				"Container replication age reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerReplicationDuration: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_containers_replication_duration",
				"Container replication duration reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectReplicationAge: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_objects_replication_age",
				"Object replication age reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectReplicationDuration: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_objects_replication_duration",
				"Object replication duration reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconReplicationTask implements the collector.collectorTask interface.
func (t *reconReplicationTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.accountReplicationAge.describe(ch)
	t.accountReplicationDuration.describe(ch)
	t.containerReplicationAge.describe(ch)
	t.containerReplicationDuration.describe(ch)
	t.objectReplicationAge.describe(ch)
	t.objectReplicationDuration.describe(ch)
}

// reconReplicationTask implements the collector.collectorTask interface.
func (t *reconReplicationTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	serverTypes := []string{"account", "container", "object"}
	for _, server := range serverTypes {
		exitCode := 0
		cmdArgs := []string{server, "--replication", "--verbose"}

		var ageTypedDesc, durTypedDesc typedDesc
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

		outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
		if err == nil {
			for hostname, dataBytes := range outputPerHost {
				var data struct {
					ReplicationLast float64 `json:"replication_last"`
					ReplicationTime float64 `json:"replication_time"`
				}
				err := json.Unmarshal(dataBytes, &data)
				if err != nil {
					exitCode = 1
					logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
					continue // to next host
				}

				ch <- ageTypedDesc.mustNewConstMetric(data.ReplicationLast, hostname)
				ch <- durTypedDesc.mustNewConstMetric(data.ReplicationTime, hostname)
			}
		} else {
			exitCode = 1
			logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		}

		ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
	}
}

// reconQuarantinedTask implements the collector.collectorTask interface.
type reconQuarantinedTask struct {
	pathToReconExecutable string

	accounts   typedDesc
	containers typedDesc
	objects    typedDesc
}

func newReconQuarantinedTask(pathToReconExecutable string) collectorTask {
	return &reconQuarantinedTask{
		pathToReconExecutable: pathToReconExecutable,
		accounts: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_accounts_quarantined",
				"Quarantined accounts reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containers: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_containers_quarantined",
				"Quarantined containers reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objects: typedDesc{
			desc: prometheus.NewDesc("swift_cluster_objects_quarantined",
				"Quarantined objects reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconQuarantinedTask implements the collector.collectorTask interface.
func (t *reconQuarantinedTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.accounts.describe(ch)
	t.containers.describe(ch)
	t.objects.describe(ch)
}

// reconQuarantinedTask implements the collector.collectorTask interface.
func (t *reconQuarantinedTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArgs := []string{"--quarantined", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var data struct {
				Objects    int64 `json:"objects"`
				Accounts   int64 `json:"accounts"`
				Containers int64 `json:"containers"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				exitCode = 1
				logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
				continue // to next host
			}

			ch <- t.accounts.mustNewConstMetric(float64(data.Accounts), hostname)
			ch <- t.containers.mustNewConstMetric(float64(data.Containers), hostname)
			ch <- t.objects.mustNewConstMetric(float64(data.Objects), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}

// reconUnmountedTask implements the collector.collectorTask interface.
type reconUnmountedTask struct {
	pathToReconExecutable string
	unmountedDrives       typedDesc
}

func newReconUnmountedTask(pathToReconExecutable string) collectorTask {
	return &reconUnmountedTask{
		pathToReconExecutable: pathToReconExecutable,
		unmountedDrives: typedDesc{
			desc: prometheus.NewDesc(
				"swift_cluster_drives_unmounted",
				"Unmounted drives reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconUnmountedTask implements the collector.collectorTask interface.
func (t *reconUnmountedTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.unmountedDrives.describe(ch)
}

// reconUnmountedTask implements the collector.collectorTask interface.
func (t *reconUnmountedTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArgs := []string{"--unmounted", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var disksData []struct {
				Device string `json:"device"`
			}
			err := json.Unmarshal(dataBytes, &disksData)
			if err != nil {
				exitCode = 1
				logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
				continue // to next host
			}

			ch <- t.unmountedDrives.mustNewConstMetric(float64(len(disksData)), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}

// reconDriveAuditTask implements the collector.collectorTask interface.
type reconDriveAuditTask struct {
	pathToReconExecutable string
	auditErrors           typedDesc
}

func newReconDriveAuditTask(pathToReconExecutable string) collectorTask {
	return &reconDriveAuditTask{
		pathToReconExecutable: pathToReconExecutable,
		auditErrors: typedDesc{
			desc: prometheus.NewDesc(
				"swift_cluster_drives_audit_errors",
				"Drive audit errors reported by the swift-recon tool.",
				[]string{"storage_ip"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// reconDriveAuditTask implements the collector.collectorTask interface.
func (t *reconDriveAuditTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.auditErrors.describe(ch)
}

// reconDriveAuditTask implements the collector.collectorTask interface.
func (t *reconDriveAuditTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArgs := []string{"--driveaudit", "--verbose"}
	outputPerHost, err := getSwiftReconOutputPerHost(t.pathToReconExecutable, cmdArgs...)
	if err == nil {
		for hostname, dataBytes := range outputPerHost {
			var data struct {
				DriveAuditErrors int64 `json:"drive_audit_errors"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				exitCode = 1
				logg.Error("swift recon: %s: %s: %s", cmdArgsToStr(cmdArgs), hostname, err.Error())
				continue // to next host
			}

			ch <- t.auditErrors.mustNewConstMetric(float64(data.DriveAuditErrors), hostname)
		}
	} else {
		exitCode = 1
		logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions.

var reconHostOutputRx = regexp.MustCompile(`(?m)^-> https?://([a-zA-Z0-9-.]+)\S*\s(.*)$`)

func getSwiftReconOutputPerHost(pathToExecutable string, cmdArgs ...string) (map[string][]byte, error) {
	out, err := runCommandWithTimeout(4*time.Second, pathToExecutable, cmdArgs...)
	if err != nil {
		return nil, err
	}

	matchList := reconHostOutputRx.FindAllSubmatch(out, -1)
	if len(matchList) == 0 {
		return nil, errors.New("command did not return any usable output")
	}

	result := make(map[string][]byte)
	for _, match := range matchList {
		hostname := string(match[1])
		data := match[2]

		logg.Debug("output from command 'swift-recon %s': %s: %s", cmdArgsToStr(cmdArgs), hostname, string(data))

		// sanitize JSON
		data = bytes.ReplaceAll(data, []byte(`u'`), []byte(`'`))
		data = bytes.ReplaceAll(data, []byte(`'`), []byte(`"`))
		data = bytes.ReplaceAll(data, []byte(`True`), []byte(`true`))
		data = bytes.ReplaceAll(data, []byte(`False`), []byte(`false`))

		result[hostname] = data
	}

	return result, nil
}
