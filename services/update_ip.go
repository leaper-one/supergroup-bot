package services

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/jreisinger/checkip/check"
	"github.com/panjf2000/ants/v2"
)

type UpdateIpAddrService struct{}

func (service *UpdateIpAddrService) Run(ctx context.Context) error {
	session.Database(ctx).Exec(ctx, `
ALTER TABLE login_log ADD COLUMN ip_addr VARCHAR NOT NULL DEFAULT '';
	`)
	distributeAntsPool, _ = ants.NewPool(10, ants.WithPreAlloc(true))
	us := make([]*models.LoginLog, 0)
	session.Database(ctx).ConnQuery(ctx, `
	SELECT user_id, client_id, addr FROM login_log
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u models.LoginLog
			if err := rows.Scan(&u.UserID, &u.ClientID, &u.Addr); err != nil {
				return err
			}
			us = append(us, &u)
		}
		return nil
	})
	log.Println("共计...", len(us))
	var wg sync.WaitGroup
	for i, u := range us {
		wg.Add(1)
		distributeAntsPool.Submit(func() {
			defer wg.Done()
			ip, err := checkIp(u.Addr)
			log.Println(i, "/", len(us))
			if err != nil {
				log.Println(err)
				return
			}
			if ip == "" {
				return
			}
			_, err = session.Database(ctx).Exec(ctx, `
		UPDATE login_log SET ip_addr = $1 WHERE user_id = $2 AND client_id = $3
		`, ip, u.UserID, u.ClientID)
			if err != nil {
				log.Println(err)
			}
		})
	}
	wg.Wait()
	return nil
}

func checkIp(ip string) (string, error) {
	ipaddr := net.ParseIP(ip)
	r, err := check.DBip(ipaddr)
	log.Println(ip, r.Info.Summary())
	return r.Info.Summary(), err
}
