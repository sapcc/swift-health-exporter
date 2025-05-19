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

// ShardingTask implements the collector.Task interface.
type ShardingTask struct {
	opts    *TaskOpts
	cmdArgs []string

	containerShardingAuditRootAttempted  *prometheus.GaugeVec
	containerShardingAuditRootFailure    *prometheus.GaugeVec
	containerShardingAuditRootSuccess    *prometheus.GaugeVec
	containerShardingAuditRootHasOverlap *prometheus.GaugeVec
	containerShardingAuditRootNumOverlap *prometheus.GaugeVec

	containerShardingAuditShardAttempted *prometheus.GaugeVec
	containerShardingAuditShardFailure   *prometheus.GaugeVec
	containerShardingAuditShardSuccess   *prometheus.GaugeVec

	containerShardingCleavedAttempted *prometheus.GaugeVec
	containerShardingCleavedFailure   *prometheus.GaugeVec
	containerShardingCleavedMaxTime   *prometheus.GaugeVec
	containerShardingCleavedMinTime   *prometheus.GaugeVec
	containerShardingCleavedSuccess   *prometheus.GaugeVec

	containerShardingCreatedAttempted *prometheus.GaugeVec
	containerShardingCreatedFailure   *prometheus.GaugeVec
	containerShardingCreatedSuccess   *prometheus.GaugeVec

	containerShardingScannedAttempted *prometheus.GaugeVec
	containerShardingScannedFailure   *prometheus.GaugeVec
	containerShardingScannedMaxTime   *prometheus.GaugeVec
	containerShardingScannedMinTime   *prometheus.GaugeVec
	containerShardingScannedSuccess   *prometheus.GaugeVec

	containerShardingMisplacedAttempted *prometheus.GaugeVec
	containerShardingMisplacedFailure   *prometheus.GaugeVec
	containerShardingMisplacedFound     *prometheus.GaugeVec
	containerShardingMisplacedPlaced    *prometheus.GaugeVec
	containerShardingMisplacedSuccess   *prometheus.GaugeVec
	containerShardingMisplacedUnplaced  *prometheus.GaugeVec

	containerShardingVisitedAttempted *prometheus.GaugeVec
	containerShardingVisitedCompleted *prometheus.GaugeVec
	containerShardingVisitedFailure   *prometheus.GaugeVec
	containerShardingVisitedSkipped   *prometheus.GaugeVec
	containerShardingVisitedSuccess   *prometheus.GaugeVec

	containerShardingInProgressActive      *prometheus.GaugeVec
	containerShardingInProgressCleaved     *prometheus.GaugeVec
	containerShardingInProgressCreated     *prometheus.GaugeVec
	containerShardingInProgressError       *prometheus.GaugeVec
	containerShardingInProgressFound       *prometheus.GaugeVec
	containerShardingInProgressObjectcount *prometheus.GaugeVec

	containerShardingCandidatesFound       *prometheus.GaugeVec
	containerShardingCandidatesObjectCount *prometheus.GaugeVec
}

