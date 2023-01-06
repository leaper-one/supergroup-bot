package views

import (
	"github.com/MixinNetwork/supergroup/handlers/common"
	"net/http"
)

func RenderUser(w http.ResponseWriter, r *http.Request, user *common.User) {
	RenderDataResponse(w, r, user)
}
