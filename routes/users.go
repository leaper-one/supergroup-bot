package routes

import (
	"encoding/json"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
	"log"
	"net/http"
	"strconv"
)

type usersImpl struct{}

func registerUsers(router *httptreemux.TreeMux) {
	impl := &usersImpl{}
	router.POST("/auth", impl.authenticate)
	//router.GET("/me", impl.me)
	router.POST("/user/chatStatus", impl.chatStatus)
	router.GET("/me", impl.me)
	router.GET("/user/block/:id", impl.blockUser)

	router.GET("/user/list", impl.userList)
	//router.GET("/user/:key/search", impl.userSearch)
	router.GET("/user/:id/guest", impl.guestUser)
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
func (impl *usersImpl) chatStatus(w http.ResponseWriter, r *http.Request, params map[string]string) {
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

func (impl *usersImpl) blockUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}

	key := r.Form.Get("key")
	if key != "zlnb" {
		views.RenderErrorResponse(w, r, session.ForbiddenError(r.Context()))
		return
	}

	if err := models.AddBlockUser(r.Context(), params["id"]); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "ok")
	}
}

func (impl *usersImpl) guestUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["id"] == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	// TODO
	if l, err := models.StatLive(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}

func (impl *usersImpl) userList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if l, err := models.GetClientUserList(r.Context(), middlewares.CurrentUser(r), page); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}
