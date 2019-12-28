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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-bits/httpee"
	"github.com/sapcc/go-bits/logg"
)

func main() {
	// this port has been allocated for a Swift health exporter
	// See: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	listenAddr := ":9520"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write(landingPageBytes) })
	http.Handle("/metrics", promhttp.Handler())
	logg.Info("listening on " + listenAddr)
	err := httpee.ListenAndServeContext(httpee.ContextWithSIGINT(context.Background()), listenAddr, nil)
	if err != nil {
		logg.Fatal(err.Error())
	}
}

var landingPageBytes = []byte(`<html>
<head><title>Swift Health Exporter</title></head>
<body>
<h1>Swift Health Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="https://github.com/sapcc/swift-health-exporter">Source Code</a></p>
</body>
</html>`)