// NewShardingTask returns a collector.Task for ShardingTask.
func NewShardingTask(opts *TaskOpts) collector.Task {
	return &ShardingTask{
		opts: opts,
		// <server-type> gets substituted in UpdateMetrics().
		cmdArgs: []string{
			fmt.Sprintf("--timeout=%d", opts.HostTimeout), "container",
			"--sharding", "--verbose",
		},
		containerShardingAuditRootAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_root_attempted",
				Help: "Container root DB auditor number attempted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditRootFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_root_failure",
				Help: "Container root DB auditor number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditRootSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_root_success",
				Help: "Container root DB auditor number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditRootHasOverlap: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_root_has_overlap",
				Help: "Container root DB auditor has_overlap reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditRootNumOverlap: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_root_num_overlap",
				Help: "Container root DB auditor number of overlaps reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditShardAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_shard_attempted",
				Help: "Container shard DB auditor number attmpted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditShardFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_shard_failure",
				Help: "Container shard DB auditor number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingAuditShardSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_audit_shard_success",
				Help: "Container shard DB auditor number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCleavedAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_cleaved_attempted",
				Help: "Container shard cleaved number attempted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCleavedFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_cleaved_failure",
				Help: "Container shard cleaved number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCleavedMaxTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_cleaved_max_time",
				Help: "Container shard cleaved max_time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCleavedMinTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_cleaved_min_time",
				Help: "Container shard cleaved min_time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCleavedSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_cleaved_success",
				Help: "Container shard cleaved number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCreatedAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_created_attempted",
				Help: "Container shard created number attempted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCreatedFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_created_failure",
				Help: "Container shard created number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCreatedSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_created_success",
				Help: "Container shard created number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingScannedAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_scanned_attempted",
				Help: "Container shard scanned number attempted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingScannedFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_scanned_failure",
				Help: "Container shard scanned number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingScannedMaxTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_scanned_max_time",
				Help: "Container shard scanned max_time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingScannedMinTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_scanned_min_time",
				Help: "Container shard scanned min_time reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingScannedSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_scanned_success",
				Help: "Container shard scanned number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_attempted",
				Help: "Container sharding stats on misplaced objects reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_failure",
				Help: "Container sharding stats on misplaced objects failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedFound: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_found",
				Help: "Container sharding stats on misplaced objects number found reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedPlaced: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_placed",
				Help: "Container sharding stats on misplaced objects number placed reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_success",
				Help: "Container sharding stats on misplaced objects number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingMisplacedUnplaced: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_misplaced_unplaced",
				Help: "Container sharding stats on misplaced objects reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingVisitedAttempted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_visited_attempted",
				Help: "Container shard visited number attempted reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingVisitedCompleted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_visited_completed",
				Help: "Container shard visited number completed reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingVisitedFailure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_visited_failure",
				Help: "Container shard visited number of failures reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingVisitedSkipped: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_visited_skipped",
				Help: "Container shard visited number skipped reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingVisitedSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_visited_success",
				Help: "Container shard visited number of successes reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingInProgressActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_active",
				Help: "Container sharding in progress number of shards active reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingInProgressCleaved: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_cleaved",
				Help: "Container sharding in progress number of shards cleaved reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingInProgressCreated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_created",
				Help: "Container sharding in progress number of shards created reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingInProgressError: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_error",
				Help: "Container sharding in progress number of errors reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingInProgressFound: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_found",
				Help: "Container sharding in progress number found reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingInProgressObjectcount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_in_progress_object_count",
				Help: "Container sharding in progress object count reported by the swift-recon tool.",
			}, []string{"storage_ip", "container", "account"}),
		containerShardingCandidatesFound: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_candidates_found",
				Help: "Number of container sharding candidates reported by the swift-recon tool.",
			}, []string{"storage_ip"}),
		containerShardingCandidatesObjectCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "swift_cluster_containers_sharding_candidates_object_count",
				Help: "Container sharding candidates object count reported by the swift-recon tool.",
			}, []string{"storage_ip", "account", "container"}),
	}
}

// Name implements the collector.Task interface.
func (t *ShardingTask) Name() string {
	return "recon-sharding"
}

