package models

import (
	"net"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jreisinger/checkip/check"
)

const login_log_DDL = `
CREATE TABLE IF NOT EXISTS login_log (
	user_id      VARCHAR(36) NOT NULL,
	client_id    VARCHAR(36) NOT NULL,
	addr         VARCHAR(255) NOT NULL,
	ua           VARCHAR(255) NOT NULL,
	ip_addr      VARCHAR NOT NULL DEFAULT '',
	updated_at   TIMESTAMP NOT NULL DEFAULT now(),
	PRIMARY KEY (user_id, client_id)
);
`

type LoginLog struct {
	UserID   string
	ClientID string
	Addr     string
	UA       string
	IpAddr   string
}

func createLoginLog(u *ClientUser, ip, ua string) {
	addr, err := checkIp(ip)
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
	query := durable.InsertQueryOrUpdate("login_log", "user_id,client_id", "addr,ua,ip_addr,updated_at")
	_, err = session.Database(_ctx).Exec(_ctx, query, u.UserID, u.ClientID, ip, ua, addr, time.Now())
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
}

func checkIp(ip string) (string, error) {
	ipaddr := net.ParseIP(ip)
	r, err := check.DBip(ipaddr)
	if err != nil {
		return "", err
	}
	return r.Info.Summary(), nil
}
