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

package collector

import (
	"encoding/json"
	"strconv"
	"time"
)

const clockSeconds int64 = 1

// timeNow replaces time.Now in unit tests.
func timeNow() time.Time {
	return time.Unix(clockSeconds, 0).UTC()
}

// For fields that are sometimes missing, sometimes an integer, sometimes a
// string.
type flexibleUint64 uint64

// UnmarshalJSON implements the json.Unmarshaler interface.
func (value *flexibleUint64) UnmarshalJSON(b []byte) error {
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
		// We don't care about the error here, default value of 0 is ok.
		v, _ := strconv.ParseUint(str, 10, 64)
		*value = flexibleUint64(v)
		return nil
	}

	var v uint64
	err := json.Unmarshal(b, &v)
	*value = flexibleUint64(v)
	return err
}
