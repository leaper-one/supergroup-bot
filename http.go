package main

import (
	"fmt"
	"net/http"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/routes"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
	"github.com/unrolled/render"
	"gorm.io/gorm"
)

func StartHTTP(db *gorm.DB, redis *durable.Redis) error {
	router := httptreemux.New()
	routes.RegisterHandlers(router)
	routes.RegisterRoutes(router)
	handler := middlewares.Authenticate(router)
	handler = middlewares.Constraint(handler)
	handler = middlewares.Context(
		handler,
		db,
		redis,
		render.New(),
		&durable.Logger{},
	)
	handler = handlers.ProxyHeaders(handler)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), handler)
}
