package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/liquidity"
	"github.com/MixinNetwork/supergroup/handlers/mint"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type mintImpl struct{}

func registerMint(router *httptreemux.TreeMux) {
	var c mintImpl
	router.GET("/mint/:id", c.getMintByID)
	router.GET("/mint/record", c.getMintRecordByID)
	router.POST("/mint", c.postMintByMintIDAndTraceID)

	router.GET("/liquidity/:id", c.getLiquidityByID)
	router.GET("/liquidity/record", c.getLiquidityRecordByID)
	router.POST("/liquidity/join", c.postLiquidity)
}

func (b *mintImpl) getMintByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := mint.GetLiquidityMiningRespByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *mintImpl) getMintRecordByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	mintID := r.Form.Get("mint_id")
	page := r.Form.Get("page")
	status := r.Form.Get("status")
	if data, err := mint.GetLiquidityMiningRecordByMiningIDAndUserID(r.Context(), middlewares.CurrentUser(r), mintID, page, status); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *mintImpl) getLiquidityRecordByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	id := r.Form.Get("id")
	if data, err := liquidity.GetLiquiditySnapshots(r.Context(), middlewares.CurrentUser(r), id); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *mintImpl) postMintByMintIDAndTraceID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		RecordID string `json:"record_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := mint.ReceivedLiquidityMiningTx(r.Context(), middlewares.CurrentUser(r), body.RecordID); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *mintImpl) getLiquidityByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := liquidity.GetLiquidityInfo(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *mintImpl) postLiquidity(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if res, err := liquidity.PostLiquidity(r.Context(), middlewares.CurrentUser(r), body.ID); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, res)
	}
}
