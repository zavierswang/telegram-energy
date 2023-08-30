package utils

import (
	"fmt"
	"telegram-energy/pkg/core/cst"
	"time"
)

func Duration(blocks int64) string {
	var str string
	durationHours := blocks * 3 / 3600
	d := durationHours / 24
	h := durationHours % 24
	if d >= 1 {
		if h > 0 {
			str = fmt.Sprintf("%d天%d小时", d, h)
		} else {
			str = fmt.Sprintf("%d天", d)
		}
	} else {
		str = fmt.Sprintf("%d小时", h)
	}
	return str
}

func DurationSec(days int64) string {
	var str string
	durationHours := days / 3600
	d := durationHours / 24
	h := durationHours % 24
	if d >= 1 {
		if h > 0 {
			str = fmt.Sprintf("%d天%d小时", d, h)
		} else {
			str = fmt.Sprintf("%d天", d)
		}
	} else {
		str = fmt.Sprintf("%d小时", h)
	}
	return str
}

func ReDurationSect(days string) int64 {

	return 0
}

func DateTime(t time.Time) string {
	return t.Format(cst.DateTimeFormatter)
}

func EnergyCount(energy int64) string {
	var count float64
	count = float64(energy) / 32000
	return fmt.Sprintf("%.1f", count)
}
