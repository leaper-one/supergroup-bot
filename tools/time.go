package tools

import (
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
