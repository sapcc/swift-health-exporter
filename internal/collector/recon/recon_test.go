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
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/assert"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

func TestReconCollector(t *testing.T) {
	isTest = true

	pathToExecutable, err := filepath.Abs("../../../build/mock-swift-recon")
	if err != nil {
		t.Error(err)
	}

	registry := prometheus.NewPedanticRegistry()
	c := collector.New()
	s := collector.NewScraper(0)
	exitCode := GetTaskExitCodeGaugeVec(registry)
	opts := &TaskOpts{
		PathToExecutable: pathToExecutable,
		HostTimeout:      1,
		CtxTimeout:       4 * time.Second,
	}
	util.AddTask(true, c, s, NewDiskUsageTask(opts), exitCode)
	util.AddTask(true, c, s, NewDriveAuditTask(opts), exitCode)
	util.AddTask(true, c, s, NewMD5Task(opts), exitCode)
	util.AddTask(true, c, s, NewQuarantinedTask(opts), exitCode)
	util.AddTask(true, c, s, NewReplicationTask(opts), exitCode)
	util.AddTask(true, c, s, NewUnmountedTask(opts), exitCode)
	util.AddTask(true, c, s, NewUpdaterSweepTask(opts), exitCode)
	registry.MustRegister(c)

	s.UpdateAllMetrics()
	assert.HTTPRequest{
		Method:       "GET",
		Path:         "/metrics",
		ExpectStatus: 200,
		ExpectBody:   assert.FixtureFile("fixtures/recon_successful_collect.prom"),
	}.Check(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}

func TestReconCollectorWithErrors(t *testing.T) {
	isTest = true

	pathToExecutable, err := filepath.Abs("../../../build/mock-swift-recon-with-errors")
	if err != nil {
		t.Error(err)
	}

	registry := prometheus.NewPedanticRegistry()
	c := collector.New()
	s := collector.NewScraper(0)
	exitCode := GetTaskExitCodeGaugeVec(registry)
	opts := &TaskOpts{
		PathToExecutable: pathToExecutable,
		HostTimeout:      1,
		CtxTimeout:       4 * time.Second,
	}
	util.AddTask(true, c, s, NewDiskUsageTask(opts), exitCode)
	util.AddTask(true, c, s, NewDriveAuditTask(opts), exitCode)
	util.AddTask(true, c, s, NewMD5Task(opts), exitCode)
	util.AddTask(true, c, s, NewQuarantinedTask(opts), exitCode)
	util.AddTask(true, c, s, NewReplicationTask(opts), exitCode)
	util.AddTask(true, c, s, NewUnmountedTask(opts), exitCode)
	util.AddTask(true, c, s, NewUpdaterSweepTask(opts), exitCode)
	registry.MustRegister(c)

	s.UpdateAllMetrics()
	assert.HTTPRequest{
		Method:       "GET",
		Path:         "/metrics",
		ExpectStatus: 200,
		ExpectBody:   assert.FixtureFile("fixtures/recon_failed_collect.prom"),
	}.Check(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
