package models

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

const activity_DDL = `
CREATE TABLE IF NOT EXISTS activity (
    activity_index      SMALLINT NOT NULL PRIMARY KEY,
    client_id           VARCHAR(36) NOT NULL,
    status              SMALLINT DEFAULT 1, -- 1 不展示 2 展示
    img_url             VARCHAR(512) DEFAULT '',
    expire_img_url      VARCHAR(512) DEFAULT '',
    action              VARCHAR(512) DEFAULT '',
    start_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    expire_at           TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type Activity struct {
	ActivityIndex int    `json:"activity_index,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	Status        int    `json:"status,omitempty"`
	ImgURL        string `json:"img_url,omitempty"`
	ExpireImgURL  string `json:"expire_img_url,omitempty"`
	Action        string `json:"action,omitempty"`

	StartAt   time.Time `json:"start_at,omitempty"`
	ExpireAt  time.Time `json:"expire_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func GetActivityByClientID(ctx context.Context, clientID string) ([]*Activity, error) {
	as := make([]*Activity, 0)
	asString, err := session.Redis(ctx).Get(ctx, "activity:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = session.Database(ctx).ConnQuery(ctx, `
SELECT activity_index,img_url,expire_img_url,action,start_at,expire_at,created_at FROM activity WHERE client_id=$1 AND status=2 ORDER BY activity_index
`, func(rows pgx.Rows) error {
				for rows.Next() {
					var a Activity
					if err := rows.Scan(&a.ActivityIndex, &a.ImgURL, &a.ExpireImgURL, &a.Action, &a.StartAt, &a.ExpireAt, &a.CreatedAt); err != nil {
						session.Logger(ctx).Println(err)
					}
					as = append(as, &a)
				}
				return nil
			}, clientID)
			asByte, _ := json.Marshal(as)
			if err := session.Redis(ctx).Set(ctx, "activity:"+clientID, string(asByte), time.Hour*12).Err(); err != nil {
				session.Logger(ctx).Println(err)
			}
		} else {
			session.Logger(ctx).Println(err)
		}
	} else {
		err = json.Unmarshal([]byte(asString), &as)
	}

	return as, err
}

func UpdateActivity(ctx context.Context, a *Activity) error {
	query := durable.InsertQueryOrUpdate("activity", "activity_index", "client_id,status,img_url,expire_img_url,action,start_at,expire_at")
	_, err := session.Database(ctx).Exec(ctx, query, a.ActivityIndex, a.ClientID, a.Status, a.ImgURL, a.ExpireImgURL, a.Action, a.StartAt, a.ExpireAt)
	return err
}
