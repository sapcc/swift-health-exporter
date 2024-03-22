// Copyright 2020 SAP SE
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
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/swift-health-exporter/internal/util"
)

const clockSeconds int64 = 1

// timeNow replaces time.Now in unit tests.
func timeNow() time.Time {
	return time.Unix(clockSeconds, 0).UTC()
}

// flexibleFloat64 is used for fields that are sometimes missing, sometimes an
// integer/float, and sometimes a string.
type flexibleFloat64 float64

// UnmarshalJSON implements the json.Unmarshaler interface.
func (value *flexibleFloat64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*value = 0
		return nil
	}

	if b[0] == '"' {
		var str string
		err := json.Unmarshal(b, &str)
		if err != nil {
			return err
		}

		if str == "None" {
			*value = -1
			return nil
		}

		//nolint:errcheck // We don't care about the error here, default value of 0 is ok.
		v, _ := strconv.ParseFloat(str, 64)
		*value = flexibleFloat64(v)
		return nil
	}

	var v float64
	err := json.Unmarshal(b, &v)
	*value = flexibleFloat64(v)
	return err
}

// reconHostOutputRx is used to extract per host output from an aggregate
// output of a recon command.
//
// Match group ref:
//
//	<1: host> <2: output>
var reconHostOutputRx = regexp.MustCompile(`(?m)^(?:->|!!) https?://([a-zA-Z0-9-.]+)\S*\s(.*)$`)

func splitOutputPerHost(output []byte, cmdArgs []string) (map[string][]byte, error) {
	matchList := reconHostOutputRx.FindAllSubmatch(output, -1)
	if len(matchList) == 0 {
		return nil, errors.New("command did not return any usable output")
	}

	result := make(map[string][]byte)
	for _, match := range matchList {
		hostname := string(match[1])
		data := match[2]

		logg.Debug("output from command 'swift-recon %s': %s: %s", util.CmdArgsToStr(cmdArgs), hostname, string(data))

		// convert Python literals to JSON (best effort)
		data = bytes.ReplaceAll(data, []byte(`u'`), []byte(`'`))
		data = bytes.ReplaceAll(data, []byte(`'`), []byte(`"`))
		data = bytes.ReplaceAll(data, []byte(`True`), []byte(`true`))
		data = bytes.ReplaceAll(data, []byte(`False`), []byte(`false`))
		data = bytes.ReplaceAll(data, []byte(`None`), []byte(`"None"`))
		data = bytes.ReplaceAll(data, []byte(`""None""`), []byte(`"None"`))
		// ^ We sometimes observe strings with the value "None".
		// The None -> "None" replacement introduces double quoting there which we need to compensate for.
		data = bytes.ReplaceAll(data, []byte(`\x`), []byte(`\\x`))
		// ^ Swift renames sharding containers to include \x which is interpreted as an invalid escape character.
		// Prefix with an additional `\` to pass unmarshalling and preserve naming.

		result[hostname] = data
	}

	return result, nil
}

func getSwiftReconOutputPerHost(ctxTimeout time.Duration, pathToExecutable string, cmdArgs ...string) (map[string][]byte, error) {
	out, err := util.RunCommandWithTimeout(ctxTimeout, pathToExecutable, cmdArgs...)
	if err != nil {
		return nil, err
	}

	return splitOutputPerHost(out, cmdArgs)
}
