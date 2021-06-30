package routes

import (
	"net/http"
	"runtime"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

func RegisterRoutes(router *httptreemux.TreeMux) {
	router.GET("/", root)
	router.GET("/_hc", healthCheck)
	registerUsers(router)
	registerGroups(router)
	registerBroadcast(router)
	registerCommon(router)
	registerLive(router)
}

func root(w http.ResponseWriter, r *http.Request, params map[string]string) {
	views.RenderDataResponse(w, r, map[string]string{
		"build":      config.BuildVersion + "-" + runtime.Version(),
		"developers": "https://snapshot-api.mixin.one",
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request, params map[string]string) {
	views.RenderBlankResponse(w, r)
}
