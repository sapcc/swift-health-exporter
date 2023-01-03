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

package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/httpext"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/collector/dispersion"
	"github.com/sapcc/swift-health-exporter/internal/collector/recon"
)

var cli struct {
	Debug            bool   `env:"DEBUG" help:"Enable debug mode."`
	WebListenAddress string `name:"web.listen-address" default:"0.0.0.0:9520" help:"Exporter listening address."`

	MaxFailures int `name:"collector.max-failures" default:"4" help:"Max allowed failures for a specific collector."`

	// In large Swift clusters the dispersion-report tool takes time, therefore we have a higher default timeout value.
	DispersionTimeout   int64 `name:"dispersion.timeout" default:"20" help:"Timeout value (in seconds) for the context that is used while executing the swift-dispersion-report command."`
	DispersionCollector bool  `name:"collector.dispersion" help:"Enable dispersion collector."`

	ReconTimeout                   int64 `name:"recon.timeout" default:"4" help:"Timeout value (in seconds) for the context that is used while executing the swift-recon command."`
	ReconHostTimeout               int   `name:"recon.timeout-host" default:"1" help:"Timeout value (in seconds) that is used for the '--timeout' flag (host timeout) of the swift-recon command."`
	NoReconMD5Collector            bool  `name:"no-collector.recon.md5" help:"Disable MD5 collector."`
	ReconDiskUsageCollector        bool  `name:"collector.recon.diskusage" help:"Enable disk usage collector."`
	ReconDriveAuditCollector       bool  `name:"collector.recon.driveaudit" help:"Enable drive audit collector."`
	ReconQuarantinedCollector      bool  `name:"collector.recon.quarantined" help:"Enable quarantined collector."`
	ReconReplicationCollector      bool  `name:"collector.recon.replication" help:"Enable replication collector."`
	ReconUnmountedCollector        bool  `name:"collector.recon.unmounted" help:"Enable unmounted collector."`
	ReconUpdaterSweepTimeCollector bool  `name:"collector.recon.updater_sweep_time" help:"Enable updater sweep time collector."`
}

func main() {
	kong.Parse(&cli)
	reconCollectorEnabled := !(cli.NoReconMD5Collector) ||
		cli.ReconDiskUsageCollector ||
		cli.ReconDriveAuditCollector ||
		cli.ReconQuarantinedCollector ||
		cli.ReconReplicationCollector ||
		cli.ReconUnmountedCollector ||
		cli.ReconUpdaterSweepTimeCollector

	if !reconCollectorEnabled && !(cli.DispersionCollector) {
		logg.Fatal("no collector enabled")
	}

	registry := prometheus.DefaultRegisterer
	c := collector.New()
	s := collector.NewScraper(cli.MaxFailures)

	if cli.DispersionCollector {
		execPath := getExecutablePath("SWIFT_DISPERSION_REPORT_PATH", "swift-dispersion-report")
		t := time.Duration(cli.DispersionTimeout) * time.Second
		exitCode := dispersion.GetTaskExitCodeGaugeVec(registry)
		addTask(true, c, s, dispersion.NewReportTask(execPath, t), exitCode)
	}

	if reconCollectorEnabled {
		exitCode := recon.GetTaskExitCodeGaugeVec(registry)
		opts := &recon.TaskOpts{
			PathToExecutable: getExecutablePath("SWIFT_RECON_PATH", "swift-recon"),
			HostTimeout:      cli.ReconHostTimeout,
			CtxTimeout:       time.Duration(cli.ReconTimeout) * time.Second,
		}
		addTask(cli.ReconDiskUsageCollector, c, s, recon.NewDiskUsageTask(opts), exitCode)
		addTask(cli.ReconDriveAuditCollector, c, s, recon.NewDriveAuditTask(opts), exitCode)
		addTask(!(cli.NoReconMD5Collector), c, s, recon.NewMD5Task(opts), exitCode)
		addTask(cli.ReconQuarantinedCollector, c, s, recon.NewQuarantinedTask(opts), exitCode)
		addTask(cli.ReconReplicationCollector, c, s, recon.NewReplicationTask(opts), exitCode)
		addTask(cli.ReconUnmountedCollector, c, s, recon.NewUnmountedTask(opts), exitCode)
		addTask(cli.ReconUpdaterSweepTimeCollector, c, s, recon.NewUpdaterSweepTask(opts), exitCode)
	}

	prometheus.MustRegister(c)

	// Run the scraper at least once so that the metric values are updated before a
	// Prometheus scrape.
	s.UpdateAllMetrics()

	// Start scraper loop.
	go s.Run()

	// Collect HTTP handlers.
	handler := httpapi.Compose(
		landingPageAPI{},
		httpapi.WithoutLogging(),
	)
	http.Handle("/", handler)
	http.Handle("/metrics", promhttp.Handler())

	err := httpext.ListenAndServeContext(httpext.ContextWithSIGINT(context.Background(), 1*time.Second), cli.WebListenAddress, nil)
	if err != nil {
		logg.Fatal(err.Error())
	}
}

// getExecutablePath gets the path to an executable from the environment
// variable using the envKey (if defined). Otherwise it attempts to find this
// path in the directories named by the "PATH" environment variable.
//
// exec.Command() already uses LookPath() in case an executable name is
// provided instead of a path, but we do this manually for two reasons:
// 1. To terminate the program early in case the executable path could not be found.
// 2. To save multiple LookPath() calls for the same executable.
func getExecutablePath(envKey, fileName string) string {
	val := os.Getenv(envKey)
	if val != "" {
		return val
	}

	path, err := exec.LookPath(fileName)
	if err != nil {
		logg.Fatal(err.Error())
	}

	return path
}

type landingPageAPI struct{}

func (landingPageAPI) AddTo(r *mux.Router) {
	r.Methods("GET", "HEAD").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageBytes := []byte(`<html>
<head><title>Swift Health Exporter</title></head>
<body>
<h1>Swift Health Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="https://github.com/sapcc/swift-health-exporter">Source Code</a></p>
</body>
</html>`)

		_, err := w.Write(pageBytes)
		if err != nil {
			logg.Error(err.Error())
		}
	})
}

// addTask adds a Task to the given Collector and the Scraper along
// with its corresponding exit code GaugeVec.
func addTask(
	shouldAdd bool,
	c *collector.Collector,
	s *collector.Scraper,
	t collector.Task,
	exitCode *prometheus.GaugeVec) {

	if shouldAdd {
		name := t.Name()
		c.Tasks[name] = t
		s.Tasks[name] = t
		s.ExitCodeGaugeVec[name] = exitCode
	}
}
