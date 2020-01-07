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

package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

var (
	reconTaskErrorDesc = prometheus.NewDesc(
		"swift_recon_exit_code",
		"The exit code for a Swift Recon query execution.",
		[]string{"query"}, nil,
	)

	sendReconErrorCount = func(ch chan<- prometheus.Metric, cmdArgs []string, count int) {
		query := strings.Join(cmdArgs, " ")
		ch <- prometheus.MustNewConstMetric(
			reconTaskErrorDesc,
			prometheus.CounterValue, float64(count),
			query,
		)
	}
)

// ReconCollector implements the prometheus.Collector interface.
type ReconCollector struct {
	pathToExecutable string
	tasks            map[string]func(string, string, chan<- prometheus.Metric)
}

// NewReconCollector creates a new ReconCollector.
func NewReconCollector(pathToExecutable string) *ReconCollector {
	return &ReconCollector{
		pathToExecutable: pathToExecutable,
		tasks: map[string]func(string, string, chan<- prometheus.Metric){
			"diskUsage":    reconDiskUsageTask,
			"driveAudit":   reconDriveAuditTask,
			"md5":          reconMD5Task,
			"quarantined":  reconQuarantinedTask,
			"replication":  reconReplicationTask,
			"unmounted":    reconUnmountedTask,
			"updaterSweep": reconUpdaterSweepTask,
		},
	}
}

// Describe implements the prometheus.Collector interface.
func (c *ReconCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface.
func (c *ReconCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.tasks))
	for name, task := range c.tasks {
		go func(name string, task func(string, string, chan<- prometheus.Metric)) {
			task(name, c.pathToExecutable, ch)
			wg.Done()
		}(name, task)
	}
	wg.Wait()
}

var (
	clusterStorageUsedPercentByDiskDesc = prometheus.NewDesc(
		"swift_cluster_storage_used_percent_by_disk",
		"Percentage of storage used by a disk as reported by the swift-recon tool.",
		[]string{"storage_ip", "disk"}, nil,
	)
	clusterStorageUsedPercentDesc = prometheus.NewDesc(
		"swift_cluster_storage_used_percent",
		"Percentage of storage used as reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterStorageUsedBytesDesc = prometheus.NewDesc(
		"swift_cluster_storage_used_bytes",
		"Used storage bytes as reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterStorageFreeBytesDesc = prometheus.NewDesc(
		"swift_cluster_storage_free_bytes",
		"Free storage bytes as reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterStorageCapacityBytesDesc = prometheus.NewDesc(
		"swift_cluster_storage_capacity_bytes",
		"Capacity storage bytes as reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
)

var specialCharRx = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func reconDiskUsageTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	errCount := 0
	cmdArgs := []string{"--diskusage", "--verbose"}
	sendReconErrorCount(ch, cmdArgs, errCount)

	result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
	if err != nil {
		logg.Error("recon collector: %s: %v", taskName, err)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	gotErr := false
	for hostname, dataBytes := range result {
		var disksData []struct {
			Device  string `json:"device"`
			Avail   int64  `json:"avail"`
			Mounted bool   `json:"mounted"`
			Used    int64  `json:"used"`
			Size    int64  `json:"size"`
		}
		err = json.Unmarshal(dataBytes, &disksData)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, hostname, err)
			gotErr = true
			continue
		}

		var totalFree, totalUsed, totalSize int64
		for _, disk := range disksData {
			if !(disk.Mounted) {
				continue
			}

			totalFree += disk.Avail
			totalUsed += disk.Used
			totalSize += disk.Size

			// submit metrics by disk (only used percent here, which is the
			// most useful for alerting)
			device := specialCharRx.ReplaceAllLiteralString(disk.Device, "")
			ch <- prometheus.MustNewConstMetric(
				clusterStorageUsedPercentByDiskDesc,
				prometheus.GaugeValue, float64(disk.Used)/float64(disk.Size),
				hostname, device,
			)
		}

		usedPercent := float64(totalUsed) / float64(totalSize)
		if totalSize == 0 {
			usedPercent = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			clusterStorageUsedPercentDesc,
			prometheus.GaugeValue, usedPercent,
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterStorageUsedBytesDesc,
			prometheus.GaugeValue, float64(totalUsed),
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterStorageFreeBytesDesc,
			prometheus.GaugeValue, float64(totalFree),
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterStorageCapacityBytesDesc,
			prometheus.GaugeValue, float64(totalSize),
			hostname,
		)
	}
	if gotErr {
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
	}
}