// DescribeMetrics implements the collector.Task interface.
//
//nolint:dupl
func (t *ShardingTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.containerShardingAuditRootAttempted.Describe(ch)
	t.containerShardingAuditRootFailure.Describe(ch)
	t.containerShardingAuditRootHasOverlap.Describe(ch)
	t.containerShardingAuditRootNumOverlap.Describe(ch)
	t.containerShardingAuditRootSuccess.Describe(ch)

	t.containerShardingAuditShardAttempted.Describe(ch)
	t.containerShardingAuditShardFailure.Describe(ch)
	t.containerShardingAuditShardSuccess.Describe(ch)

	t.containerShardingCleavedAttempted.Describe(ch)
	t.containerShardingCleavedFailure.Describe(ch)
	t.containerShardingCleavedMaxTime.Describe(ch)
	t.containerShardingCleavedMinTime.Describe(ch)
	t.containerShardingCleavedSuccess.Describe(ch)

	t.containerShardingCreatedAttempted.Describe(ch)
	t.containerShardingCreatedFailure.Describe(ch)
	t.containerShardingCreatedSuccess.Describe(ch)

	t.containerShardingMisplacedAttempted.Describe(ch)
	t.containerShardingMisplacedFailure.Describe(ch)
	t.containerShardingMisplacedFound.Describe(ch)
	t.containerShardingMisplacedPlaced.Describe(ch)
	t.containerShardingMisplacedSuccess.Describe(ch)
	t.containerShardingMisplacedUnplaced.Describe(ch)

	t.containerShardingScannedAttempted.Describe(ch)
	t.containerShardingScannedFailure.Describe(ch)
	t.containerShardingScannedMaxTime.Describe(ch)
	t.containerShardingScannedMinTime.Describe(ch)
	t.containerShardingScannedSuccess.Describe(ch)

	t.containerShardingVisitedAttempted.Describe(ch)
	t.containerShardingVisitedCompleted.Describe(ch)
	t.containerShardingVisitedFailure.Describe(ch)
	t.containerShardingVisitedSkipped.Describe(ch)
	t.containerShardingVisitedSuccess.Describe(ch)

	t.containerShardingInProgressActive.Describe(ch)
	t.containerShardingInProgressCleaved.Describe(ch)
	t.containerShardingInProgressCreated.Describe(ch)
	t.containerShardingInProgressError.Describe(ch)
	t.containerShardingInProgressFound.Describe(ch)
	t.containerShardingInProgressObjectcount.Describe(ch)

	t.containerShardingCandidatesFound.Describe(ch)
	t.containerShardingCandidatesObjectCount.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
//
//nolint:dupl
func (t *ShardingTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.containerShardingAuditRootAttempted.Collect(ch)
	t.containerShardingAuditRootFailure.Collect(ch)
	t.containerShardingAuditRootHasOverlap.Collect(ch)
	t.containerShardingAuditRootNumOverlap.Collect(ch)
	t.containerShardingAuditRootSuccess.Collect(ch)

	t.containerShardingAuditShardAttempted.Collect(ch)
	t.containerShardingAuditShardFailure.Collect(ch)
	t.containerShardingAuditShardSuccess.Collect(ch)

	t.containerShardingCleavedAttempted.Collect(ch)
	t.containerShardingCleavedFailure.Collect(ch)
	t.containerShardingCleavedMaxTime.Collect(ch)
	t.containerShardingCleavedMinTime.Collect(ch)
	t.containerShardingCleavedSuccess.Collect(ch)

	t.containerShardingCreatedAttempted.Collect(ch)
	t.containerShardingCreatedFailure.Collect(ch)
	t.containerShardingCreatedSuccess.Collect(ch)

	t.containerShardingMisplacedAttempted.Collect(ch)
	t.containerShardingMisplacedFailure.Collect(ch)
	t.containerShardingMisplacedFound.Collect(ch)
	t.containerShardingMisplacedPlaced.Collect(ch)
	t.containerShardingMisplacedSuccess.Collect(ch)
	t.containerShardingMisplacedUnplaced.Collect(ch)

	t.containerShardingScannedAttempted.Collect(ch)
	t.containerShardingScannedFailure.Collect(ch)
	t.containerShardingScannedMaxTime.Collect(ch)
	t.containerShardingScannedMinTime.Collect(ch)
	t.containerShardingScannedSuccess.Collect(ch)

	t.containerShardingVisitedAttempted.Collect(ch)
	t.containerShardingVisitedCompleted.Collect(ch)
	t.containerShardingVisitedFailure.Collect(ch)
	t.containerShardingVisitedSkipped.Collect(ch)
	t.containerShardingVisitedSuccess.Collect(ch)

	t.containerShardingInProgressActive.Collect(ch)
	t.containerShardingInProgressCleaved.Collect(ch)
	t.containerShardingInProgressCreated.Collect(ch)
	t.containerShardingInProgressError.Collect(ch)
	t.containerShardingInProgressFound.Collect(ch)
	t.containerShardingInProgressObjectcount.Collect(ch)

	t.containerShardingCandidatesFound.Collect(ch)
	t.containerShardingCandidatesObjectCount.Collect(ch)
}

type ShardingAuditRoot struct {
	Attempted  int `json:"attempted"`
	Failure    int `json:"failure"`
	HasOverlap int `json:"has_overlap"`
	NumOverlap int `json:"num_overlap"`
	Success    int `json:"success"`
}

type ShardingAuditShard struct {
	Attempted int `json:"attempted"`
	Failure   int `json:"failure"`
	Success   int `json:"success"`
}

type ShardingShardCleaved struct {
	Attempted int `json:"attempted"`
	Failure   int `json:"failure"`
	MaxTime   int `json:"max_time"`
	MinTime   int `json:"min_time"`
	Success   int `json:"success"`
}

type ShardingShardCreated struct {
	Attempted int `json:"attempted"`
	Failure   int `json:"failure"`
	Success   int `json:"success"`
}

type ShardingMisplaced struct {
	Attempted int `json:"attempted"`
	Failure   int `json:"failure"`
	Found     int `json:"found"`
	Placed    int `json:"placed"`
	Success   int `json:"success"`
	Unplaced  int `json:"unplaced"`
}

type ShardingScanned struct {
	Attempted int `json:"attempted"`
	Failure   int `json:"failure"`
	MaxTime   int `json:"max_time"`
	MinTime   int `json:"min_time"`
	Success   int `json:"success"`
}

type ShardingVisited struct {
	Attempted int `json:"attempted"`
	Completed int `json:"completed"`
	Failure   int `json:"failure"`
	Skipped   int `json:"skipped"`
	Success   int `json:"success"`
}

type ShardingInProgress struct {
	All []struct {
		Account     string `json:"account"`
		Active      int    `json:"active"`
		Cleaved     int    `json:"cleaved"`
		Container   string `json:"container"`
		Created     int    `json:"created"`
		DBState     string `json:"db_state"`
		Error       string `json:"error"`
		FileSize    int64  `json:"file_size"`
		Found       int    `json:"found"`
		NodeIndex   int    `json:"node_index"`
		ObjectCount int64  `json:"object_count"`
		Path        string `json:"path"`
		State       string `json:"state"`
		Root        string `json:"root"`
	} `json:"all"`
}

type ShardingInProgressTop struct {
	Account     string `json:"account"`
	Container   string `json:"container"`
	FileSize    int64  `json:"file_size"`
	NodeIndex   int    `json:"node_index"`
	ObjectCount int64  `json:"object_count"`
	Path        string `json:"path"`
	Root        string `json:"root"`
}

type ShardingCandidates struct {
	Found int                     `json:"found"`
	Top   []ShardingInProgressTop `json:"top"`
}

type Sharding struct {
	AuditRoot          ShardingAuditRoot    `json:"audit_root"`
	AuditShard         ShardingAuditShard   `json:"audit_shard"`
	Cleaved            ShardingShardCleaved `json:"cleaved"`
	Created            ShardingShardCreated `json:"created"`
	Misplaced          ShardingMisplaced    `json:"misplaced"`
	Scanned            ShardingScanned      `json:"scanned"`
	Visited            ShardingVisited      `json:"visited"`
	ShardingInProgress ShardingInProgress   `json:"sharding_in_progress"`
	ShardingCandidates ShardingCandidates   `json:"sharding_candidates"`
}

type ShardingStats struct {
	Sharding Sharding `json:"sharding"`
}

// UpdateMetrics implements the collector.Task interface.
func (t *ShardingTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	cmdArgs := t.cmdArgs
	q := util.CmdArgsToStr(cmdArgs)
	queries := map[string]int{q: 0}
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
			ShardingStats ShardingStats `json:"sharding_stats"`
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

		t.containerShardingAuditRootAttempted.With(l).Set(float64(data.ShardingStats.Sharding.AuditRoot.Attempted))
		t.containerShardingAuditRootFailure.With(l).Set(float64(data.ShardingStats.Sharding.AuditRoot.Failure))
		t.containerShardingAuditRootHasOverlap.With(l).Set(float64(data.ShardingStats.Sharding.AuditRoot.NumOverlap))
		t.containerShardingAuditRootNumOverlap.With(l).Set(float64(data.ShardingStats.Sharding.AuditRoot.NumOverlap))
		t.containerShardingAuditRootSuccess.With(l).Set(float64(data.ShardingStats.Sharding.AuditRoot.Success))

		t.containerShardingAuditShardAttempted.With(l).Set(float64(data.ShardingStats.Sharding.AuditShard.Attempted))
		t.containerShardingAuditShardFailure.With(l).Set(float64(data.ShardingStats.Sharding.AuditShard.Failure))
		t.containerShardingAuditShardSuccess.With(l).Set(float64(data.ShardingStats.Sharding.AuditShard.Success))

		t.containerShardingCleavedAttempted.With(l).Set(float64(data.ShardingStats.Sharding.Cleaved.Attempted))
		t.containerShardingCleavedFailure.With(l).Set(float64(data.ShardingStats.Sharding.Cleaved.Failure))
		t.containerShardingCleavedMaxTime.With(l).Set(float64(data.ShardingStats.Sharding.Cleaved.MaxTime))
		t.containerShardingCleavedMinTime.With(l).Set(float64(data.ShardingStats.Sharding.Cleaved.MinTime))
		t.containerShardingCleavedSuccess.With(l).Set(float64(data.ShardingStats.Sharding.Cleaved.Success))

		t.containerShardingCreatedAttempted.With(l).Set(float64(data.ShardingStats.Sharding.Created.Attempted))
		t.containerShardingCreatedFailure.With(l).Set(float64(data.ShardingStats.Sharding.Created.Failure))
		t.containerShardingCreatedSuccess.With(l).Set(float64(data.ShardingStats.Sharding.Created.Success))

		t.containerShardingMisplacedAttempted.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Attempted))
		t.containerShardingMisplacedFailure.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Failure))
		t.containerShardingMisplacedFound.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Found))
		t.containerShardingMisplacedPlaced.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Placed))
		t.containerShardingMisplacedSuccess.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Success))
		t.containerShardingMisplacedUnplaced.With(l).Set(float64(data.ShardingStats.Sharding.Misplaced.Unplaced))

		t.containerShardingScannedAttempted.With(l).Set(float64(data.ShardingStats.Sharding.Scanned.Attempted))
		t.containerShardingScannedFailure.With(l).Set(float64(data.ShardingStats.Sharding.Scanned.Failure))
		t.containerShardingScannedMaxTime.With(l).Set(float64(data.ShardingStats.Sharding.Scanned.MaxTime))
		t.containerShardingScannedMinTime.With(l).Set(float64(data.ShardingStats.Sharding.Scanned.MinTime))
		t.containerShardingScannedSuccess.With(l).Set(float64(data.ShardingStats.Sharding.Scanned.Success))

		t.containerShardingVisitedAttempted.With(l).Set(float64(data.ShardingStats.Sharding.Visited.Attempted))
		t.containerShardingVisitedCompleted.With(l).Set(float64(data.ShardingStats.Sharding.Visited.Completed))
		t.containerShardingVisitedFailure.With(l).Set(float64(data.ShardingStats.Sharding.Visited.Failure))
		t.containerShardingVisitedSkipped.With(l).Set(float64(data.ShardingStats.Sharding.Visited.Skipped))
		t.containerShardingVisitedSuccess.With(l).Set(float64(data.ShardingStats.Sharding.Visited.Success))

		for _, shardingProcess := range data.ShardingStats.Sharding.ShardingInProgress.All {
			l := prometheus.Labels{"storage_ip": hostname, "container": shardingProcess.Container, "account": shardingProcess.Account}
			t.containerShardingInProgressActive.With(l).Set(float64(shardingProcess.Active))
			t.containerShardingInProgressCleaved.With(l).Set(float64(shardingProcess.Cleaved))
			t.containerShardingInProgressCreated.With(l).Set(float64(shardingProcess.Created))
			if shardingProcess.Error != "None" {
				t.containerShardingInProgressError.With(l).Set(float64(1))
			} else {
				t.containerShardingInProgressError.With(l).Set(float64(0))
			}
			t.containerShardingInProgressFound.With(l).Set(float64(shardingProcess.Found))
			t.containerShardingInProgressObjectcount.With(l).Set(float64(shardingProcess.ObjectCount))
		}

		t.containerShardingCandidatesFound.With(prometheus.Labels{"storage_ip": hostname}).Set(float64(data.ShardingStats.Sharding.ShardingCandidates.Found))

		for _, shardingCandidate := range data.ShardingStats.Sharding.ShardingCandidates.Top {
			t.containerShardingCandidatesObjectCount.With(prometheus.Labels{"storage_ip": hostname, "container": shardingCandidate.Container, "account": shardingCandidate.Account}).Set(float64(shardingCandidate.ObjectCount))
		}
	}

	return queries, nil
}
