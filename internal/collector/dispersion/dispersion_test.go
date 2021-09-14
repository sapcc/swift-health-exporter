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

package dispersion

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/assert"
	"github.com/sapcc/swift-health-exporter/internal/collector"
)

func TestDispersionCollector(t *testing.T) {
	pathToExecutable, err := filepath.Abs("../../../build/mock-swift-dispersion-report")
	if err != nil {
		t.Error(err)
	}

	registry := prometheus.NewPedanticRegistry()
	c := collector.New(0)
	exitCode := GetTaskExitCodeTypedDesc(registry)
	c.AddTask(true, NewReportTask(pathToExecutable, 20*time.Second), exitCode)
	registry.MustRegister(c)

	assert.HTTPRequest{
		Method:       "GET",
		Path:         "/metrics",
		ExpectStatus: 200,
		ExpectBody:   assert.FixtureFile("fixtures/dispersion_successful_collect.prom"),
	}.Check(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}

func TestDispersionCollectorWithErrors(t *testing.T) {
	pathToExecutable, err := filepath.Abs("../../../build/mock-swift-dispersion-report-with-errors")
	if err != nil {
		t.Error(err)
	}

	registry := prometheus.NewPedanticRegistry()
	c := collector.New(0)
	exitCode := GetTaskExitCodeTypedDesc(registry)
	c.AddTask(true, NewReportTask(pathToExecutable, 20*time.Second), exitCode)
	registry.MustRegister(c)

	assert.HTTPRequest{
		Method:       "GET",
		Path:         "/metrics",
		ExpectStatus: 200,
		ExpectBody:   assert.FixtureFile("fixtures/dispersion_failed_collect.prom"),
	}.Check(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
