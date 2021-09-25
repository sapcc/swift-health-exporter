// Copyright 2021 SAP SE
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
	"github.com/prometheus/client_golang/prometheus"
)

// Collector holds a collection of Task(s) and implements the prometheus.Collector
// interface.
//
// The metric values are updated independently of the Collector through Scraper.
type Collector struct {
	Tasks map[string]Task // map of task name to Task
}

// New returns a new Collector.
func New() *Collector {
	return &Collector{
		Tasks: make(map[string]Task),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, t := range c.Tasks {
		t.DescribeMetrics(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for _, t := range c.Tasks {
		t.CollectMetrics(ch)
	}
}
