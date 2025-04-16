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
	"os"

	flag "github.com/spf13/pflag"
)

var reportData = []byte(
	`ERROR: 10.0.0.1:6001/sdb-01: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.1:6001/sdb-01: Giving up on /123/AUTH_123/dispersion_objects_0/dispersion_01: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.1:6001/sdb-02: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.1:6001/sdb-02: Giving up on /456/AUTH_456/dispersion_objects_0/dispersion_02: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.2:6001/sdb-01: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.2:6001/sdb-01: Giving up on /789/AUTH_789/dispersion_objects_0/dispersion_01: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.2:6001/sdb-02: [Errno 111] ECONNREFUSED
ERROR: 10.0.0.2:6001/sdb-02: Giving up on /012/AUTH_012/dispersion_objects_0/dispersion_02: [Errno 111] ECONNREFUSED
{"object": {"retries": 0, "missing_0": 655, "copies_expected": 1965, "pct_found": 96.69, "overlapping": 0, "copies_found": 1900, "missing_0": 60, "missing_1": 5}, "container": {"retries": 0, "copies_expected": 120, "pct_found": 91.66, "overlapping": 0, "copies_found": 110, "missing_0": 10}}`)

func main() {
	var dumpJSON bool

	flag.BoolVarP(&dumpJSON, "dump-json", "j", false, "Dump dispersion report in json format.")
	flag.Parse()

	if dumpJSON {
		os.Stdout.Write(reportData)
	}
}
