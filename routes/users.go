package routes

import (
	"encoding/json"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
	"net/http"
)

type usersImpl struct{}

func registerUsers(router *httptreemux.TreeMux) {
	impl := &usersImpl{}
	router.POST("/auth", impl.authenticate)
	//router.GET("/me", impl.me)
	router.POST("/user/chatStatus", impl.charStatus)
	router.GET("/me", impl.me)
}

func (impl *usersImpl) authenticate(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := models.AuthenticateUserByOAuth(r.Context(), r.Header.Get("Origin"), body.Code); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderUser(w, r, user)
	}
}
func (impl *usersImpl) charStatus(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		IsReceived   bool `json:"is_received"`
		IsNoticeJoin bool `json:"is_notice_join"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := models.UpdateClientUserChatStatusByHost(r.Context(), middlewares.CurrentUser(r), body.IsReceived, body.IsNoticeJoin); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, user)
	}
}

func (impl *usersImpl) me(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderDataResponse(w, r, middlewares.CurrentUser(r))
}
