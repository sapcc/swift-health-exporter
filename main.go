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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/httpee"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/swift-health-exporter/collectors"
)

func main() {
	swiftDispersionReportPath := getExecutablePath("SWIFT_DISPERSION_REPORT", "swift-dispersion-report")

	prometheus.MustRegister(collectors.NewDispersionCollector(swiftDispersionReportPath))

	// this port has been allocated for Swift health exporter
	// See: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	listenAddr := ":9520"
	http.HandleFunc("/", landingPageHandler)
	http.Handle("/metrics", promhttp.Handler())
	logg.Info("listening on " + listenAddr)
	err := httpee.ListenAndServeContext(httpee.ContextWithSIGINT(context.Background()), listenAddr, nil)
	if err != nil {
		logg.Fatal(err.Error())
	}
}

// getExecutablePath gets the path to an executable from the environment
// variable using the envKey (if defined).
// Otherwise it attempts to find the path in the directories defined in the
// "PATH" environment variable.
func getExecutablePath(envKey, fileName string) string {
	val := os.Getenv(envKey)
	if val != "" {
		return val
	}

	path, err := exec.LookPath(fileName)
	if err != nil {
		logg.Fatal("could not find an executable named %q in the directories named by the PATH environment variable", fileName)
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
