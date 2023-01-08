package common

import (
	"net"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jreisinger/checkip/check"
)

func createLoginLog(u *models.ClientUser, ip, ua string) {
	if strings.HasPrefix(ip, "192.168.") {
		return
	}
	addr, err := checkIp(ip)
	if err != nil {
		tools.Println(err)
	}
	if err := session.DB(models.Ctx).Save(&models.LoginLog{
		UserID:    u.UserID,
		ClientID:  u.ClientID,
		Addr:      ip,
		UA:        ua,
		IpAddr:    addr,
		UpdatedAt: time.Now(),
	}); err != nil {
		tools.Println(err)
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
