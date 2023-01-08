package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MixinNetwork/supergroup/handlers/invitation"
	"github.com/MixinNetwork/supergroup/handlers/lottery"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type claimImpl struct{}

func registerClaim(router *httptreemux.TreeMux) {
	var b claimImpl
	router.GET("/claim", b.getClaim)
	router.POST("/claim", b.postClaim)
	router.GET("/power/record", b.getPowerRecord)
	router.POST("/lottery/exchange", b.postClaimExchange)
	router.POST("/lottery", b.PostLottery)
	router.POST("/lottery/reward", b.postLotteryReward)
	router.GET("/lottery/record", b.getLotteryRecord)

	router.GET("/invitation", b.getInvitationCode)
	router.GET("/invitation/record", b.getInvitationRecord)
	router.POST("/voucher", b.postVoucher)

}

func (b *claimImpl) getClaim(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := lottery.GetClaimAndLotteryInitData(r.Context(), middlewares.CurrentUser(r)); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *claimImpl) postClaim(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := lottery.PostClaim(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *claimImpl) getPowerRecord(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	page := r.Form.Get("page")
	pageInt, _ := strconv.Atoi(page)
	if list, err := lottery.GetPowerRecordList(r.Context(), middlewares.CurrentUser(r), pageInt); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, list)
	}
}

func (b *claimImpl) getLotteryRecord(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	page := r.Form.Get("page")
	pageInt, _ := strconv.Atoi(page)
	if list, err := lottery.GetLotteryRecordList(r.Context(), middlewares.CurrentUser(r), pageInt); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, list)
	}
}

func (b *claimImpl) getInvitationCode(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := invitation.GetInviteDataByUserID(r.Context(), middlewares.CurrentUser(r).UserID); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *claimImpl) getInvitationRecord(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	page := r.Form.Get("page")
	pageInt, _ := strconv.Atoi(page)
	if list, err := invitation.GetInvitationListByUserID(r.Context(), middlewares.CurrentUser(r), pageInt); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, list)
	}
}
func (b *claimImpl) postClaimExchange(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := lottery.PostExchangeLottery(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *claimImpl) PostLottery(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if id, err := lottery.PostLottery(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, map[string]string{"lottery_id": id})
	}
}

func (b *claimImpl) postLotteryReward(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		TraceID string `json:"trace_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if client, err := lottery.PostLotteryReward(r.Context(), middlewares.CurrentUser(r), body.TraceID); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else if client != nil {
		views.RenderDataResponse(w, r, client)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *claimImpl) postVoucher(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if status, err := lottery.CheckVoucher(r.Context(), middlewares.CurrentUser(r), body.Code); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, map[string]int{"status": status})
	}
}
