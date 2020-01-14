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

	"github.com/alecthomas/kingpin"
)

var reportData = []byte(
	`ERROR: 10.0.0.1:6001/sdb-01: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.1:6001/sdb-02: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.1:6001/sdb-03: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.1:6001/sdb-04: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.1:6001/sdb-05: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.2:6001/sdb-01: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.2:6001/sdb-02: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.2:6001/sdb-03: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.2:6001/sdb-04: [Errno 111] ECONNREFUSED
	ERROR: 10.0.0.2:6001/sdb-05: [Errno 111] ECONNREFUSED`)

func main() {
	dumpJSONFlag := kingpin.Flag("dump-json", "Dump dispersion report in json format.").Short('j').Required().Bool()
	kingpin.Parse()
	if *dumpJSONFlag {
		os.Stdout.Write(reportData)
	}
}