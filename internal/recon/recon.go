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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// task is the interface that a specific recon task must implement.
type task interface {
	describeMetrics(ch chan<- *prometheus.Desc)
	collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc *promhelper.TypedDesc)
}

// Collector implements the prometheus.Collector interface.
type Collector struct {
	tasks        []task
	taskExitCode *promhelper.TypedDesc
}

// CollectorOpts contains options that define the recon collector's behavior.
type CollectorOpts struct {
	IsTest               bool
	WithDiskUsage        bool
	WithDriveAudit       bool
	WithMD5              bool
	WithQuarantined      bool
	WithReplication      bool
	WithUnmounted        bool
	WithUpdaterSweepTime bool
	HostTimeout          int
	CtxTimeout           time.Duration
}

// NewCollector creates a new ReconCollector.
func NewCollector(pathToExecutable string, opts CollectorOpts) *Collector {
	var tasks []task
	if opts.WithDiskUsage {
		tasks = append(tasks, newDiskUsageTask(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithDriveAudit {
		tasks = append(tasks, newDriveAuditTask(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithMD5 {
		tasks = append(tasks, newMD5Task(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithQuarantined {
		tasks = append(tasks, newQuarantinedTask(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithReplication {
		tasks = append(tasks, newReplicationTask(pathToExecutable, opts.IsTest, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithUnmounted {
		tasks = append(tasks, newUnmountedTask(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}
	if opts.WithUpdaterSweepTime {
		tasks = append(tasks, newUpdaterSweepTask(pathToExecutable, opts.HostTimeout, opts.CtxTimeout))
	}

	return &Collector{
		tasks: tasks,
		taskExitCode: promhelper.NewGaugeTypedDesc(
			"swift_recon_task_exit_code",
			"The exit code for a Swift Recon query execution.", []string{"query"}),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.taskExitCode.Describe(ch)

	for _, t := range c.tasks {
		t.describeMetrics(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.tasks))
	for _, t := range c.tasks {
		go func(t task) {
			t.collectMetrics(ch, c.taskExitCode)
			wg.Done()
		}(t)
	}
	wg.Wait()
}
