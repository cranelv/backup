package manparams

import (
	"bytes"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

const (
	VersionAlpha = "1.0.0.0"
	VersionBeta  = "1.0.0.1"
)

var VersionList [][]byte
var VersionSignatureMap map[string][]common.Signature

func init() {
	VersionList = [][]byte{[]byte(VersionAlpha), []byte(VersionBeta)}
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

func GetVersionSignature(parentBlock *types.Block, version []byte) []common.Signature {
	if len(version) == 0 {
		return nil
	}
	if string(version) == string(parentBlock.Version()) {
		return parentBlock.VersionSignature()
	}
	if sig, ok := VersionSignatureMap[string(version)]; ok {
		return sig
	}

	return nil
}
