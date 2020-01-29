package utils

import "time"

var clockSeconds int64 = 1

//TimeNow replaces time.Now in unit tests.
func TimeNow() time.Time {
	return time.Unix(clockSeconds, 0).UTC()
}
