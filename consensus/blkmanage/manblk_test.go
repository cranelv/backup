package blkmanage

import (
	"testing"

	"github.com/matrix/go-matrix/common"
)

func TestManBlkBasePlug_Prepare(t *testing.T) {
	test, _ := New(nil)
	base, _ := NewBlkBasePlug()
	test.RegisterManBLkPlugs("common", AVERSION, base)

	test.Prepare("common", AVERSION, 0, nil, common.Hash{1})
}
