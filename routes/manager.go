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
	// TODO...
	router.GET("/group/advance/setting", impl.groupSetting)
	router.PUT("/group/advance/setting", impl.updateAdvanceGroupSetting)
}

func (impl *managerImpl) groupStat(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if dailyData, err := models.GetDailyDataByClientID(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, dailyData)
	}
}

// TODO...
func (impl *managerImpl) groupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if setting, err := models.GetClientSetting(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, setting)
	}
}

// TODO...
func (impl *managerImpl) updateAdvanceGroupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
	// var body struct {
	// 	Mute string `json:"mute,omitempty"`
	// 	NewMemberNotic     string `json:"new_member_notice,omitempty"`
	// }

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
