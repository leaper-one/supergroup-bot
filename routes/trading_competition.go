package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type tradingCompetition struct{}

func registerTradingCompetetion(router *httptreemux.TreeMux) {
	var c tradingCompetition
	router.GET("/trading_competetion/:id", c.getTradingCompetetion)
	router.GET("/trading_competetion/:id/rank", c.getTradingCompetetionRank)
}

func (b *tradingCompetition) getTradingCompetetion(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := models.GetTradingCompetetionByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *tradingCompetition) getTradingCompetetionRank(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, err)
		return
	}
	if data, err := models.GetRandingCompetetionRankByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}
