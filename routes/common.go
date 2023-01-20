package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/live"
	"github.com/MixinNetwork/supergroup/middlewares"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
	"github.com/dimfeld/httptreemux"
)

type commonImpl struct{}

func registerCommon(router *httptreemux.TreeMux) {
	var c commonImpl
	router.POST("/upload", c.uploadFile)
}

func (b *commonImpl) uploadFile(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if url, err := live.UploadLiveImgToMixinStatistics(r.Context(), middlewares.CurrentUser(r), r); err != nil {
		views.RenderErrorResponse(w, r, session.BadDataError(r.Context()))
	} else {
		views.RenderDataResponse(w, r, map[string]string{"view_url": url})
	}
}
