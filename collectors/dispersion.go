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

package collectors

import (
	"encoding/json"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

var (
	dispersionCntrCopiesExpectedDesc = prometheus.NewDesc(
		"swift_dispersion_container_copies_expected",
		"Expected container copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionCntrCopiesFoundDesc = prometheus.NewDesc(
		"swift_dispersion_container_copies_found",
		"Found container copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionCntrCopiesMissingDesc = prometheus.NewDesc(
		"swift_dispersion_container_copies_missing",
		"Missing container copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionCntrOverlappingDesc = prometheus.NewDesc(
		"swift_dispersion_container_overlapping",
		"Expected container copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)

	dispersionObjCopiesExpectedDesc = prometheus.NewDesc(
		"swift_dispersion_object_copies_expected",
		"Expected object copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionObjCopiesFoundDesc = prometheus.NewDesc(
		"swift_dispersion_object_copies_found",
		"Found object copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionObjCopiesMissingDesc = prometheus.NewDesc(
		"swift_dispersion_object_copies_missing",
		"Missing object copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
	dispersionObjOverlappingDesc = prometheus.NewDesc(
		"swift_dispersion_object_overlapping",
		"Expected object copies reported by the swift-dispersion-report tool.",
		nil, nil,
	)
)

// DispersionCollector implements the prometheus.Collector interface.
type DispersionCollector struct {
	pathToExecutable string
}

// NewDispersionCollector creates a new DispersionCollector.
func NewDispersionCollector(pathToExecutable string) *DispersionCollector {
	return &DispersionCollector{
		pathToExecutable: pathToExecutable,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *DispersionCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface.
func (c *DispersionCollector) Collect(ch chan<- prometheus.Metric) {
	var dispersionReport struct {
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

	out, err := exec.Command(c.pathToExecutable, "-j").Output()
	if err != nil {
		logg.Error("swift-dispersion-report: %v", err)
		return
	}
	err = json.Unmarshal(out, &dispersionReport)
	if err != nil {
		logg.Error("swift-dispersion-report: %v", err)
		return
	}

	cntr := dispersionReport.Container
	if cntr.Expected > 0 && cntr.Found > 0 {
		cntr.Missing = cntr.Expected - cntr.Found
	}
	ch <- prometheus.MustNewConstMetric(
		dispersionCntrCopiesExpectedDesc,
		prometheus.GaugeValue, float64(cntr.Expected),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionCntrCopiesFoundDesc,
		prometheus.GaugeValue, float64(cntr.Found),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionCntrCopiesMissingDesc,
		prometheus.GaugeValue, float64(cntr.Missing),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionCntrOverlappingDesc,
		prometheus.GaugeValue, float64(cntr.Overlapping),
	)

	obj := dispersionReport.Object
	if obj.Expected > 0 && obj.Found > 0 {
		obj.Missing = obj.Expected - obj.Found
	}
	ch <- prometheus.MustNewConstMetric(
		dispersionObjCopiesExpectedDesc,
		prometheus.GaugeValue, float64(obj.Expected),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionObjCopiesFoundDesc,
		prometheus.GaugeValue, float64(obj.Found),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionObjCopiesMissingDesc,
		prometheus.GaugeValue, float64(obj.Missing),
	)
	ch <- prometheus.MustNewConstMetric(
		dispersionObjOverlappingDesc,
		prometheus.GaugeValue, float64(obj.Overlapping),
	)
}
