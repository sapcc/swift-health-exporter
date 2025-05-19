// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package recon

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// IsTest variable is set to true in unit tests.
var IsTest = false

// TaskOpts holds common parameters that are used by all recon tasks.
type TaskOpts struct {
	PathToExecutable string
	HostTimeout      int
	CtxTimeout       time.Duration
}

// GetTaskExitCodeGaugeVec returns a *prometheus.GaugeVec for use with recon tasks.
func GetTaskExitCodeGaugeVec(r prometheus.Registerer) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "swift_recon_task_exit_code",
			Help: "The exit code for a Swift Recon query execution.",
		}, []string{"query"},
	)
	r.MustRegister(gaugeVec)
	return gaugeVec
}
