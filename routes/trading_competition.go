package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type tradingCompetition struct{}

func registerTradingCompetition(router *httptreemux.TreeMux) {
	var c tradingCompetition
	router.GET("/trading_competition/:id", c.getTradingCompetition)
	router.GET("/trading_competition/:id/rank", c.getTradingCompetitionRank)
}

func (b *tradingCompetition) getTradingCompetition(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := models.GetTradingCompetitionByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *tradingCompetition) getTradingCompetitionRank(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, err)
		return
	}
	if data, err := models.GetTradingCompetitionRankByID(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}
