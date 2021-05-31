package views

import (
	"github.com/MixinNetwork/supergroup/models"
	"net/http"
)

func RenderUser(w http.ResponseWriter, r *http.Request, user *models.User) {
	RenderDataResponse(w, r, user)
}

