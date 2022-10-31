package tools

import (
	"net/http"
	"time"

	"github.com/fox-one/mixin-sdk-go"
)

func UseAutoFasterRoute() {
	for {
		var r string
		select {
		case r = <-useApi(mixin.DefaultApiHost):
		case r = <-useApi(mixin.ZeromeshApiHost):
		case <-time.After(time.Second * 10):
			continue
		}
		if r == mixin.DefaultApiHost {
			mixin.UseApiHost(mixin.DefaultApiHost)
			mixin.UseBlazeHost(mixin.DefaultBlazeHost)
		} else if r == mixin.ZeromeshApiHost {
			mixin.UseApiHost(mixin.ZeromeshApiHost)
			mixin.UseBlazeHost(mixin.ZeromeshBlazeHost)
		}
		time.Sleep(time.Second * 10)
	}
}

func useApi(url string) <-chan string {
	r := make(chan string, 1)
	go func() {
		defer close(r)
		resp, err := http.Get(url)
		if err == nil {
			defer resp.Body.Close()
			r <- url
		}
	}()
	return r
}
