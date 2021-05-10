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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// Collector implements the prometheus.Collector interface.
type Collector struct {
	ctxTimeout       time.Duration
	pathToExecutable string

	// errRe is used to match errors and capture the hostname and error message.
	// E.g.:
	//   ERROR: 10.0.0.1:6000/swift-09: [Errno 111] ECONNREFUSED
	errRe *regexp.Regexp

	exitCode                *promhelper.TypedDesc
	errors                  *promhelper.TypedDesc
	containerCopiesExpected *promhelper.TypedDesc
	containerCopiesFound    *promhelper.TypedDesc
	containerCopiesMissing  *promhelper.TypedDesc
	containerOverlapping    *promhelper.TypedDesc
	objectCopiesExpected    *promhelper.TypedDesc
	objectCopiesFound       *promhelper.TypedDesc
	objectCopiesMissing     *promhelper.TypedDesc
	objectOverlapping       *promhelper.TypedDesc
}

// NewCollector creates a new DispersionCollector.
func NewCollector(pathToExecutable string, ctxTimeout time.Duration) *Collector {
	return &Collector{
		ctxTimeout:       ctxTimeout,
		pathToExecutable: pathToExecutable,
		errRe:            regexp.MustCompile(`(?m)^ERROR:\s*([\d.]+)\S*\s*(.*)$`),
		exitCode: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_task_exit_code",
			"The exit code for a Swift dispersion report query execution.", []string{"query"}),
		errors: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_errors",
			"The number of errors in the Swift dispersion report.", nil),
		containerCopiesExpected: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_container_copies_expected",
			"Expected container copies reported by the swift-dispersion-report tool.", nil),
		containerCopiesFound: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_container_copies_found",
			"Found container copies reported by the swift-dispersion-report tool.", nil),
		containerCopiesMissing: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_container_copies_missing",
			"Missing container copies reported by the swift-dispersion-report tool.", nil),
		containerOverlapping: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_container_overlapping",
			"Expected container copies reported by the swift-dispersion-report tool.", nil),
		objectCopiesExpected: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_object_copies_expected",
			"Expected object copies reported by the swift-dispersion-report tool.", nil),
		objectCopiesFound: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_object_copies_found",
			"Found object copies reported by the swift-dispersion-report tool.", nil),
		objectCopiesMissing: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_object_copies_missing",
			"Missing object copies reported by the swift-dispersion-report tool.", nil),
		objectOverlapping: promhelper.NewGaugeTypedDesc(
			"swift_dispersion_object_overlapping",
			"Expected object copies reported by the swift-dispersion-report tool.", nil),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.exitCode.Describe(ch)
	c.errors.Describe(ch)
	c.containerCopiesExpected.Describe(ch)
	c.containerCopiesFound.Describe(ch)
	c.containerCopiesMissing.Describe(ch)
	c.containerOverlapping.Describe(ch)
	c.objectCopiesExpected.Describe(ch)
	c.objectCopiesFound.Describe(ch)
	c.objectCopiesMissing.Describe(ch)
	c.objectOverlapping.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	exitCode := 0
	cmdArg := "--dump-json"
	ctx, cancel := context.WithTimeout(context.Background(), c.ctxTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, c.pathToExecutable, cmdArg).CombinedOutput()
	if err == nil {
		// Remove errors from the output.
		var reportErrors float64
		out = c.errRe.ReplaceAllFunc(out, func(m []byte) []byte {
			// Skip unmounted errors. Recon collector's unmounted task will
			// take care of it.
			if !bytes.Contains(m, []byte("is unmounted")) {
				reportErrors++
				exitCode = 1
				mList := c.errRe.FindStringSubmatch(string(m))
				if len(mList) > 0 {
					host := mList[1]
					logg.Error("swift dispersion: %s: %s: %s", cmdArg, host, mList[2])
				}
			}
			return []byte{}
		})
		ch <- c.errors.MustNewConstMetric(reportErrors)

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
			err = fmt.Errorf("%s: output follows:\n%s", err.Error(), string(out))
		} else {
			cntr := data.Container
			if cntr.Expected > 0 && cntr.Found > 0 {
				cntr.Missing = cntr.Expected - cntr.Found
			}
			ch <- c.containerCopiesExpected.MustNewConstMetric(float64(cntr.Expected))
			ch <- c.containerCopiesFound.MustNewConstMetric(float64(cntr.Found))
			ch <- c.containerCopiesMissing.MustNewConstMetric(float64(cntr.Missing))
			ch <- c.containerOverlapping.MustNewConstMetric(float64(cntr.Overlapping))

			obj := data.Object
			if obj.Expected > 0 && obj.Found > 0 {
				obj.Missing = obj.Expected - obj.Found
			}
			ch <- c.objectCopiesExpected.MustNewConstMetric(float64(obj.Expected))
			ch <- c.objectCopiesFound.MustNewConstMetric(float64(obj.Found))
			ch <- c.objectCopiesMissing.MustNewConstMetric(float64(obj.Missing))
			ch <- c.objectOverlapping.MustNewConstMetric(float64(obj.Overlapping))
		}
	}
	if err != nil {
		exitCode = 1
		logg.Error("swift dispersion: %s: %s", cmdArg, err.Error())
	}

	ch <- c.exitCode.MustNewConstMetric(float64(exitCode), cmdArg)
}
