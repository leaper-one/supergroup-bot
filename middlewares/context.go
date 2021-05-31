package middlewares

import (
	"net/http"

	"github.com/fox-one/mixin-sdk-go"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/unrolled/render"
)

func Context(handler http.Handler, db *durable.Database, render *render.Render, logger *durable.Logger, client *mixin.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := session.WithRequest(r.Context(), r)
		ctx = session.WithDatabase(ctx, db)
		ctx = session.WithRender(ctx, render)
		ctx = session.WithLogger(ctx, logger)
		ctx = session.WithMixinClient(ctx, client)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
