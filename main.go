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
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/httpee"
	"github.com/sapcc/go-bits/logg"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/collector/dispersion"
	"github.com/sapcc/swift-health-exporter/internal/collector/recon"
)

func main() {
	logg.ShowDebug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	// In large Swift clusters the dispersion-report tool takes time, hence the longer timeout.
	maxFailures := kingpin.Flag("collector.max-failures", "Max allowed failures for a specific collector").Default("4").Int()
	dispersionTimeout := kingpin.Flag("dispersion.timeout", "The swift-dispersion-report command context timeout value (in seconds).").Default("20").Int64()
	dispersionCollector := kingpin.Flag("collector.dispersion", "Enable dispersion collector.").Bool()
	reconTimeout := kingpin.Flag("recon.timeout", "The swift-recon command context timeout value (in seconds).").Default("4").Int64()
	reconHostTimeout := kingpin.Flag("recon.timeout-host", "The swift-recon host timeout value (in seconds).").Default("1").Int()
	noReconMD5Collector := kingpin.Flag("no-collector.recon.md5", "Disable MD5 collector.").Bool()
	reconDiskUsageCollector := kingpin.Flag("collector.recon.diskusage", "Enable disk usage collector.").Bool()
	reconDriveAuditCollector := kingpin.Flag("collector.recon.driveaudit", "Enable drive audit collector.").Bool()
	reconQuarantinedCollector := kingpin.Flag("collector.recon.quarantined", "Enable quarantined collector.").Bool()
	reconReplicationCollector := kingpin.Flag("collector.recon.replication", "Enable replication collector.").Bool()
	reconUnmountedCollector := kingpin.Flag("collector.recon.unmounted", "Enable unmounted collector.").Bool()
	reconUpdaterSweepTimeCollector := kingpin.Flag("collector.recon.updater_sweep_time", "Enable updater sweep time collector.").Bool()

	kingpin.Parse()

	reconCollector := *reconDiskUsageCollector || *reconDriveAuditCollector || !(*noReconMD5Collector) ||
		*reconQuarantinedCollector || *reconReplicationCollector || *reconUnmountedCollector || *reconUpdaterSweepTimeCollector

	if !reconCollector && !(*dispersionCollector) {
		logg.Fatal("no collector enabled")
	}

	registry := prometheus.NewRegistry()
	c := collector.New()
	s := collector.NewScraper(*maxFailures)

	if *dispersionCollector {
		execPath := getExecutablePath("SWIFT_DISPERSION_REPORT_PATH", "swift-dispersion-report")
		t := time.Duration(*dispersionTimeout) * time.Second
		exitCode := dispersion.GetTaskExitCodeGaugeVec(registry)
		addTask(true, c, s, dispersion.NewReportTask(execPath, t), exitCode)
	}

	if reconCollector {
		exitCode := recon.GetTaskExitCodeGaugeVec(registry)
		opts := &recon.TaskOpts{
			PathToExecutable: getExecutablePath("SWIFT_RECON_PATH", "swift-recon"),
			HostTimeout:      *reconHostTimeout,
			CtxTimeout:       time.Duration(*reconTimeout) * time.Second,
		}
		addTask(*reconDiskUsageCollector, c, s, recon.NewDiskUsageTask(opts), exitCode)
		addTask(*reconDriveAuditCollector, c, s, recon.NewDriveAuditTask(opts), exitCode)
		addTask(!(*noReconMD5Collector), c, s, recon.NewMD5Task(opts), exitCode)
		addTask(*reconQuarantinedCollector, c, s, recon.NewQuarantinedTask(opts), exitCode)
		addTask(*reconReplicationCollector, c, s, recon.NewReplicationTask(opts), exitCode)
		addTask(*reconUnmountedCollector, c, s, recon.NewUnmountedTask(opts), exitCode)
		addTask(*reconUpdaterSweepTimeCollector, c, s, recon.NewUpdaterSweepTask(opts), exitCode)
	}

	registry.MustRegister(c)

	// Run the scraper at least once so that the metric values are updated before a
	// Prometheus scrape.
	s.UpdateAllMetrics()

	// Start scraper loop.
	go s.Run()

	// this port has been allocated for Swift health exporter
	// See: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	listenAddr := ":9520"
	mux := http.NewServeMux()
	mux.HandleFunc("/", landingPageHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	handler := logg.Middleware{}.Wrap(mux)
	logg.Info("listening on " + listenAddr)
	err := httpee.ListenAndServeContext(httpee.ContextWithSIGINT(context.Background(), 1*time.Second), listenAddr, handler)
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

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
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
