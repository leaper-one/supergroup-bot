package models

import (
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
)

const login_log_DDL = `
CREATE TABLE IF NOT EXISTS login_log (
	user_id      VARCHAR(36) NOT NULL,
	client_id    VARCHAR(36) NOT NULL,
	addr         VARCHAR(255) NOT NULL,
	ua           VARCHAR(255) NOT NULL,
	updated_at   TIMESTAMP NOT NULL DEFAULT now(),
	PRIMARY KEY (user_id, client_id)
);
`

func createLoginLog(u *ClientUser, addr, ua string) {
	query := durable.InsertQueryOrUpdate("login_log", "user_id,client_id", "addr,ua,updated_at")
	_, err := session.Database(_ctx).Exec(_ctx, query, u.UserID, u.ClientID, addr, ua, time.Now())
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
}
