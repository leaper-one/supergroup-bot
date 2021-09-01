package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type groupsImpl struct{}

func registerGroups(router *httptreemux.TreeMux) {
	impl := &groupsImpl{}
	router.GET("/group", impl.getGroupInfo)
	router.GET("/group/vip", impl.getGroupVip)
	router.GET("/groupList", impl.getGroupInfoList)
	router.GET("/msgCount", impl.getMsgCount)
	router.GET("/group/status", impl.getGroupStatus)
	router.PUT("/group/setting", impl.updateGroupSetting)
	router.GET("/swapList/:id", impl.swapList)
	router.DELETE("/group", impl.leaveGroup)
	router.GET("/group/stat", impl.groupStat)
}

func (impl *groupsImpl) getGroupInfo(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := models.GetClientInfoByHostOrID(r.Context(), r.Header.Get("Origin"), ""); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) getGroupVip(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := models.GetClientVipAmount(r.Context(), r.Header.Get("Origin")); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) getGroupStatus(w http.ResponseWriter, r *http.Request, params map[string]string) {
	status := models.GetClientStatusByID(r.Context(), middlewares.CurrentUser(r))
	views.RenderDataResponse(w, r, status)
}

func (impl *groupsImpl) getGroupInfoList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := models.GetAllClientInfo(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) getMsgCount(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := models.GetMsgStatistics(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) swapList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	if swapList, err := models.GetSwapList(r.Context(), id); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, swapList)
	}
}

func (impl *groupsImpl) groupStat(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if dailyData, err := models.GetDailyDataByClientID(r.Context(), middlewares.CurrentUser(r)); err != nil {
		log.Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, dailyData)
	}
}

func (impl *groupsImpl) leaveGroup(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := models.LeaveGroup(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (impl *groupsImpl) updateGroupSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
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
