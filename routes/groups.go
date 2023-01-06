package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type groupsImpl struct{}

func registerGroups(router *httptreemux.TreeMux) {
	impl := &groupsImpl{}
	router.GET("/group", impl.getGroupInfo)
	router.GET("/group/vip", impl.getGroupVip)
	router.GET("/groupList", impl.getGroupInfoList)
	router.GET("/group/status", impl.getGroupStatus)
	router.GET("/swapList/:id", impl.swapList)
	router.DELETE("/group", impl.leaveGroup)
}

func (impl *groupsImpl) getGroupInfo(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := common.GetClientInfoByHostOrID(r.Context(), r.Header.Get("Origin")); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) getGroupVip(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := common.GetClientVipAmount(r.Context(), r.Header.Get("Origin")); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) getGroupStatus(w http.ResponseWriter, r *http.Request, params map[string]string) {
	status := common.GetClientStatusByID(r.Context(), middlewares.CurrentUser(r))
	views.RenderDataResponse(w, r, status)
}

func (impl *groupsImpl) getGroupInfoList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := common.GetAllConfigClientInfo(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, client)
	}
}

func (impl *groupsImpl) swapList(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	if swapList, err := common.GetSwapList(r.Context(), id); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, swapList)
	}
}
func (impl *groupsImpl) leaveGroup(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := common.LeaveGroup(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}
