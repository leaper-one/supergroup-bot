package user

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jreisinger/checkip/check"
)

type UserMeResp struct {
	*models.ClientUser
	FullName string `json:"full_name"`
	IsClaim  bool   `json:"is_claim"`
	IsBlock  bool   `json:"is_block"`
	IsProxy  bool   `json:"is_proxy"`
}

func GetMe(ctx context.Context, u *models.ClientUser) UserMeResp {
	req := session.Request(ctx)
	go createLoginLog(u, req.RemoteAddr, req.Header.Get("User-Agent"))
	proxy, _ := common.GetClientUserProxyByProxyID(ctx, u.ClientID, u.UserID)
	me := UserMeResp{
		ClientUser: u,
		IsClaim:    common.CheckIsClaim(ctx, u.UserID),
		IsBlock:    u.Status == models.ClientUserStatusBlock,
		IsProxy:    proxy.Status == models.ClientUserProxyStatusActive,
		FullName:   proxy.FullName,
	}
	return me
}

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
	}).Error; err != nil {
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
