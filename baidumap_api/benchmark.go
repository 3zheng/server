package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

const DIST_PER_DEGREE = math.Pi * 35433.88889 // πr/180° 解释为每度所表示的实际长度
const EARTH_RADIUS = 6378100.00               // 6378137	地球半径，单位m
const TEST_DATA_NUM = 3000000

//https://tech.meituan.com/lucene-distance.html 算法详细解释
//根据两个点的经纬度来计算两点之间的距离：lat1,log1表示点1的纬度和经度；lat2,log2表示点2的纬度和经度

type EarthPointPair struct {
	lat1 float64
	log1 float64
	lat2 float64
	log2 float64
}

//基准距离算法，标准算法，基于球面模型来处理的（立体几何），即Haversine公式
func CalcDistanceByStardard(lat1, log1, lat2, log2 float64) float64 {

	rad := math.Pi / 180.0 //弧度

	lat1 = lat1 * rad
	log1 = log1 * rad
	lat2 = lat2 * rad
	log2 = log2 * rad

	theta := log2 - log1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))

	return float64(dist * EARTH_RADIUS)
}

//简化算法，基于平面几何来处理的（平面几何）。
//原理是当两个点足够近时（比如1km之内），把原来在大尺度上表现为弯曲的球面近似的看做一个二维平面来处理（以此提高效率）
//把两点的距离通过勾股定理来简化计算c^2 = a^2 + b^2
func CalcDistanceByQuick(lat1, log1, lat2, log2 float64) float64 {
	diffLat := (lat1 - lat2) //纬度差的实际距离，单位m
	diffLog := (log1 - log2) //经度差的实际距离，单位m

	//根据勾股定理
	dist := DIST_PER_DEGREE * math.Sqrt(diffLat*diffLat+diffLog*diffLog)
	return dist
}

//用三角函数试试
func CalcDistanceByQuick2(lat1, log1, lat2, log2 float64) float64 {
	diffLat := (lat1 - lat2) //纬度差的实际距离，单位m
	diffLog := (log1 - log2) //经度差的实际距离，单位m
	rad := math.Pi / 180.0   //弧度
	//根据勾股定理
	radian := math.Atan(math.Sin(diffLat*rad) / math.Sin(diffLog*rad)) //由纬度差为对边、经度差为邻边所获得的弧度值
	dist := DIST_PER_DEGREE * math.Sin(diffLat*rad) / math.Sin(radian)
	return dist
}

func TestDisctance() {
	benchmarkPoint := EarthPointPair{31.0, 121.0, 31.0, 121.0} //北纬31度，东经121度
	var arrPoint [TEST_DATA_NUM]EarthPointPair
	var tmpPoint EarthPointPair
	//	var tmp int
	now := time.Now().Unix()
	rand.Seed(now)
	var precision float64 = 0.01
	for i := 0; i < TEST_DATA_NUM; i++ {
		//随机生成1000000组经纬度坐标
		//以1000000为除数求余，然后把得到的余数再与除以1000000作为benchmarkPoint的小数部分
		tmpPoint.lat1 = benchmarkPoint.lat1 + float64(rand.Intn(1000000))/1000000*precision
		tmpPoint.log1 = benchmarkPoint.log1 + float64(rand.Intn(1000000))/1000000*precision
		tmpPoint.lat2 = benchmarkPoint.lat2 + float64(rand.Intn(1000000))/1000000*precision
		tmpPoint.log2 = benchmarkPoint.log2 + float64(rand.Intn(1000000))/1000000*precision
		arrPoint[i] = tmpPoint
	}

	var resultStardard, resultQuick, resultQuick2 [TEST_DATA_NUM]float64
	//计算标准三维距离算法的时间
	before := time.Now().UnixNano()
	for i := 0; i < TEST_DATA_NUM; i++ {
		resultStardard[i] = CalcDistanceByStardard(arrPoint[i].lat1, arrPoint[i].log1, arrPoint[i].lat2, arrPoint[i].log2)
	}
	after := time.Now().UnixNano()
	fmt.Printf("Haversine公式标准球体算法耗时：%E ns\n", float64(after-before))

	//计算简化二维距离算法的时间
	before = time.Now().UnixNano()
	for i := 0; i < TEST_DATA_NUM; i++ {
		resultQuick[i] = CalcDistanceByQuick(arrPoint[i].lat1, arrPoint[i].log1, arrPoint[i].lat2, arrPoint[i].log2)
	}
	after = time.Now().UnixNano()
	fmt.Printf("平面近似二维距离算法（勾股定理）的耗时：%E ns\n", float64(after-before))

	//计算简化二维距离算法的时间,使用三角函数
	before = time.Now().UnixNano()
	for i := 0; i < TEST_DATA_NUM; i++ {
		resultQuick2[i] = CalcDistanceByQuick2(arrPoint[i].lat1, arrPoint[i].log1, arrPoint[i].lat2, arrPoint[i].log2)
	}
	after = time.Now().UnixNano()
	fmt.Printf("平面近似二维距离算法（三角函数）的耗时：%E ns\n", float64(after-before))

	//logFile, err := os.Open("distance.log")	//直接使用open是打开只读的文件
	logFile, err := os.OpenFile("distance.log", os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("打开distance.log失败")
		return
	}
	log.SetOutput(logFile)

	var variance float64 = 0.0
	for i := 0; i < TEST_DATA_NUM; i++ {
		//	log.Printf("第%d组数据: EarthPointPair = %v, stardard = %f, quick = %f, quick2 = %f\n",
		//		i, arrPoint[i], resultStardard[i], resultQuick[i], resultQuick[i])
		variance += math.Pow(resultQuick[i]-resultStardard[i], 2)
	}
	log.Printf("精度为%f度时，搜索范围距离在%f内时, 标准差为%f\n",
		precision, DIST_PER_DEGREE*precision, math.Sqrt(variance/TEST_DATA_NUM))
	return
}
