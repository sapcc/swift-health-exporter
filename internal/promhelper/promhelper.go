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

package promhelper

import "github.com/prometheus/client_golang/prometheus"

// TypedDesc holds the descriptor and value type for a specific metric.
type TypedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

// NewGaugeTypedDesc returns a new TypedDesc with Gauge as the value type.
func NewGaugeTypedDesc(fqName, help string, variableLabels []string) *TypedDesc {
	return &TypedDesc{
		desc:      prometheus.NewDesc(fqName, help, variableLabels, nil),
		valueType: prometheus.GaugeValue,
	}
}

// MustNewConstMetric is a wrapper for prometheus.MustNewConstMetric.
func (d *TypedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

// Describe describe the underlying metric on the channel.
func (d *TypedDesc) Describe(ch chan<- *prometheus.Desc) {
	ch <- d.desc
}
