package utils

import "time"

const clockSeconds int64 = 1

//TimeNow replaces time.Now in unit tests.
func TimeNow() time.Time {
	return time.Unix(clockSeconds, 0).UTC()
}
