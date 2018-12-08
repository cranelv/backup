package mc

type RewardRateCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励

	MinerOutRate        uint64 //出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type BlkRewardCfg struct {
	MinerMount     uint64 //矿工奖励单位man
	ValidatorMount uint64 //验证者奖励 单位man
	RewardRate     RewardRateCfg
}

type TxsRewardCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励

	RewardRate RewardRateCfg
}

type LotteryInfo struct {
	PrizeLevel uint8
	PrizeNum   uint64
	PrizeMoney uint64 //奖励金额 单位man
}

type LotteryCfg struct {
	LotteryCalc string
	LotteryInfo []LotteryInfo
}

type InterestRate struct {
	VIPLevel uint8
	Rate     uint64 //(分母待定为1000w)
}

type InterestCfg struct {
	CalcInterval uint64
	PayInterval  uint64
	InterestRate []InterestRate
}

type SlashCfg struct {
	SlashRate uint64
}
