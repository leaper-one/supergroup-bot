package services

import (
	"context"
	"fmt"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"gorm.io/gorm"
)

type Hub struct {
	context  context.Context
	services map[string]Service
}

func NewHub(db *gorm.DB, redis *durable.Redis) *Hub {
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
	hub.services["add_client"] = &AddClientService{}
	hub.services["swap"] = &SwapService{}
	hub.services["migration"] = &MigrationService{}
	hub.services["airdrop"] = &AirdropService{}
	hub.services["add_voucher"] = &VoucherService{}
	hub.services["reset_live"] = &ResetLiveService{}
}
