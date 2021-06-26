package main

import (
	"context"
	"flag"
	"github.com/fox-one/mixin-sdk-go"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/services"
)

func main() {
	service := flag.String("service", "http", "run a service")
	flag.Parse()

	database := durable.NewDatabase(context.Background())
	redis := durable.NewRedis(context.Background())
	log.Println(*service)

	mixin.UseApiHost(mixin.ZeromeshApiHost)
	//mixin.UseBlazeHost(mixin.ZkeromeshBlazeHost)

	go func() {
		runtime.SetBlockProfileRate(1) // 开启对阻塞操作的跟踪
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	switch *service {
	case "http":
		err := StartHTTP(database, redis)
		if err != nil {
			log.Println(err)
		}
	default:
		hub := services.NewHub(database, redis)
		err := hub.StartService(*service)
		if err != nil {
			log.Println(err)
		}
	}
}
