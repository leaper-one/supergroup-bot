package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/handlers/user"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type usersImpl struct{}

func registerUsers(router *httptreemux.TreeMux) {
	impl := &usersImpl{}
	router.POST("/auth", impl.authenticate)
	router.POST("/user/chatStatus", impl.chatStatus)
	router.GET("/me", impl.me)
	router.GET("/user/block/:id", impl.blockUser)

	router.GET("/user/list", impl.userList)
	router.GET("/user/adminAndGuest", impl.adminAndGuestList)
	router.GET("/user/search", impl.userSearch)
	router.GET("/user/stat", impl.statClientUser)
	router.PUT("/user/status", impl.updateUserStatus)
	// router.PUT("/user/proxy", impl.updateUserProxy)
	router.PUT("/user/mute", impl.muteClientUser)
	router.PUT("/user/block", impl.blockClientUser)
}

func (impl *usersImpl) authenticate(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Code       string `json:"code"`
		InviteCode string `json:"c"` // 邀请码
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := user.AuthenticateUserByOAuth(r.Context(), r.Header.Get("Origin"), body.Code, body.InviteCode); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, user)
	}
}
func (impl *usersImpl) chatStatus(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		IsReceived   bool `json:"is_received"`
		IsNoticeJoin bool `json:"is_notice_join"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := user.UpdateClientUserChatStatus(r.Context(), middlewares.CurrentUser(r), body.IsReceived, body.IsNoticeJoin); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, user)
	}
}
func (impl *usersImpl) userSearch(w http.ResponseWriter, r *http.Request, params map[string]string) {
	key := r.URL.Query().Get("key")
	if key == "" {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else if users, err := user.GetClientUserByIDOrName(r.Context(), middlewares.CurrentUser(r), key); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, users)
	}
}
func (impl *usersImpl) statClientUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if users, err := user.GetClientUserStat(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, users)
	}
}

func (impl *usersImpl) me(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderDataResponse(w, r, user.GetMe(r.Context(), middlewares.CurrentUser(r)))
}

func (impl *usersImpl) blockUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := common.SuperAddBlockUser(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
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
	} else if err := user.UpdateClientUserStatus(r.Context(), middlewares.CurrentUser(r), body.UserID, body.Status, body.IsCancel); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) updateUserProxy(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var body struct {
		IsProxy  bool   `json:"is_proxy"`
		FullName string `json:"full_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := common.UpdateClientUserProxy(r.Context(), middlewares.CurrentUser(r), body.IsProxy, body.FullName, ""); err != nil {
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
	} else if err := user.BlockUserByID(r.Context(), middlewares.CurrentUser(r), body.UserID, body.IsCancel); err != nil {
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
	} else if err := user.MuteUserByID(r.Context(), middlewares.CurrentUser(r), body.UserID, body.MuteTime); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *usersImpl) userList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	status := r.URL.Query().Get("status")
	if l, err := user.GetClientUserList(r.Context(), middlewares.CurrentUser(r), page, status); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}

func (impl *usersImpl) adminAndGuestList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if l, err := user.GetAdminAndGuestUserList(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, l)
	}
}
