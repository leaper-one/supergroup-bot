package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
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
	router.GET("/user/search", impl.userSearch)
	router.GET("/user/stat", impl.statClientUser)
	router.PUT("/user/status", impl.updateUserStatus)
	router.PUT("/user/mute", impl.muteClientUser)
	router.PUT("/user/block", impl.blockClientUser)
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
func (impl *usersImpl) userSearch(w http.ResponseWriter, r *http.Request, params map[string]string) {
	key := r.URL.Query().Get("key")
	if key == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else if users, err := models.GetClientUserByIDOrName(r.Context(), middlewares.CurrentUser(r), key); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, users)
	}
}
func (impl *usersImpl) statClientUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if users, err := models.GetClientUserStat(r.Context(), middlewares.CurrentUser(r)); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, users)
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
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) updateUserStatus(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var body struct {
		UserID   string `json:"user_id"`
		Status   int    `json:"status"`
		IsCancel bool   `json:"is_cancel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.UpdateClientUserStatus(r.Context(), middlewares.CurrentUser(r), body.UserID, body.Status, body.IsCancel); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) blockClientUser(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var body struct {
		UserID   string `json:"user_id"`
		IsCancel bool   `json:"is_cancel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.BlockUserByID(r.Context(), middlewares.CurrentUser(r), body.UserID, body.IsCancel); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) muteClientUser(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var body struct {
		UserID   string `json:"user_id"`
		MuteTime string `json:"mute_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.MuteUserByID(r.Context(), middlewares.CurrentUser(r), body.UserID, body.MuteTime); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) userList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	status := r.URL.Query().Get("status")
	if l, err := models.GetClientUserList(r.Context(), middlewares.CurrentUser(r), page, status); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}
