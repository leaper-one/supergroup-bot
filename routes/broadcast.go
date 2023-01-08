package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/broadcast"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type broadcastImpl struct{}

func registerBroadcast(router *httptreemux.TreeMux) {
	var b broadcastImpl
	router.GET("/broadcast", b.getBroadcast)
	router.POST("/broadcast", b.postBroadcast)
	router.DELETE("/broadcast/:id", b.deleteBroadcast)
}

func (b *broadcastImpl) getBroadcast(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if broadcasts, err := broadcast.GetBroadcast(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, broadcasts)
	}
}

func (b *broadcastImpl) postBroadcast(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Data     string `json:"data"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := broadcast.CreateBroadcast(r.Context(), middlewares.CurrentUser(r), body.Data, body.Category); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *broadcastImpl) deleteBroadcast(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if err := broadcast.DeleteBroadcast(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}
