package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/swift-health-exporter/internal/promhelper"
)

// Task represents a collector task that deals with a specific set of metrics, their
// measurement, and reporting.
type Task interface {
	Name() string
	DescribeMetrics(ch chan<- *prometheus.Desc)
	CollectMetrics(ch chan<- prometheus.Metric)
	// Measure returns a map of query to its exit code and an error.
	Measure() (queries map[string]int, err error)
}

// Collector implements the prometheus.Collector interface.
type Collector struct {
	MaxFailures     int
	Tasks           map[string]Task                  // map of task name to Task
	FailureCounts   map[string]int                   // map of task name to its failure count
	ExitCodeMetrics map[string]*promhelper.TypedDesc // map of task name to its exit code metric
}

func New(maxFailures int) *Collector {
	return &Collector{
		MaxFailures:     maxFailures,
		Tasks:           make(map[string]Task),
		FailureCounts:   make(map[string]int),
		ExitCodeMetrics: make(map[string]*promhelper.TypedDesc),
	}
}

func (c *Collector) AddTask(shouldAdd bool, task Task, exitCode *promhelper.TypedDesc) {
	if shouldAdd {
		c.Tasks[task.Name()] = task
		c.ExitCodeMetrics[task.Name()] = exitCode
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
		exitCodeTypedDesc := c.ExitCodeMetrics[t.Name()]
		queries, err := t.Measure()
		if err == nil {
			c.FailureCounts[t.Name()] = 0
		} else {
			c.FailureCounts[t.Name()]++
			if c.FailureCounts[t.Name()] >= c.MaxFailures {
				logg.Error(err.Error())
			}
		}

		// Report exit code metrics.
		for query, exitCode := range queries {
			if c.FailureCounts[t.Name()] < c.MaxFailures {
				// We only report the true exit code (other than success) when the max
				// failure count has been exceeded.
				exitCode = 0
			}
			ch <- exitCodeTypedDesc.MustNewConstMetric(float64(exitCode), query)
		}

		t.CollectMetrics(ch)
	}
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
