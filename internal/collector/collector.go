// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

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
