package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func RandPoint() float64 {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)
	return float64(r) / 1000
}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	return value
}
