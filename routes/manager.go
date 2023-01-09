package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/clients"
	"github.com/MixinNetwork/supergroup/handlers/statistic"
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
	if dailyData, err := statistic.GetDailyDataByClientID(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, dailyData)
	}
}

func (impl *managerImpl) groupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if setting, err := clients.GetClientAdvanceSetting(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, setting)
	}
}

func (impl *managerImpl) updateAdvanceGroupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body clients.ClientAdvanceSetting
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := clients.UpdateClientAdvanceSetting(r.Context(), middlewares.CurrentUser(r), body); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *managerImpl) getGroupMemberAuth(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if setting, err := clients.GetClientMemberAuth(r.Context(), middlewares.CurrentUser(r).ClientID); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, setting)
	}
}

func (impl *managerImpl) updateGroupMemberAuth(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body models.ClientMemberAuth
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := clients.UpdateClientMemberAuth(r.Context(), middlewares.CurrentUser(r), body); err != nil {
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
	} else if err := clients.UpdateClientSetting(r.Context(), middlewares.CurrentUser(r), body.Description, body.Welcome); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}
