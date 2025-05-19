// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Task represents a collector task that deals with a specific set of metrics, their
// measurement, and reporting.
type Task interface {
	Name() string
	DescribeMetrics(ch chan<- *prometheus.Desc)
	CollectMetrics(ch chan<- prometheus.Metric)
	// UpdateMetrics returns a map of query to its exit code, and an error.
	UpdateMetrics(ctx context.Context) (queries map[string]int, err error)
}

// TaskError is the error type that a task can return.
type TaskError struct {
	Inner     error
	Cmd       string
	CmdArgs   []string
	CmdOutput string // optional
	Hostname  string // optional
}

func (e *TaskError) Error() string {
	s := e.Cmd + " " + strings.Join(e.CmdArgs, " ")
	if e.Hostname != "" {
		s += ": " + e.Hostname
	}
	s += ": " + e.Inner.Error()
	if e.CmdOutput != "" {
		s += "output follows:\n" + e.CmdOutput
	}
	return s
}
