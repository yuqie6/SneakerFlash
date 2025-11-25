package vip

// GrowthLevel 定义成长等级阈值（单位：分），越往后间隔越大。
// L1: 0-99,999 分（0-999元）
// L2: 100,000-499,999 分（1,000-4,999元）
// L3: 500,000-1,999,999 分（5,000-19,999元）
// L4: 2,000,000 分以上（20,000元+）
var growthThresholds = []struct {
	Level int
	Min   int64
}{
	{Level: 1, Min: 0},
	{Level: 2, Min: 100_000},
	{Level: 3, Min: 500_000},
	{Level: 4, Min: 2_000_000},
}

// CalcGrowthLevel 按累计实付金额（分）计算成长等级。
func CalcGrowthLevel(totalSpentCents int64) int {
	level := 1
	for _, th := range growthThresholds {
		if totalSpentCents >= th.Min && th.Level > level {
			level = th.Level
		}
	}
	return level
}
