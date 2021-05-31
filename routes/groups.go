package routes

import (
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
	"net/http"
)

type groupsImpl struct{}

func registerGroups(router *httptreemux.TreeMux) {
	impl := &groupsImpl{}
	router.GET("/group", impl.getGroupInfo)
	//router.POST("/group", impl.create)
	//router.PUT("/group/manager/groupStatus", impl.updateGroupStatusSetting)
	//router.DELETE("/group/manager/:groupID/:userID", impl.deleteManager)
	router.GET("/swapList/:id", impl.swapList)
}

//func (impl *groupsImpl) checkGroup(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	conversationID := params["conversationID"]
//	if conversationID == "" {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if checked, err := models.CheckConversationByConversationID(r.Context(), middlewares.CurrentUser(r), conversationID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderConversation(w, r, checked)
//	}
//}

func (impl *groupsImpl) getGroupInfo(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if client, err := models.GetClientInfoByHost(r.Context(), r.Header.Get("Origin")); err != nil {
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

// func (impl *groupsImpl) sharesList(w http.ResponseWriter, r *http.Request, params map[string]string) {
// 	conversationID := params["conversationID"]
// 	if conversationID == "" {
// 		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
// 	} else if checked, err := models.GetSharesCheckingListByConversationID(r.Context(),  conversationID); err != nil {
// 		views.RenderErrorResponse(w, r, err)
// 	} else {
// 		//views.RenderConversation(w, r, checked)
// 		log.Println(checked)
// 	}
// }

//func (impl *groupsImpl) create(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body models.UpdateGroupProps
//	user := middlewares.CurrentUser(r)
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if groupID, err := models.CreateGroup(r.Context(), user, &body); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderCreateGroup(w, r, groupID)
//	}
//}
//
//func (impl *groupsImpl) update(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body models.Group
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if err := models.UpdateGroup(r.Context(), middlewares.CurrentUser(r), &body); err != nil {
//		log.Println(err)
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) updateSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body models.UpdateGroupProps
//	user := middlewares.CurrentUser(r)
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if err := models.UpdateGroupSetting(r.Context(), user, &body); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//type groupStatusSetting struct {
//	GroupID    string `json:"group_id"`
//	StatusName string `json:"status_name"`
//	Status     string `json:"status"`
//}
//
//func (impl *groupsImpl) updateGroupStatusSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body groupStatusSetting
//	user := middlewares.CurrentUser(r)
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if err := models.UpdateGroupStatusSetting(r.Context(), user, body.GroupID, body.StatusName, body.Status); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) updateGroupInviteSetting(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body models.GroupInviteSetting
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if err := models.UpdateGroupInviteSetting(r.Context(), middlewares.CurrentUser(r), body); err != nil {
//		log.Println(err)
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) managerList(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupID := params["groupID"]
//	if groupID == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if users, err := models.GetManagerListByGroupID(r.Context(), middlewares.CurrentUser(r), groupID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderUserList(w, r, users)
//	}
//}
//func (impl *groupsImpl) updateManager(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupID := params["groupID"]
//	var body struct {
//		Users []string `json:"users"`
//	}
//	user := middlewares.CurrentUser(r)
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
//	} else if err := models.AddManager(r.Context(), user, groupID, body.Users); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) deleteManager(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupID := params["groupID"]
//	userID := params["userID"]
//	if groupID == "" || userID == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if err := models.RemoveManager(r.Context(), middlewares.CurrentUser(r), groupID, userID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) groupAssetSnapshots(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupID := params["groupID"]
//	assetID := params["assetID"]
//	if groupID == "" || assetID == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if snapshots, err := models.GetGroupSnapshotsByGroupIDAndAssetID(r.Context(), middlewares.CurrentUser(r), groupID, assetID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderGroupAssetSnapshots(w, r, snapshots)
//	}
//}
//
//type withdrawalParams struct {
//	GroupID string `json:"group_id"`
//	AssetID string `json:"asset_id"`
//	Amount  string `json:"amount"`
//}
//
//func (impl *groupsImpl) groupAssetWithdrawal(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	var body withdrawalParams
//	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if body.GroupID == "" || body.AssetID == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if err := models.WithdrawalGroupAssets(r.Context(), middlewares.CurrentUser(r), body.GroupID, body.AssetID, body.Amount); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderBoolResponse(w, r, true)
//	}
//}
//
//func (impl *groupsImpl) groupInfo(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupNumber := params["groupNumber"]
//	if groupNumber == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if groupInfo, err := models.GetGroupInfoByNumber(r.Context(), middlewares.CurrentUser(r), groupNumber); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderGroupInfo(w, r, groupInfo)
//	}
//}
//
//func (impl *groupsImpl) checkCanJoin(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupNumber := params["groupID"]
//	err := r.ParseForm()
//	if err != nil {
//		session.Logger(r.Context()).Println(err)
//	}
//	conversationID := r.Form.Get("conversation_id")
//	if groupNumber == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if canJoin, err := models.CheckUserCanJoinGroup(r.Context(), middlewares.CurrentUser(r), groupNumber, conversationID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderCanJoin(w, r, canJoin)
//	}
//}
//
//func (impl *groupsImpl) groupStat(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	groupID := params["groupID"]
//	if groupID == "" {
//		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
//	} else if stat, err := models.GetStatisticsByConversationID(r.Context(), middlewares.CurrentUser(r), groupID); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderGroupStat(w, r, stat)
//	}
//}
//
//
//func (impl *groupsImpl) groupList(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	if groupList, err := models.GetGroupList(r.Context()); err != nil {
//		views.RenderErrorResponse(w, r, err)
//	} else {
//		views.RenderGroupList(w, r, groupList)
//	}
//}
