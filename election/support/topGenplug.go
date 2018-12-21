package support

import (
	"math/big"
	"sort"
)

func getValidator_Order(weight []Pnormalized, ValidatorNum int, BackupNum int, RandSeed *big.Int) ([]Strallyint, []Strallyint, []Strallyint) {
	Master, weight := GetList(weight, ValidatorNum, RandSeed.Int64())

	Backup, weight := GetList(weight, BackupNum, RandSeed.Int64())

	Candidate, weight := GetList(weight, len(weight), RandSeed.Int64())
	return Master, Backup, Candidate
}
func getValidator_Direct(weight []Pnormalized, ValidatorNum int, BackupNum int, RandSeed *big.Int) ([]Strallyint, []Strallyint, []Strallyint) {
	TopAll, weight := GetList(weight, ValidatorNum+BackupNum, RandSeed.Int64())
	Candidate, weight := GetList(weight, len(weight), RandSeed.Int64())
	sort.Sort(SortStrallyint(TopAll))

	Master := []Strallyint{}
	Backup := []Strallyint{}
	for _, v := range TopAll {
		if len(Master) < ValidatorNum {
			Master = append(Master, v)
		}
		if len(Backup) < BackupNum {
			Backup = append(Backup, v)
		}
	}
	return Master, Backup, Candidate
}

type SortStrallyint []Strallyint

func (self SortStrallyint) Len() int {
	return len(self)
}
func (self SortStrallyint) Less(i, j int) bool {
	return self[i].Value > self[j].Value
}
func (self SortStrallyint) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}
