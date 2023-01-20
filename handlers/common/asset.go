package common

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func GetAssetByID(ctx context.Context, client *mixin.Client, assetID string) (models.Asset, error) {
	if assetID == "" {
		return models.Asset{}, nil
	}
	var a models.Asset
	err := session.Redis(ctx).StructScan(ctx, "asset:"+assetID, &a)
	if err == nil || !errors.Is(err, redis.Nil) {
		return a, err
	}
	defer func() {
		if a.AssetID != "" {
			if err := session.Redis(ctx).StructSet(ctx, "asset:"+assetID, a); err != nil {
				tools.Println(err)
			}
		}
	}()
	err = session.DB(ctx).Take(&a, "asset_id = ?", assetID).Error
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		return a, err
	}
	asset, err := SetAssetByID(ctx, client, assetID)
	if err != nil {
		return a, err
	}
	a = *asset
	return a, nil
}

func SetAssetByID(ctx context.Context, client *mixin.Client, assetID string) (*models.Asset, error) {
	_a, err := mixin.ReadNetworkAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	a := models.Asset{
		AssetID:   assetID,
		ChainID:   _a.ChainID,
		IconUrl:   _a.IconURL,
		Symbol:    _a.Symbol,
		Name:      _a.Name,
		PriceUsd:  _a.PriceUSD,
		ChangeUsd: _a.ChangeUsd.String(),
	}

	if err := session.DB(ctx).Save(&a).Error; err != nil {
		tools.Println(err)
		return nil, err
	}

	return &a, nil
}

func GetUserAssets(ctx context.Context, u *models.ClientUser) ([]*mixin.Asset, error) {
	var assets []*mixin.Asset
	var err error
	if u.AccessToken != "" {
		assets, err = mixin.ReadAssets(ctx, u.AccessToken)
	} else if u.AuthorizationID != "" {
		assets, err = getUserAssetsByClientUser(ctx, u)
	} else {
		return nil, session.ForbiddenError(ctx)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			return GetUserAssets(ctx, u)
		}
	}
	return assets, nil
}

func getUserAssetsByClientUser(ctx context.Context, u *models.ClientUser) ([]*mixin.Asset, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadAssets(ctx)
}

func GetUserAsset(ctx context.Context, u *models.ClientUser, assetID string) (*mixin.Asset, error) {
	var asset *mixin.Asset
	var err error
	if u.AccessToken != "" {
		asset, err = mixin.ReadAsset(ctx, u.AccessToken, assetID)
	} else if u.AuthorizationID != "" {
		asset, err = getUserAssetByClientUser(ctx, u, assetID)
	} else {
		return nil, session.ForbiddenError(ctx)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			tools.Println(err)
			return GetUserAsset(ctx, u, assetID)
		}
	}
	return asset, nil
}

func getUserAssetByClientUser(ctx context.Context, u *models.ClientUser, id string) (*mixin.Asset, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadAsset(ctx, id)
}

func GetUserSnapshots(ctx context.Context, u *models.ClientUser, assetID string, offset time.Time, order string, limit int) ([]*mixin.Snapshot, error) {
	var ss []*mixin.Snapshot
	var err error
	if u.AccessToken != "" {
		ss, err = mixin.ReadSnapshots(ctx, u.AccessToken, assetID, offset, order, limit)
	} else if u.AuthorizationID != "" {
		ss, err = getUserSnapshotByClientUser(ctx, u, assetID, offset, order, limit)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			tools.Println(err)
			return GetUserSnapshots(ctx, u, assetID, offset, order, limit)
		}
	}
	return ss, err
}

func getUserSnapshotByClientUser(ctx context.Context, u *models.ClientUser, assetID string, offset time.Time, order string, limit int) ([]*mixin.Snapshot, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadSnapshots(ctx, assetID, offset, order, limit)
}
