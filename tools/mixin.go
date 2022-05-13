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
		case r = <-timer():
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
	r := make(chan string)
	go func() {
		defer close(r)
		_, err := http.Get(url)
		if err == nil {
			r <- url
		}
	}()
	return r
}

func timer() <-chan string {
	r := make(chan string)
	go func() {
		defer close(r)
		time.Sleep(time.Second * 10)
		r <- ""
	}()
	return r
}
