// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/assert"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/collector/dispersion"
	"github.com/sapcc/swift-health-exporter/internal/collector/recon"
)

func TestCollector(t *testing.T) {
	testCollector(t,
		"build/mock-swift-dispersion-report",
		"build/mock-swift-recon",
		"test/fixtures/successful_collect.prom")
}

func TestCollectorWithErrors(t *testing.T) {
	testCollector(t,
		"build/mock-swift-dispersion-report-with-errors",
		"build/mock-swift-recon-with-errors",
		"test/fixtures/failed_collect.prom")
}

func testCollector(t *testing.T, dispersionReportPath, reconPath, fixturesPath string) {
	recon.IsTest = true

	dispersionReportAbsPath, err := filepath.Abs(dispersionReportPath)
	if err != nil {
		t.Error(err)
	}
	reconAbsPath, err := filepath.Abs(reconPath)
	if err != nil {
		t.Error(err)
	}

	registry := prometheus.NewPedanticRegistry()
	c := collector.New()
	s := collector.NewScraper(0)

	dispersionExitCode := dispersion.GetTaskExitCodeGaugeVec(registry)
	addTask(true, c, s, dispersion.NewReportTask(dispersionReportAbsPath, 20*time.Second), dispersionExitCode)

	reconExitCode := recon.GetTaskExitCodeGaugeVec(registry)
	opts := &recon.TaskOpts{
		PathToExecutable: reconAbsPath,
		HostTimeout:      1,
		CtxTimeout:       4 * time.Second,
	}
	addTask(true, c, s, recon.NewDiskUsageTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewDriveAuditTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewMD5Task(opts), reconExitCode)
	addTask(true, c, s, recon.NewQuarantinedTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewReplicationTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewUnmountedTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewUpdaterSweepTask(opts), reconExitCode)
	addTask(true, c, s, recon.NewShardingTask(opts), reconExitCode)

	registry.MustRegister(c)

	s.UpdateAllMetrics(t.Context())
	assert.HTTPRequest{
		Method:       "GET",
		Path:         "/metrics",
		ExpectStatus: 200,
		ExpectBody:   assert.FixtureFile(fixturesPath),
	}.Check(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
