package params

type LessDiskConfig struct {
	OptInterval     int64  // 操作间隔，单位秒
	HeightThreshold uint64 // 高度阈值
	TimeThreshold   int64  // 事件阈值，单位秒
}

var DefLessDiskConfig = &LessDiskConfig{
	OptInterval:     30,
	HeightThreshold: 1000,
	TimeThreshold:   2 * 60 * 60,
}
