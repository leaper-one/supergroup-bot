package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
)

type Hub struct {
	context  context.Context
	services map[string]Service
}

func NewHub(db *durable.Database, redis *durable.Redis) *Hub {
	hub := &Hub{services: make(map[string]Service)}
	hub.context = session.WithDatabase(context.Background(), db)
	hub.context = session.WithRedis(hub.context, redis)

	hub.registerServices()
	return hub
}

func (hub *Hub) StartService(name string) error {
	service := hub.services[name]
	if service == nil {
		return fmt.Errorf("no service found: %s", name)
	}

	return service.Run(hub.context)
}

func (hub *Hub) registerServices() {
	hub.services["scan"] = &ScanService{}
	hub.services["distribute_message"] = &DistributeMessageService{}
	hub.services["create_message"] = &CreateDistributeMsgService{}
	hub.services["blaze"] = &BlazeService{}
	hub.services["assets_check"] = &AssetsCheckService{}
	hub.services["add_client"] = &AddClientService{}
	hub.services["swap"] = &SwapService{}
	hub.services["update_lp_check"] = &UpdateLpCheckService{}
	hub.services["monitor"] = &MonitorService{}
	hub.services["update_activity"] = &UpdateActivityService{}
	hub.services["u"] = &UpdateClientUserStatusService{}
	hub.services["migration"] = &MigrationService{}
	hub.services["airdrop"] = &AirdropService{}
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
		time.Sleep(time.Second * 30)
		r <- ""
	}()
	return r
}

func UseAutoFasterRoute() {
	for {
		var r string
		select {
		case r = <-useApi(mixin.DefaultApiHost):
		case r = <-useApi(mixin.ZeromeshApiHost):
		case r = <-timer():
		}
		if r == mixin.DefaultApiHost {
			log.Println("use default api...")
			mixin.UseApiHost(mixin.DefaultApiHost)
			mixin.UseBlazeHost(mixin.DefaultBlazeHost)
		} else if r == mixin.ZeromeshApiHost {
			log.Println("use zeromesh api...")
			mixin.UseApiHost(mixin.ZeromeshApiHost)
			mixin.UseBlazeHost(mixin.ZeromeshBlazeHost)
		}
		time.Sleep(time.Minute * 5)
	}
}
