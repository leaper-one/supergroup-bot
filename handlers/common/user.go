package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func SearchUser(ctx context.Context, clientID, userIDOrIdentityNumber string) (*models.User, error) {
	var u models.User
	err := session.Redis(ctx).StructScan(ctx, fmt.Sprintf("user:%s", userIDOrIdentityNumber), &u)
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, redis.Nil) {
		return nil, err
	}
	defer func() {
		if u.UserID != "" {
			if err := session.Redis(ctx).StructSet(ctx, fmt.Sprintf("user:%s", userIDOrIdentityNumber), &u); err != nil {
				tools.Println(err)
			}
		}
	}()
	err = session.DB(ctx).Take(&u, "user_id = ? OR identity_number = ?", userIDOrIdentityNumber, userIDOrIdentityNumber).Error
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return nil, err
	}
	_u, err := client.ReadUser(ctx, userIDOrIdentityNumber)
	if err != nil {
		return nil, err
	}
	u = models.User{
		UserID:         _u.UserID,
		IdentityNumber: _u.IdentityNumber,
		FullName:       _u.FullName,
		AvatarURL:      _u.AvatarURL,
		IsScam:         _u.IsScam,
		CreatedAt:      _u.CreatedAt,
	}
	if err := session.DB(ctx).Create(&u).Error; err != nil {
		return nil, err
	}
	return &u, err
}
