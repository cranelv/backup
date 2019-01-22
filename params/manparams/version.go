package manparams

import "bytes"

const (
	VersionAlpha = "1.0.0.0"
)

var VersionList [][]byte

func init() {
	VersionList = [][]byte{[]byte(VersionAlpha)}
}

func IsCorrectVersion(version []byte) bool {
	if len(version) == 0 {
		return false
	}
	for _, item := range VersionList {
		if bytes.Equal(version, item) {
			return true
		}
	}
	return false
}
