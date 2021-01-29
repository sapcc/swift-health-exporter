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
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

// DispersionCollector implements the prometheus.Collector interface.
type DispersionCollector struct {
	ctxTimeout       time.Duration
	pathToExecutable string
	policies         []string

	// unmountedErrRe is used to match unmounted errors.
	// E.g.:
	//   ERROR: 10.0.0.1:6000/swift-09 is unmounted -- This will cause...
	unmountedErrRe *regexp.Regexp

	exitCode                typedDesc
	containerCopiesExpected typedDesc
	containerCopiesFound    typedDesc
	containerCopiesMissing  typedDesc
	containerOverlapping    typedDesc
	objectCopiesExpected    typedDesc
	objectCopiesFound       typedDesc
	objectCopiesMissing     typedDesc
	objectOverlapping       typedDesc
}

// NewDispersionCollector creates a new DispersionCollector.
func NewDispersionCollector(pathToExecutable string, ctxTimeout time.Duration, policies []string) *DispersionCollector {
	return &DispersionCollector{
		ctxTimeout:       ctxTimeout,
		pathToExecutable: pathToExecutable,
		policies:         policies,
		unmountedErrRe:   regexp.MustCompile(`(?m)^ERROR:\s*\d+\.\d+\.\d+\.\d+:\d+\/[a-zA-Z0-9-]+\s*is\s*unmounted.*$`),
		exitCode: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_task_exit_code",
				"The exit code for a Swift Dispersion Report query execution.",
				[]string{"query"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerCopiesExpected: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_expected",
				"Expected container copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerCopiesFound: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_found",
				"Found container copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerCopiesMissing: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_copies_missing",
				"Missing container copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		containerOverlapping: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_container_overlapping",
				"Expected container copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesExpected: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_expected",
				"Expected object copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesFound: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_found",
				"Found object copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectCopiesMissing: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_copies_missing",
				"Missing object copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
		objectOverlapping: typedDesc{
			desc: prometheus.NewDesc(
				"swift_dispersion_object_overlapping",
				"Expected object copies reported by the swift-dispersion-report tool.",
				[]string{"policy"}, nil),
			valueType: prometheus.GaugeValue,
		},
	}
}

// Describe implements the prometheus.Collector interface.
func (c *DispersionCollector) Describe(ch chan<- *prometheus.Desc) {
	c.exitCode.describe(ch)
	c.containerCopiesExpected.describe(ch)
	c.containerCopiesFound.describe(ch)
	c.containerCopiesMissing.describe(ch)
	c.containerOverlapping.describe(ch)
	c.objectCopiesExpected.describe(ch)
	c.objectCopiesFound.describe(ch)
	c.objectCopiesMissing.describe(ch)
	c.objectOverlapping.describe(ch)
}

// Collect implements the prometheus.Collector interface.
// Ubi patch: collect from multiple policies instead of the default one
func (c *DispersionCollector) Collect(ch chan<- prometheus.Metric) {
	exitCode := 0
	policies := c.policies
	// Keep it simple at time, use as many workers as policies
	workers := len(policies)
	cn := make(chan string)

	var wg sync.WaitGroup
	wg.Add(workers)
	for ii := 0; ii < workers; ii++ {
		go func(cn chan string) {
			for {
				policy, more := <-cn
				if more == false {
					wg.Done()
					return
				}
				cmdArg := []string{"--dump-json", "-P", policy}
				out, err := runCommandWithTimeout(c.ctxTimeout, c.pathToExecutable, cmdArg...)
				if err == nil {
					// Get rid of unmounted errors.
					out = c.unmountedErrRe.ReplaceAll(out, []byte{})
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
						err = fmt.Errorf("%s: output follows:\n%s", err.Error(), string(out))
					} else {
						cntr := data.Container
						if cntr.Expected > 0 && cntr.Found > 0 {
							cntr.Missing = cntr.Expected - cntr.Found
						}
						ch <- c.containerCopiesExpected.mustNewConstMetric(float64(cntr.Expected), policy)
						ch <- c.containerCopiesFound.mustNewConstMetric(float64(cntr.Found), policy)
						ch <- c.containerCopiesMissing.mustNewConstMetric(float64(cntr.Missing), policy)
						ch <- c.containerOverlapping.mustNewConstMetric(float64(cntr.Overlapping), policy)

						obj := data.Object
						if obj.Expected > 0 && obj.Found > 0 {
							obj.Missing = obj.Expected - obj.Found
						}
						ch <- c.objectCopiesExpected.mustNewConstMetric(float64(obj.Expected), policy)
						ch <- c.objectCopiesFound.mustNewConstMetric(float64(obj.Found), policy)
						ch <- c.objectCopiesMissing.mustNewConstMetric(float64(obj.Missing), policy)
						ch <- c.objectOverlapping.mustNewConstMetric(float64(obj.Overlapping), policy)
					}
				}
				if err != nil {
					exitCode = 1
					logg.Error("swift dispersion: %s: %s", cmdArg, err.Error())
				}
				//exit with the arguments to let user figure things out
				ch <- c.exitCode.mustNewConstMetric(float64(exitCode), cmdArg[2])
			}
		}(cn)
	}
	for _, a := range policies {
		cn <- a
	}
	close(cn)
	wg.Wait()
}
