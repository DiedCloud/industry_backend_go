package main

import "math"

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) (res Stats) {
	if len(nums) < 2 {
		return Stats{}
	}

	stats := Stats{Min: math.MaxInt, Max: math.MinInt}
	for i := 1; i < len(nums); i++ {
		stats.Count += 1
		diff := nums[i] - nums[i-1]
		stats.Sum += diff
		if diff < stats.Min {
			stats.Min = diff
		}
		if diff > stats.Max {
			stats.Max = diff
		}
	}
	return stats
}
