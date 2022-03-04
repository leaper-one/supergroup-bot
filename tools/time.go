package tools

import (
	"log"
	"sync"
	"time"
)

func GetZeroTime(date time.Time) time.Time {
	timeStr := date.Format("2006-1-2")
	t, _ := time.Parse("2006-1-2", timeStr)
	return t
}

func Debounce(interval time.Duration) func(f func()) {
	var l sync.Mutex
	var timer *time.Timer

	return func(f func()) {
		l.Lock()
		defer l.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(interval, f)
	}
}

func PrintTimeDuration(info string, start time.Time) {
	log.Printf("%s 耗时为: %f ms", info, time.Since(start).Seconds()*1000)
}

func GetMinuteTime(date time.Time) string {
	timeStr := date.Format("2006-1-2-15:04")
	return timeStr
}

func GetNextMinuteTime(s string) int64 {
	t, err := time.Parse("2006-1-2-15:04", s)
	if err != nil {
		return int64(time.Second)
	}
	return t.Add(time.Minute).UnixNano() - time.Now().UnixNano()
}