var (
	clusterMD5AllDesc = prometheus.NewDesc(
		"swift_cluster_md5_all",
		"Sum of matched-, not matched hosts, and errors encountered while check hosts for md5sum(s) as reported by the swift-recon tool.",
		[]string{"kind"}, nil,
	)
	clusterMD5MatchedDesc = prometheus.NewDesc(
		"swift_cluster_md5_matched",
		"Matched hosts for md5sum(s) reported by the swift-recon tool.",
		[]string{"kind"}, nil,
	)
	clusterMD5NotMatchedDesc = prometheus.NewDesc(
		"swift_cluster_md5_not_matched",
		"Not matched hosts for md5sum(s) reported by the swift-recon tool.",
		[]string{"kind"}, nil,
	)
	clusterMD5ErrorsDesc = prometheus.NewDesc(
		"swift_cluster_md5_errors",
		"Errors encountered while checking hosts for md5sum(s) as reported by the swift-recon tool.",
		[]string{"kind"}, nil,
	)
)

var reconMD5Rx = regexp.MustCompile(
	`(?m)^.* Checking ([\.a-zA-Z0-9_]+) md5sum(?:s)?\s*([0-9]+)/([0-9]+) hosts matched, ([0-9]+) error.*$`)

func reconMD5Task(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	errCount := 0
	cmdArgs := []string{"--md5"}
	sendReconErrorCount(ch, cmdArgs, errCount)

	cmd := exec.Command(pathToReconExecutable, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		logg.Error("recon collector: %s: %v", taskName, err)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	matchList := reconMD5Rx.FindAllSubmatch(out, -1)
	if len(matchList) == 0 {
		logg.Error("recon collector: %s: command %q did not return any usable output", taskName, cmd)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	gotErr := false
	for _, match := range matchList {
		kind := strings.ReplaceAll(string(match[1]), ".", "")
		matchedHosts, err := strconv.ParseFloat(string(match[2]), 64)
		if err != nil {
			logg.Error("recon collector: %s: %v", taskName, err)
			gotErr = true
			continue
		}
		totalHosts, err := strconv.ParseFloat(string(match[3]), 64)
		if err != nil {
			logg.Error("recon collector: %s: %v", taskName, err)
			gotErr = true
			continue
		}
		notMatchedHosts := totalHosts - matchedHosts
		errsEncountered, err := strconv.ParseFloat(string(match[4]), 64)
		if err != nil {
			logg.Error("recon collector: %s: %v", taskName, err)
			gotErr = true
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			clusterMD5AllDesc,
			prometheus.GaugeValue, matchedHosts+notMatchedHosts+errsEncountered,
			kind,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterMD5MatchedDesc,
			prometheus.GaugeValue, matchedHosts,
			kind,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterMD5NotMatchedDesc,
			prometheus.GaugeValue, notMatchedHosts,
			kind,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterMD5ErrorsDesc,
			prometheus.GaugeValue, errsEncountered,
			kind,
		)
	}
	if gotErr {
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
	}
}

var (
	clusterCntrUpdaterSweepTimeDesc = prometheus.NewDesc(
		"swift_cluster_containers_updater_sweep_time",
		"Container updater sweep time reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterObjUpdaterSweepTimeDesc = prometheus.NewDesc(
		"swift_cluster_objects_updater_sweep_time",
		"Object updater sweep time reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
)

func reconUpdaterSweepTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	serverTypes := []string{"container", "object"}
	for _, server := range serverTypes {
		errCount := 0
		cmdArgs := []string{server, "--updater", "--verbose"}
		sendReconErrorCount(ch, cmdArgs, errCount)

		result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, server, err)
			errCount++
			sendReconErrorCount(ch, cmdArgs, errCount)
			return
		}

		gotErr := false
		for hostname, dataBytes := range result {
			var data struct {
				ContainerUpdaterSweepTime float64 `json:"container_updater_sweep"`
				ObjectUpdaterSweepTime    float64 `json:"object_updater_sweep"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				logg.Error("recon collector: %s: %s: %s: %v", taskName, server, hostname, err)
				gotErr = true
				continue
			}

			val := data.ContainerUpdaterSweepTime
			desc := clusterCntrUpdaterSweepTimeDesc
			if server == "object" {
				val = data.ObjectUpdaterSweepTime
				desc = clusterObjUpdaterSweepTimeDesc
			}

			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue, val,
				hostname,
			)
		}
		if gotErr {
			errCount++
			sendReconErrorCount(ch, cmdArgs, errCount)
		}
	}
}

var (
	clusterAccReplAgeDesc = prometheus.NewDesc(
		"swift_cluster_accounts_replication_age",
		"Account replication age reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterAccReplDurDesc = prometheus.NewDesc(
		"swift_cluster_accounts_replication_duration",
		"Account replication duration reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)

	clusterCntrReplAgeDesc = prometheus.NewDesc(
		"swift_cluster_containers_replication_age",
		"Container replication age reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterCntrReplDurDesc = prometheus.NewDesc(
		"swift_cluster_containers_replication_duration",
		"Container replication duration reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)

	clusterObjReplAgeDesc = prometheus.NewDesc(
		"swift_cluster_objects_replication_age",
		"Object replication age reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterObjReplDurDesc = prometheus.NewDesc(
		"swift_cluster_objects_replication_duration",
		"Object replication duration reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
)

func reconReplicationTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	serverTypes := []string{"account", "container", "object"}
	for _, server := range serverTypes {
		var ageDesc, durDesc *prometheus.Desc
		switch server {
		case "account":
			ageDesc = clusterAccReplAgeDesc
			durDesc = clusterAccReplDurDesc
		case "container":
			ageDesc = clusterCntrReplAgeDesc
			durDesc = clusterCntrReplDurDesc
		case "object":
			ageDesc = clusterObjReplAgeDesc
			durDesc = clusterObjReplDurDesc
		}

		errCount := 0
		cmdArgs := []string{server, "--replication", "--verbose"}
		sendReconErrorCount(ch, cmdArgs, errCount)

		result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, server, err)
			errCount++
			sendReconErrorCount(ch, cmdArgs, errCount)
			return
		}

		gotErr := false
		for hostname, dataBytes := range result {
			var data struct {
				ReplicationLast float64 `json:"replication_last"`
				ReplicationTime float64 `json:"replication_time"`
			}
			err := json.Unmarshal(dataBytes, &data)
			if err != nil {
				logg.Error("recon collector: %s: %s: %s: %v", taskName, server, hostname, err)
				gotErr = true
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				ageDesc,
				prometheus.GaugeValue, data.ReplicationLast,
				hostname,
			)
			ch <- prometheus.MustNewConstMetric(
				durDesc,
				prometheus.GaugeValue, data.ReplicationTime,
				hostname,
			)
		}
		if gotErr {
			errCount++
			sendReconErrorCount(ch, cmdArgs, errCount)
		}
	}
}

var (
	clusterAccQuarantinedDesc = prometheus.NewDesc(
		"swift_cluster_accounts_quarantined",
		"Quarantined accounts reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterCntrQuarantinedDesc = prometheus.NewDesc(
		"swift_cluster_containers_quarantined",
		"Quarantined containers reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
	clusterObjQuarantinedDesc = prometheus.NewDesc(
		"swift_cluster_objects_quarantined",
		"Quarantined objects reported by the swift-recon tool.",
		[]string{"storage_ip"}, nil,
	)
)

func reconQuarantinedTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	errCount := 0
	cmdArgs := []string{"--quarantined", "--verbose"}
	sendReconErrorCount(ch, cmdArgs, errCount)

	result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
	if err != nil {
		logg.Error("recon collector: %s: %v", taskName, err)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	gotErr := false
	for hostname, dataBytes := range result {
		var data struct {
			Objects    int64 `json:"objects"`
			Accounts   int64 `json:"accounts"`
			Containers int64 `json:"containers"`
		}
		err := json.Unmarshal(dataBytes, &data)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, hostname, err)
			gotErr = true
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			clusterAccQuarantinedDesc,
			prometheus.GaugeValue, float64(data.Accounts),
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterCntrQuarantinedDesc,
			prometheus.GaugeValue, float64(data.Containers),
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			clusterObjQuarantinedDesc,
			prometheus.GaugeValue, float64(data.Objects),
			hostname,
		)
	}
	if gotErr {
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
	}
}

var clusterDrivesUnmountedDesc = prometheus.NewDesc(
	"swift_cluster_drives_unmounted",
	"Unmounted drives reported by the swift-recon tool.",
	[]string{"storage_ip"}, nil,
)

func reconUnmountedTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	errCount := 0
	cmdArgs := []string{"--unmounted", "--verbose"}
	sendReconErrorCount(ch, cmdArgs, errCount)

	result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
	if err != nil {
		logg.Error("recon collector: %s: %v", taskName, err)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	gotErr := false
	for hostname, dataBytes := range result {
		var disksData []struct {
			Device string `json:"device"`
		}
		err := json.Unmarshal(dataBytes, &disksData)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, hostname, err)
			gotErr = true
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			clusterDrivesUnmountedDesc,
			prometheus.GaugeValue, float64(len(disksData)),
			hostname,
		)
	}
	if gotErr {
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
	}
}

var clusterDrivesAuditErrsDesc = prometheus.NewDesc(
	"swift_cluster_drives_audit_errors",
	"Drive audit errors reported by the swift-recon tool.",
	[]string{"storage_ip"}, nil,
)

func reconDriveAuditTask(taskName, pathToReconExecutable string, ch chan<- prometheus.Metric) {
	errCount := 0
	cmdArgs := []string{"--driveaudit", "--verbose"}
	sendReconErrorCount(ch, cmdArgs, errCount)

	result, err := getSwiftReconOutputPerHost(pathToReconExecutable, cmdArgs...)
	if err != nil {
		logg.Error("recon collector: %s: %v", taskName, err)
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
		return
	}

	gotErr := false
	for hostname, dataBytes := range result {
		var data struct {
			DriveAuditErrors int64 `json:"drive_audit_errors"`
		}
		err := json.Unmarshal(dataBytes, &data)
		if err != nil {
			logg.Error("recon collector: %s: %s: %v", taskName, hostname, err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			clusterDrivesAuditErrsDesc,
			prometheus.GaugeValue, float64(data.DriveAuditErrors),
			hostname,
		)
	}
	if gotErr {
		errCount++
		sendReconErrorCount(ch, cmdArgs, errCount)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions.

var reconHostOutputRx = regexp.MustCompile(`(?m)^-> https?://([a-zA-Z0-9-.]+)\S*\s(.*)$`)

func getSwiftReconOutputPerHost(pathToExecutable string, cmdArgs ...string) (map[string][]byte, error) {
	cmd := exec.Command(pathToExecutable, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	matchList := reconHostOutputRx.FindAllSubmatch(out, -1)
	if len(matchList) == 0 {
		return nil, fmt.Errorf("command %q did not return any usable output", cmd)
	}

	result := make(map[string][]byte)
	for _, match := range matchList {
		hostname := string(match[1])
		data := match[2]

		logg.Debug("output from command %q: %s: %s", cmd, hostname, string(data))

		// sanitize JSON
		data = bytes.ReplaceAll(data, []byte(`u'`), []byte(`'`))
		data = bytes.ReplaceAll(data, []byte(`'`), []byte(`"`))
		data = bytes.ReplaceAll(data, []byte(`True`), []byte(`true`))
		data = bytes.ReplaceAll(data, []byte(`False`), []byte(`false`))

		result[hostname] = data
	}

	return result, nil
}
