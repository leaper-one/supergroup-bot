package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type liveImpl struct{}

func registerLive(router *httptreemux.TreeMux) {
	var b liveImpl
	router.GET("/live", b.getLiveList)
	router.POST("/live", b.postLive)
	router.GET("/live/:id/start", b.startLive)
	router.GET("/live/:id", b.liveInfo)
	router.GET("/live/:id/stop", b.stopLive)
	router.GET("/live/:id/stat", b.statLive)

	router.GET("/news/:id/top", b.topNews)
	router.GET("/news/:id/cancelTop", b.cancelTopNews)
	router.GET("/live/:id/replay", b.getReplayList)

}

func (b *liveImpl) getLiveList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if lives, err := models.GetLivesByClientID(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, lives)
	}
}

func (b *liveImpl) postLive(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body models.Live
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.UpdateLive(r.Context(), middlewares.CurrentUser(r), &body); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *liveImpl) startLive(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if err := models.StartLive(r.Context(), middlewares.CurrentUser(r), params["id"], r.URL.Query().Get("url")); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *liveImpl) liveInfo(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if l, err := models.GetLiveByID(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}

func (b *liveImpl) stopLive(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if err := models.StopLive(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *liveImpl) statLive(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if l, err := models.StatLive(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}

func (b *liveImpl) topNews(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if err := models.TopNews(r.Context(), middlewares.CurrentUser(r), params["id"], false); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *liveImpl) cancelTopNews(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if err := models.TopNews(r.Context(), middlewares.CurrentUser(r), params["id"], true); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *liveImpl) getReplayList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	if lrs, err := models.GetLiveReplayByLiveID(r.Context(), middlewares.CurrentUser(r), params["id"], r.RemoteAddr); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, lrs)
	}
}
