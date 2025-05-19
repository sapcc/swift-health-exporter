// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

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
