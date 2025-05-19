// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
)

// How long to wait before re-running the scraper for all tasks.
const scrapeInterval = 1 * time.Minute

// Scraper holds a collection of Task(s) and other parameters that are required for
// scraping (read: updating) the metric values.
type Scraper struct {
	MaxFailures      int
	Tasks            map[string]Task                 // key = task name
	FailureCount     map[string]int                  // map of task name to its failure count
	ExitCodeGaugeVec map[string]*prometheus.GaugeVec // map of task name to its relevant exit code GaugeVec
}

// NewScraper returns a new Scraper.
func NewScraper(maxFailures int) *Scraper {
	return &Scraper{
		MaxFailures:      maxFailures,
		Tasks:            make(map[string]Task),
		FailureCount:     make(map[string]int),
		ExitCodeGaugeVec: make(map[string]*prometheus.GaugeVec),
	}
}

// Run updates the metrics for all tasks periodically as per the scrapeInterval.
func (s *Scraper) Run(ctx context.Context) {
	for {
		startedAt := time.Now()
		s.UpdateAllMetrics(ctx)
		// Slow down if UpdateAllMetrics() finished faster than the desired scrape
		// interval.
		sleepDuration := scrapeInterval - time.Since(startedAt)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}
	}
}

func (s *Scraper) UpdateAllMetrics(ctx context.Context) {
	for _, t := range s.Tasks {
		name := t.Name()
		exitCodeGaugeVec := s.ExitCodeGaugeVec[name]
		queries, err := t.UpdateMetrics(ctx)
		if err == nil {
			s.FailureCount[name] = 0
		} else {
			s.FailureCount[name]++
			if s.FailureCount[name] >= s.MaxFailures {
				logg.Error(err.Error())
			}
		}

		// Update exit code metric(s).
		for query, exitCode := range queries {
			if s.FailureCount[name] < s.MaxFailures {
				// We only report a non-success exit code (i.e. 1) when the max
				// failure count has been exceeded.
				exitCode = 0
			}
			exitCodeGaugeVec.WithLabelValues(query).Set(float64(exitCode))
		}
	}
}
