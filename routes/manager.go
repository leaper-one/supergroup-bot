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

type managerImpl struct{}

func registerManager(router *httptreemux.TreeMux) {
	impl := &managerImpl{}
	router.PUT("/group/setting", impl.updateGroupSetting)
	router.GET("/group/stat", impl.groupStat)

	router.GET("/group/advance/setting", impl.groupSetting)
	router.PUT("/group/advance/setting", impl.updateAdvanceGroupSetting)

	router.GET("/group/member/auth", impl.getGroupMemberAuth)
	router.PUT("/group/member/auth", impl.updateGroupMemberAuth)
}

func (impl *managerImpl) groupStat(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if dailyData, err := models.GetDailyDataByClientID(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, dailyData)
	}
}

func (impl *managerImpl) groupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if setting, err := models.GetClientAdvanceSetting(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, setting)
	}
}

func (impl *managerImpl) updateAdvanceGroupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body models.ClientAdvanceSetting
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.UpdateClientAdvanceSetting(r.Context(), middlewares.CurrentUser(r), body); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *managerImpl) getGroupMemberAuth(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if setting, err := models.GetClientMemberAuth(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, setting)
	}
}

func (impl *managerImpl) updateGroupMemberAuth(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body models.ClientMemberAuth
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.UpdateClientMemberAuth(r.Context(), middlewares.CurrentUser(r), body); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *managerImpl) updateGroupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Description string `json:"description,omitempty"`
		Welcome     string `json:"welcome,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.UpdateClientSetting(r.Context(), middlewares.CurrentUser(r), body.Description, body.Welcome); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}
