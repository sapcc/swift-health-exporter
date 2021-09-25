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

package recon

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// This value is overwritten in unit tests.
var isTest = false

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
