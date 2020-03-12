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
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

// DispersionCollector implements the prometheus.Collector interface.
type DispersionCollector struct {
	ctxTimeout               time.Duration
	taskExitCode             typedDesc
	dispersionReportDumpTask collectorTask
}

// NewDispersionCollector creates a new DispersionCollector.
func NewDispersionCollector(pathToExecutable string, ctxTimeout time.Duration) *DispersionCollector {
	return &DispersionCollector{
		taskExitCode: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_task_exit_code",
				"The exit code for a Swift Dispersion Report query execution.",
				[]string{"query"}, nil),
			valueType: prometheus.GaugeValue,
		},
		dispersionReportDumpTask: newDispersionReportDumpTask(pathToExecutable, ctxTimeout),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *DispersionCollector) Describe(ch chan<- *prometheus.Desc) {
	c.taskExitCode.describe(ch)
	c.dispersionReportDumpTask.describeMetrics(ch)
}

// Collect implements the prometheus.Collector interface.
func (c *DispersionCollector) Collect(ch chan<- prometheus.Metric) {
	c.dispersionReportDumpTask.collectMetrics(ch, c.taskExitCode)
}

///////////////////////////////////////////////////////////////////////////////
// Dispersion collector tasks.

// dispersionReportDumpTask implements the collector.collectorTask interface.
type dispersionReportDumpTask struct {
	ctxTimeout                 time.Duration
	pathToDispersionExecutable string

	containerCopiesExpected typedDesc
	containerCopiesFound    typedDesc
	containerCopiesMissing  typedDesc
	containerOverlapping    typedDesc
	objectCopiesExpected    typedDesc
	objectCopiesFound       typedDesc
	objectCopiesMissing     typedDesc
	objectOverlapping       typedDesc
}

func newDispersionReportDumpTask(pathToDispersionExecutable string, ctxTimeout time.Duration) collectorTask {
	return &dispersionReportDumpTask{
		ctxTimeout:                 ctxTimeout,
		pathToDispersionExecutable: pathToDispersionExecutable,
		containerCopiesExpected: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_expected",
				"Expected container copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		containerCopiesFound: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_found",
				"Found container copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		containerCopiesMissing: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_missing",
				"Missing container copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		containerOverlapping: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_overlapping",
				"Expected container copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesExpected: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_expected",
				"Expected object copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesFound: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_found",
				"Found object copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesMissing: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_missing",
				"Missing object copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
		objectOverlapping: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_overlapping",
				"Expected object copies reported by the swift-dispersion-report tool.",
				nil, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// dispersionReportDumpTask implements the collector.collectorTask interface.
func (t *dispersionReportDumpTask) describeMetrics(ch chan<- *prometheus.Desc) {
	t.containerCopiesExpected.describe(ch)
	t.containerCopiesFound.describe(ch)
	t.containerCopiesMissing.describe(ch)
	t.containerOverlapping.describe(ch)
	t.objectCopiesExpected.describe(ch)
	t.objectCopiesFound.describe(ch)
	t.objectCopiesMissing.describe(ch)
	t.objectOverlapping.describe(ch)
}

// dispersionReportDumpTask implements the collector.collectorTask interface.
func (t *dispersionReportDumpTask) collectMetrics(ch chan<- prometheus.Metric, exitCodeTypedDesc typedDesc) {
	exitCode := 0
	cmdArg := "--dump-json"
	// in large Swift clusters, the dispersion-report tool takes time. Hence the longer timeout.
	out, err := runCommandWithTimeout(t.ctxTimeout, t.pathToDispersionExecutable, cmdArg)
	if err == nil {
		var data struct {
			Object struct {
				Expected    int64 `json:"copies_expected"`
				Found       int64 `json:"copies_found"`
				Overlapping int64 `json:"overlapping"`
				Missing     int64
			} `json:"object"`
			Container struct {
				Expected    int64 `json:"copies_expected"`
				Found       int64 `json:"copies_found"`
				Overlapping int64 `json:"overlapping"`
				Missing     int64
			} `json:"container"`
		}
		err = json.Unmarshal(out, &data)
		if err == nil {
			cntr := data.Container
			if cntr.Expected > 0 && cntr.Found > 0 {
				cntr.Missing = cntr.Expected - cntr.Found
			}
			ch <- t.containerCopiesExpected.mustNewConstMetric(float64(cntr.Expected))
			ch <- t.containerCopiesFound.mustNewConstMetric(float64(cntr.Found))
			ch <- t.containerCopiesMissing.mustNewConstMetric(float64(cntr.Missing))
			ch <- t.containerOverlapping.mustNewConstMetric(float64(cntr.Overlapping))

			obj := data.Object
			if obj.Expected > 0 && obj.Found > 0 {
				obj.Missing = obj.Expected - obj.Found
			}
			ch <- t.objectCopiesExpected.mustNewConstMetric(float64(obj.Expected))
			ch <- t.objectCopiesFound.mustNewConstMetric(float64(obj.Found))
			ch <- t.objectCopiesMissing.mustNewConstMetric(float64(obj.Missing))
			ch <- t.objectOverlapping.mustNewConstMetric(float64(obj.Overlapping))
		}
	}
	if err != nil {
		exitCode = 1
		logg.Error("swift dispersion: %s: %s", cmdArg, err.Error())
	}

	ch <- exitCodeTypedDesc.mustNewConstMetric(float64(exitCode), cmdArg)
}
