package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type airdropImpl struct{}

func registerAirdrop(router *httptreemux.TreeMux) {
	var b airdropImpl
	router.GET("/airdrop/:airdropID", b.getAirdrop)
	router.POST("/airdrop/:airdropID", b.postAirdrop)
}

func (b *airdropImpl) getAirdrop(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if airdrops, err := models.GetAirdrop(r.Context(), middlewares.CurrentUser(r), params["airdropID"]); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, airdrops)
	}
}

func (b *airdropImpl) postAirdrop(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if status, err := models.ClaimAirdrop(r.Context(), middlewares.CurrentUser(r), params["airdropID"]); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, status)
	}
}
