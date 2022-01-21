package services

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	lrs := make([]*models.LiveReplay, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT category,data,created_at
FROM live_replay
WHERE live_id=$1
ORDER BY created_at
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var lr models.LiveReplay
			if err := rows.Scan(&lr.Category, &lr.Data, &lr.CreatedAt); err != nil {
				return err
			}
			lrs = append(lrs, &lr)
		}
		return nil
	}, "223602a0-e159-42ea-a52b-b4f071c8ff1b")
	if err != nil {
		return err
	}
	for _, v := range lrs {
		if v.Category == mixin.MessageCategoryPlainText {
			session.Database(ctx).Exec(ctx, `
UPDATE live_replay
SET data=$1
WHERE created_at=$2
`, string(tools.Base64Decode(v.Data)), v.CreatedAt)
		}
	}
	return nil
}
