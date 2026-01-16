// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-api-declarations/bininfo"
	"github.com/sapcc/go-bits/httpapi"
	"github.com/sapcc/go-bits/httpapi/pprofapi"
	"github.com/sapcc/go-bits/httpext"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
	"github.com/sapcc/go-bits/osext"
	flag "github.com/spf13/pflag"

	"github.com/sapcc/swift-health-exporter/internal/collector"
	"github.com/sapcc/swift-health-exporter/internal/collector/dispersion"
	"github.com/sapcc/swift-health-exporter/internal/collector/recon"
)

func main() {
	var (
		debug            bool
		showVersion      bool
		webListenAddress string

		maxFailures int

		// In large Swift clusters the dispersion-report tool takes time, therefore we have a higher default timeout value.
		dispersionTimeout   int64
		dispersionCollector bool

		reconTimeout                   int64
		reconHostTimeout               int
		noReconMD5Collector            bool
		reconDiskUsageCollector        bool
		reconDriveAuditCollector       bool
		reconQuarantinedCollector      bool
		reconReplicationCollector      bool
		reconShardingCollector         bool
		reconUnmountedCollector        bool
		reconUpdaterSweepTimeCollector bool
	)

	flag.BoolVar(&debug, "debug", false, "Enable debug mode.")
	flag.BoolVarP(&showVersion, "version", "v", false, "Report version string and exit.")
	flag.StringVar(&webListenAddress, "web.listen-address", "0.0.0.0:9520", "Exporter listening address.")

	flag.IntVar(&maxFailures, "collector.max-failures", 4, "Max allowed failures for a specific collector.")

	flag.Int64Var(&dispersionTimeout, "dispersion.timeout", 20, "Timeout value (in seconds) for the context that is used while executing the swift-dispersion-report command.")
	flag.BoolVar(&dispersionCollector, "collector.dispersion", false, "Enable dispersion collector.")

	flag.Int64Var(&reconTimeout, "recon.timeout", 4, "Timeout value (in seconds) for the context that is used while executing the swift-recon command.")
	flag.IntVar(&reconHostTimeout, "recon.timeout-host", 1, "Timeout value (in seconds) that is used for the '--timeout' flag (host timeout) of the swift-recon command.")
	flag.BoolVar(&noReconMD5Collector, "no-collector.recon.md5", false, "Disable MD5 collector.")
	flag.BoolVar(&reconDiskUsageCollector, "collector.recon.diskusage", false, "Enable disk usage collector.")
	flag.BoolVar(&reconDriveAuditCollector, "collector.recon.driveaudit", false, "Enable drive audit collector.")
	flag.BoolVar(&reconQuarantinedCollector, "collector.recon.quarantined", false, "Enable quarantined collector.")
	flag.BoolVar(&reconReplicationCollector, "collector.recon.replication", false, "Enable replication collector.")
	flag.BoolVar(&reconShardingCollector, "collector.recon.sharding", false, "Enable sharding collector.")
	flag.BoolVar(&reconUnmountedCollector, "collector.recon.unmounted", false, "Enable unmounted collector.")
	flag.BoolVar(&reconUpdaterSweepTimeCollector, "collector.recon.updater_sweep_time", false, "Enable updater sweep time collector.")
	flag.Parse()

	if showVersion {
		fmt.Println(bininfo.VersionOr("unknown"))
		return
	}

	logg.ShowDebug = debug || osext.GetenvBool("DEBUG")

	reconCollectorEnabled := !(noReconMD5Collector) ||
		reconDiskUsageCollector ||
		reconDriveAuditCollector ||
		reconQuarantinedCollector ||
		reconReplicationCollector ||
		reconShardingCollector ||
		reconUnmountedCollector ||
		reconUpdaterSweepTimeCollector

	if !reconCollectorEnabled && !(dispersionCollector) {
		logg.Fatal("no collector enabled")
	}

	registry := prometheus.DefaultRegisterer
	c := collector.New()
	s := collector.NewScraper(maxFailures)

	if dispersionCollector {
		execPath := getExecutablePath("SWIFT_DISPERSION_REPORT_PATH", "swift-dispersion-report")
		t := time.Duration(dispersionTimeout) * time.Second
		exitCode := dispersion.GetTaskExitCodeGaugeVec(registry)
		addTask(true, c, s, dispersion.NewReportTask(execPath, t), exitCode)
	}

	if reconCollectorEnabled {
		exitCode := recon.GetTaskExitCodeGaugeVec(registry)
		opts := &recon.TaskOpts{
			PathToExecutable: getExecutablePath("SWIFT_RECON_PATH", "swift-recon"),
			HostTimeout:      reconHostTimeout,
			CtxTimeout:       time.Duration(reconTimeout) * time.Second,
		}
		addTask(reconDiskUsageCollector, c, s, recon.NewDiskUsageTask(opts), exitCode)
		addTask(reconDriveAuditCollector, c, s, recon.NewDriveAuditTask(opts), exitCode)
		addTask(!(noReconMD5Collector), c, s, recon.NewMD5Task(opts), exitCode)
		addTask(reconQuarantinedCollector, c, s, recon.NewQuarantinedTask(opts), exitCode)
		addTask(reconReplicationCollector, c, s, recon.NewReplicationTask(opts), exitCode)
		addTask(reconShardingCollector, c, s, recon.NewShardingTask(opts), exitCode)
		addTask(reconUnmountedCollector, c, s, recon.NewUnmountedTask(opts), exitCode)
		addTask(reconUpdaterSweepTimeCollector, c, s, recon.NewUpdaterSweepTask(opts), exitCode)
	}

	prometheus.MustRegister(c)

	ctx := httpext.ContextWithSIGINT(context.Background(), 1*time.Second)

	// Run the scraper at least once so that the metric values are updated before a
	// Prometheus scrape.
	s.UpdateAllMetrics(ctx)

	// Start scraper loop.
	go s.Run(ctx)

	// Collect HTTP handlers.
	handler := httpapi.Compose(
		landingPageAPI{},
		httpapi.WithoutLogging(),
		pprofapi.API{IsAuthorized: pprofapi.IsRequestFromLocalhost},
	)
	smux := http.NewServeMux()
	smux.Handle("/", handler)
	smux.Handle("/metrics", promhttp.Handler())

	must.Succeed(httpext.ListenAndServeContext(ctx, webListenAddress, smux))
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
