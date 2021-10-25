package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/services"
)

func main() {
	service := flag.String("service", "http", "run a service")
	flag.Parse()

	database := durable.NewDatabase(context.Background())
	redis := durable.NewRedis(context.Background())
	log.Println(*service)

	// mixin.UseApiHost(mixin.ZeromeshApiHost)
	//mixin.UseBlazeHost(mixin.ZkeromeshBlazeHost)

	switch *service {
	case "http":
		go func() {
			runtime.SetBlockProfileRate(1) // 开启对阻塞操作的跟踪
			models.StartWithHttpServiceJob()
			_ = http.ListenAndServe("0.0.0.0:6060", nil)
		}()
		err := StartHTTP(database, redis)
		if err != nil {
			log.Println("start http error...", err)
		}
	default:
		hub := services.NewHub(database, redis)
		err := hub.StartService(*service)
		if err != nil {
			log.Println("service error...", err)
		}
	}
}
