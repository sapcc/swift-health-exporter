// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

func (d *typedDesc) describe(ch chan<- *prometheus.Desc) {
	ch <- d.desc
}

// collectorTask is the interface that a specific collector task must implement.
type collectorTask interface {
	describeMetrics(ch chan<- *prometheus.Desc)
	collectMetrics(ch chan<- prometheus.Metric, taskExitCodeTypedDesc typedDesc)
}

func cmdArgsToStr(cmdArgs []string) string {
	return strings.Join(cmdArgs, " ")
}
