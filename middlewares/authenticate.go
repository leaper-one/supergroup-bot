package middlewares

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/views"
)

var whitelist = [][2]string{
	{"POST", "^/auth$"},
	{"GET", "^/_hc$"},
	{"GET", "^/group$"},
	{"GET", "^/groupList$"},
	{"GET", "^/msgCount$"},
}

type contextValueKey struct{ int }

var keyCurrentUser = contextValueKey{1000}

func CurrentUser(r *http.Request) *models.ClientUser {
	user, _ := r.Context().Value(keyCurrentUser).(*models.ClientUser)
	return user
}

func Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			handleUnauthorized(handler, w, r)
			return
		}
		user, err := common.AuthenticateUserByToken(r.Context(), r.Header.Get("Origin"), auth[7:])
		if durable.CheckNotEmptyError(err) != nil {
			views.RenderErrorResponse(w, r, err)
		} else if user == nil {
			handleUnauthorized(handler, w, r)
		} else {
			ctx := context.WithValue(r.Context(), keyCurrentUser, user)
			handler.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func handleUnauthorized(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	for _, pp := range whitelist {
		if pp[0] != r.Method {
			continue
		}
		if matched, err := regexp.MatchString(pp[1], r.URL.Path); err == nil && matched {
			handler.ServeHTTP(w, r)
			return
		}
	}

	views.RenderErrorResponse(w, r, session.AuthorizationError(r.Context()))
}
