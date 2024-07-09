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
