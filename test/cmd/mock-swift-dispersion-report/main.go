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
	`ERROR: 10.0.0.1:6000/sdb-01 is unmounted -- This will cause replicas designated for that device to be considered missing until resolved or the ring is updated.
ERROR: 10.0.0.2:6000/sdb-01 is unmounted -- This will cause replicas designated for that device to be considered missing until resolved or the ring is updated.
{"object": {"retries": 0, "missing_0": 655, "copies_expected": 1965, "pct_found": 100.0, "overlapping": 0, "copies_found": 1965}, "container": {"retries": 0, "copies_expected": 120, "pct_found": 100.0, "overlapping": 0, "copies_found": 120}}`)

func main() {
	var dumpJSON bool

	flag.BoolVarP(&dumpJSON, "dump-json", "j", false, "Dump dispersion report in json format.")
	flag.Parse()

	if dumpJSON {
		os.Stdout.Write(reportData)
	}
}
