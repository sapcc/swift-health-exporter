// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package dispersion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/util"
)

// GetTaskExitCodeGaugeVec returns a *prometheus.GaugeVec for use with dispersion report
// tasks.
func GetTaskExitCodeGaugeVec(r prometheus.Registerer) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "swift_dispersion_task_exit_code",
			Help: "The exit code for a Swift dispersion report query execution.",
		}, []string{"query"},
	)
	r.MustRegister(gaugeVec)
	return gaugeVec
}

// ReportTask implements the collector.Task interface.
type ReportTask struct {
	ctxTimeout       time.Duration
	pathToExecutable string
	cmdArgs          []string

	// errRe is used to match errors and capture the hostname and error message.
	// E.g.:
	//   ERROR: 10.0.0.1:6000/swift-09: [Errno 111] ECONNREFUSED
	errRe *regexp.Regexp

	errors                  prometheus.Gauge
	containerCopiesExpected prometheus.Gauge
	containerCopiesFound    prometheus.Gauge
	containerCopiesMissing  prometheus.Gauge
	containerOverlapping    prometheus.Gauge
	objectCopiesExpected    prometheus.Gauge
	objectCopiesFound       prometheus.Gauge
	objectCopiesMissing     prometheus.Gauge
	objectOverlapping       prometheus.Gauge
}

// NewReportTask returns a collector.Task for ReportTask.
func NewReportTask(pathToExecutable string, ctxTimeout time.Duration) collector.Task {
	return &ReportTask{
		ctxTimeout:       ctxTimeout,
		pathToExecutable: pathToExecutable,
		cmdArgs:          []string{"--dump-json"},
		errRe:            regexp.MustCompile(`(?m)^ERROR:\s*([\d.]+)\S*\s*(.*)$`),
		errors: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_errors",
				Help: "The number of errors in the Swift dispersion report.",
			}),
		containerCopiesExpected: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_container_copies_expected",
				Help: "Expected container copies reported by the swift-dispersion-report tool.",
			}),
		containerCopiesFound: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_container_copies_found",
				Help: "Found container copies reported by the swift-dispersion-report tool.",
			}),
		containerCopiesMissing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_container_copies_missing",
				Help: "Missing container copies reported by the swift-dispersion-report tool.",
			}),
		containerOverlapping: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_container_overlapping",
				Help: "Expected container copies reported by the swift-dispersion-report tool.",
			}),
		objectCopiesExpected: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_object_copies_expected",
				Help: "Expected object copies reported by the swift-dispersion-report tool.",
			}),
		objectCopiesFound: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_object_copies_found",
				Help: "Found object copies reported by the swift-dispersion-report tool.",
			}),
		objectCopiesMissing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_object_copies_missing",
				Help: "Missing object copies reported by the swift-dispersion-report tool.",
			}),
		objectOverlapping: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "swift_dispersion_object_overlapping",
				Help: "Expected object copies reported by the swift-dispersion-report tool.",
			}),
	}
}

// Name implements the collector.Task interface.
func (t *ReportTask) Name() string {
	return "disperion-report"
}

// DescribeMetrics implements the collector.Task interface.
func (t *ReportTask) DescribeMetrics(ch chan<- *prometheus.Desc) {
	t.errors.Describe(ch)
	t.containerCopiesExpected.Describe(ch)
	t.containerCopiesFound.Describe(ch)
	t.containerCopiesMissing.Describe(ch)
	t.containerOverlapping.Describe(ch)
	t.objectCopiesExpected.Describe(ch)
	t.objectCopiesFound.Describe(ch)
	t.objectCopiesMissing.Describe(ch)
	t.objectOverlapping.Describe(ch)
}

// CollectMetrics implements the collector.Task interface.
func (t *ReportTask) CollectMetrics(ch chan<- prometheus.Metric) {
	t.errors.Collect(ch)
	t.containerCopiesExpected.Collect(ch)
	t.containerCopiesFound.Collect(ch)
	t.containerCopiesMissing.Collect(ch)
	t.containerOverlapping.Collect(ch)
	t.objectCopiesExpected.Collect(ch)
	t.objectCopiesFound.Collect(ch)
	t.objectCopiesMissing.Collect(ch)
	t.objectOverlapping.Collect(ch)
}

// UpdateMetrics implements the collector.Task interface.
func (t *ReportTask) UpdateMetrics(ctx context.Context) (map[string]int, error) {
	q := util.CmdArgsToStr(t.cmdArgs)
	queries := map[string]int{q: 0}
	e := &collector.TaskError{
		Cmd:     "swift-dispersion-report",
		CmdArgs: t.cmdArgs,
	}

	out, err := util.RunCommandWithTimeout(ctx, t.ctxTimeout, t.pathToExecutable, t.cmdArgs...)
	if err != nil {
		queries[q] = 1
		e.Inner = err
		return queries, e
	}

	// Remove errors from the output.
	var reportErrors float64
	out = t.errRe.ReplaceAllFunc(out, func(m []byte) []byte {
		// Skip unmounted errors. Recon collector's unmounted task will
		// take care of it.
		if !bytes.Contains(m, []byte("is unmounted")) {
			queries[q] = 1
			reportErrors++
			mList := t.errRe.FindStringSubmatch(string(m))
			if len(mList) > 0 {
				e.Inner = errors.New(mList[2])
				e.Hostname = mList[1]
				logg.Info(e.Error())
			}
		}
		return []byte{}
	})
	t.errors.Set(reportErrors)

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
	if err != nil {
		// Removing errors from output might have resulted in empty lines
		// therefore we remove whitespace before logging.
		out = bytes.TrimSpace(out)

		queries[q] = 1
		e.Inner = err
		e.Hostname = ""
		e.CmdOutput = string(out)
		return queries, e
	}

	cntr := data.Container
	if cntr.Expected > 0 && cntr.Found > 0 {
		cntr.Missing = cntr.Expected - cntr.Found
	}
	t.containerCopiesExpected.Set(float64(cntr.Expected))
	t.containerCopiesFound.Set(float64(cntr.Found))
	t.containerCopiesMissing.Set(float64(cntr.Missing))
	t.containerOverlapping.Set(float64(cntr.Overlapping))

	obj := data.Object
	if obj.Expected > 0 && obj.Found > 0 {
		obj.Missing = obj.Expected - obj.Found
	}
	t.objectCopiesExpected.Set(float64(obj.Expected))
	t.objectCopiesFound.Set(float64(obj.Found))
	t.objectCopiesMissing.Set(float64(obj.Missing))
	t.objectOverlapping.Set(float64(obj.Overlapping))

	return queries, nil
}
