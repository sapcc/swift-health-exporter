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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// replicationTask implements the collector.collectorTask interface.
type replicationTask struct {
	isTest                bool
	pathToReconExecutable string
	hostTimeout           int
	ctxTimeout            time.Duration

	accountReplicationAge        *promhelper.TypedDesc
	accountReplicationDuration   *promhelper.TypedDesc
	containerReplicationAge      *promhelper.TypedDesc
	containerReplicationDuration *promhelper.TypedDesc
	objectReplicationAge         *promhelper.TypedDesc
	objectReplicationDuration    *promhelper.TypedDesc
}

func newReplicationTask(pathToReconExecutable string, isTest bool, hostTimeout int, ctxTimeout time.Duration) task {
	return &replicationTask{
		hostTimeout:           hostTimeout,
		ctxTimeout:            ctxTimeout,
		pathToReconExecutable: pathToReconExecutable,
		isTest:                isTest,
		accountReplicationAge: promhelper.NewGaugeTypedDesc(
			"swift_cluster_accounts_replication_age",
			"Account replication age reported by the swift-recon tool.", []string{"storage_ip"}),
		accountReplicationDuration: promhelper.NewGaugeTypedDesc(
			"swift_cluster_accounts_replication_duration",
			"Account replication duration reported by the swift-recon tool.", []string{"storage_ip"}),
		containerReplicationAge: promhelper.NewGaugeTypedDesc(
			"swift_cluster_containers_replication_age",
			"Container replication age reported by the swift-recon tool.", []string{"storage_ip"}),
		containerReplicationDuration: promhelper.NewGaugeTypedDesc(
			"swift_cluster_containers_replication_duration",
			"Container replication duration reported by the swift-recon tool.", []string{"storage_ip"}),
		objectReplicationAge: promhelper.NewGaugeTypedDesc(
			"swift_cluster_objects_replication_age",
			"Object replication age reported by the swift-recon tool.", []string{"storage_ip"}),
		objectReplicationDuration: promhelper.NewGaugeTypedDesc(
			"swift_cluster_objects_replication_duration",
			"Object replication duration reported by the swift-recon tool.", []string{"storage_ip"}),
	}
}

// describeMetrics implements the task interface.
func (t *replicationTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.accountReplicationAge.Describe(ch)
	t.accountReplicationDuration.Describe(ch)
	t.containerReplicationAge.Describe(ch)
	t.containerReplicationDuration.Describe(ch)
	t.objectReplicationAge.Describe(ch)
	t.objectReplicationDuration.Describe(ch)
}

// collectMetrics implements the task interface.
func (t *replicationTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc) {
	serverTypes := []string{"account", "container", "object"}
	for _, server := range serverTypes {
		exitCode := 0
		cmdArgs := []string{fmt.Sprintf("--timeout=%d", t.hostTimeout), server, "--replication", "--verbose"}

		var ageTypedDesc, durTypedDesc *promhelper.TypedDesc
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

		currentTime := float64(time.Now().Unix())
		outputPerHost, err := getSwiftReconOutputPerHost(t.ctxTimeout, t.pathToReconExecutable, cmdArgs...)
		if err == nil {
			for hostname, dataBytes := range outputPerHost {
				var data struct {
					ReplicationLast float64 `json:"replication_last"`
					ReplicationTime float64 `json:"replication_time"`
				}
				err := json.Unmarshal(dataBytes, &data)
				if err != nil {
					exitCode = 1
					outStr := fmt.Sprintf("output follows:\n%s", string(dataBytes))
					logg.Error("swift recon: %s: %s: %s: %s",
						cmdArgsToStr(cmdArgs), hostname, err.Error(), outStr)
					continue // to next host
				}

				if data.ReplicationLast > 0 {
					if t.isTest {
						currentTime = float64(timeNow().Second())
					}
					tDiff := currentTime - data.ReplicationLast
					ch <- ageTypedDesc.MustNewConstMetric(tDiff, hostname)
				}
				ch <- durTypedDesc.MustNewConstMetric(data.ReplicationTime, hostname)
			}
		} else {
			exitCode = 1
			logg.Error("swift recon: %s: %s", cmdArgsToStr(cmdArgs), err.Error())
		}

		ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), cmdArgsToStr(cmdArgs))
	}
}
