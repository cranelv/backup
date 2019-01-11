package supertxsstate

import (
	"fmt"
	"testing"

	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/params/manparams"
)

func Test_newManager(t *testing.T) {
	log.InitLog(3)
	a := newManager(manparams.VersionAlpha)
	var slash mc.SlashCfg
	slash.SlashRate = 7500
	a.Check(mc.MSKeySlashCfg, slash)
	fmt.Println(a.Output(mc.MSKeySlashCfg, slash))

	var electMiner mc.ElectMinerNumStruct
	electMiner.MinerNum = 1
	a.Check(mc.MSKeyElectMinerNum, electMiner)
	fmt.Println(a.Output(mc.MSKeyElectMinerNum, electMiner))

	var black []common.Address
	black = append(black, common.HexToAddress("0x01"))
	a.Check(mc.MSKeyElectBlackList, black)
	fmt.Println(a.Output(mc.MSKeyElectBlackList, black))

	var white []common.Address
	white = append(white, common.HexToAddress("0x02"))
	a.Check(mc.MSKeyElectWhiteList, white)
	fmt.Println(a.Output(mc.MSKeyElectWhiteList, white))
}
