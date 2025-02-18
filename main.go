package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/jobs"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/services"
	"github.com/MixinNetwork/supergroup/tools"
)

func main() {
	service := flag.String("service", "http", "run a service")
	flag.Parse()

	database := durable.NewDatabase(context.Background())
	redis := durable.NewRedis(context.Background())
	log.Println(*service)

	go func() {
		if config.Config.Pprof != nil && config.Config.Pprof[*service] != "" {
			runtime.SetBlockProfileRate(1)
			_ = http.ListenAndServe(config.Config.Pprof[*service], nil)
		}
	}()
	switch *service {
	case "auto_migrate":
		models.AutoMigrate()
	case "http":
		go tools.UseAutoFasterRoute()
		go jobs.StartWithHttpServiceJob()
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
