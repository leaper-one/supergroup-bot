package routes

import (
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
	"net/http"
)

type commonImpl struct{}

func registerCommon(router *httptreemux.TreeMux) {
	var c commonImpl
	router.POST("/upload", c.uploadFile)
}

func (b *commonImpl) uploadFile(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if url, err := models.UploadLiveImgToMixinStatistics(r.Context(), middlewares.CurrentUser(r), r); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, map[string]string{"view_url": url})
	}
}
