package services

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
)

type UpdateClientUserStatusService struct{}

type updateUserStatusProps struct {
	ClientID string   `json:"client_id,omitempty"`
	Status   int      `json:"status,omitempty"`
	Users    []string `json:"users,omitempty"`
	Cancel   bool     `json:"cancel,omitempty"`
}

func (service *UpdateClientUserStatusService) Run(ctx context.Context) error {
	var err error
	data, err := ioutil.ReadFile("user.json")
	if err != nil {
		log.Println("user.json open fail...")
		return err
	}
	var u updateUserStatusProps
	err = json.Unmarshal(data, &u)
	if err != nil {
		return err
	}
	tools.PrintJson(u)

	if u.Status == models.ClientUserStatusAdmin {
		// 新增管理员
		for _, uid := range u.Users {
			if err := updateManagerList(ctx, u.ClientID, uid); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}
	if u.Status == models.ClientUserStatusGuest {
		// 新增嘉宾
		if u.Cancel {
			// 取消所有嘉宾
			if _, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET status=$3 WHERE status=$2 AND client_id=$1`, u.ClientID, models.ClientUserStatusGuest, models.ClientUserStatusLarge); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
		for _, uid := range u.Users {
			var _u string
			log.Println(uid)
			if err := session.Database(ctx).QueryRow(ctx, `
SELECT cu.user_id FROM client_users cu
LEFT JOIN users u on cu.user_id=u.user_id
WHERE cu.client_id=$1 AND u.identity_number=$2
`, u.ClientID, uid).Scan(&_u); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					user, err := models.GetMixinClientByID(ctx, u.ClientID).ReadUser(ctx, uid)
					if err != nil {
						session.Logger(ctx).Println(err)
					}
					models.WriteUser(ctx, &models.User{
						UserID:         user.UserID,
						IdentityNumber: user.IdentityNumber,
						FullName:       user.FullName,
						AvatarURL:      user.AvatarURL,
					})
					log.Println(user.UserID, "...")
					if _, err := session.Database(ctx).Exec(ctx, `INSERT INTO client_users(client_id,user_id,priority,status) VALUES($1,$2,$3,$4)`, u.ClientID, user.UserID, models.ClientUserPriorityHigh, models.ClientUserStatusGuest); err != nil {
						session.Logger(ctx).Println(err)
					}
					continue
				}
			}
			log.Println(_u, err)
			// 有了
			if _, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3,status=$4 WHERE client_id=$1 AND user_id=$2`, u.ClientID, _u, models.ClientUserPriorityHigh, models.ClientUserStatusGuest); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
		log.Println(3)
	}
	return nil
}
