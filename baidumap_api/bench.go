package main

import (
	"math"
)

//基准距离算法
func Calc(lat1, log1, lat2, log2 float64) float64 {
	radius := float64(6378100) // 6378137	地球半径，单位m
	rad := math.Pi / 180.0     //弧度

	lat1 = lat1 * rad
	log1 = log1 * rad
	lat2 = lat2 * rad
	log2 = log2 * rad

	theta := log2 - log1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))

	return float64(dist * radius)
}
