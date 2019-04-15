package orcprometheus

import "math"

var DefTimeBuckets = ExponentialBuckets(0.001, math.Sqrt(2), 32)

var DefBytesToKilobytesBuckets = ExponentialBuckets(1, math.Sqrt(2), 50)

func ExponentialBuckets(start, factor float64, n int) []float64 {
	var rv []float64
	x := start
	for i := 0; i < n; i++ {
		rv = append(rv, x)
		x *= factor
	}
	return rv
}
