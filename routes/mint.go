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

type mintImpl struct{}

func registerMint(router *httptreemux.TreeMux) {
	var c mintImpl
	router.GET("/mint/:id", c.getMintByID)
	router.GET("/mint/record", c.getMintRecordByID)
	router.POST("/mint", c.postMintByMintIDAndTraceID)
}

func (b *mintImpl) getMintByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := models.GetLiquidityMiningRespByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
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
	if data, err := models.GetLiquidtityMiningRecordByMiningIDAndUserID(r.Context(), middlewares.CurrentUser(r), mintID, page, status); err != nil {
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
	} else if err := models.ReceivedLiquidityMiningTx(r.Context(), middlewares.CurrentUser(r), body.RecordID); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}
