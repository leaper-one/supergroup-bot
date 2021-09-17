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

type guessImpl struct{}

func registerGuess(router *httptreemux.TreeMux) {
	var b guessImpl
	router.GET("/guess/:id", b.getGuess)
	router.POST("/guess", b.postGuess)
	router.GET("/guess/record", b.getGuessRecord)
}

func (b *guessImpl) getGuess(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if data, err := models.GetGuessPageInitData(r.Context(), middlewares.CurrentUser(r), params["id"]); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, data)
	}
}

func (b *guessImpl) postGuess(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		GuessID   string `json:"guess_id"`
		GuessType int    `json:"guess_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if err := models.PostGuess(r.Context(), middlewares.CurrentUser(r), body.GuessID, body.GuessType); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, "success")
	}
}

func (b *guessImpl) getGuessRecord(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := r.ParseForm(); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
		return
	}
	id := r.Form.Get("id")
	if list, err := models.GetGuessRecordListByUserID(r.Context(), middlewares.CurrentUser(r), id); err != nil {
		session.Logger(r.Context()).Println(err)
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, list)
	}
}
