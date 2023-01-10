package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/golang-jwt/jwt"
)

func AuthenticateUserByToken(ctx context.Context, host, authenticationToken string) (*models.ClientUser, error) {
	client, err := common.GetMixinClientByIDOrHost(ctx, host)
	if err != nil {
		return nil, err
	}
	if client.ClientID == "" {
		return nil, session.BadDataError(ctx)
	}
	var user models.ClientUser
	var queryErr error
	token, err := jwt.Parse(authenticationToken, func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, session.BadDataError(ctx)
		}
		_, ok = token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, session.BadDataError(ctx)
		}
		user, queryErr = common.GetClientUserByClientIDAndUserID(ctx, client.ClientID, fmt.Sprint(claims["jti"]))
		if queryErr != nil {
			return nil, queryErr
		}
		if user.UserID == "" {
			return nil, session.BadDataError(ctx)
		}
		var key string
		if user.AccessToken != "" {
			key = user.AccessToken
		} else if user.AuthorizationID != "" {
			key = user.AuthorizationID
		} else {
			return nil, session.BadDataError(ctx)
		}
		sum := sha256.Sum256([]byte(key))
		return sum[:], nil
	})
	if queryErr != nil {
		return nil, queryErr
	}
	if err != nil || !token.Valid {
		return nil, nil
	}
	return &user, nil
}

func AuthenticateUserByOAuth(ctx context.Context, host, authCode, inviteCode string) (*models.User, error) {
	client, err := common.GetMixinClientByIDOrHost(ctx, host)
	if err != nil {
		return nil, err
	}
	var accessToken string
	oauth := new(mixin.OauthKeystore)
	if strings.Contains(client.C.PrivateKey, "RSA PRIVATE KEY") {
		accessToken, oauth.Scope, err = mixin.AuthorizeToken(ctx, client.ClientID, client.C.ClientSecret, authCode, "")
	} else {
		key := mixin.GenerateEd25519Key()
		oauth, err = mixin.AuthorizeEd25519(ctx, client.ClientID, client.C.ClientSecret, authCode, "", key)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			return nil, session.ForbiddenError(ctx)
		}
		return nil, session.BadDataError(ctx)
	}
	if !strings.Contains(oauth.Scope, "PROFILE:READ") {
		return nil, session.ForbiddenError(ctx)
	}
	if !strings.Contains(oauth.Scope, "MESSAGES:REPRESEN") {
		return nil, session.ForbiddenError(ctx)
	}
	user, err := checkAndWriteUser(ctx, client, accessToken, oauth)
	if err != nil || user == nil {
		return nil, session.BadDataError(ctx)
	}
	go common.HandleUserInvite(inviteCode, client.ClientID, user.UserID)
	if err != nil {
		return nil, err
	}
	accessKey := accessToken
	if accessKey == "" {
		accessKey = oauth.AuthID
	}
	authenticationToken, err := GenerateAuthenticationToken(ctx, user.UserID, accessKey)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	user.AuthenticationToken = authenticationToken
	return user, nil
}

func checkAndWriteUser(ctx context.Context, client *common.MixinClient, accessToken string, store *mixin.OauthKeystore) (*models.User, error) {
	var u *mixin.User
	var err error
	if accessToken != "" {
		u, err = mixin.UserMe(ctx, accessToken)
	} else if store != nil && store.AuthID != "" {
		client, err1 := mixin.NewFromOauthKeystore(store)
		if err1 != nil {
			return nil, err1
		}
		u, err = client.UserMe(ctx)
	} else {
		return nil, session.BadDataError(ctx)
	}
	if err != nil {
		return nil, err
	}
	if u.AvatarURL == "" {
		u.AvatarURL = models.DefaultAvatar
	}
	user := models.User{
		UserID:         u.UserID,
		FullName:       u.FullName,
		IdentityNumber: u.IdentityNumber,
		AvatarURL:      u.AvatarURL,
		IsScam:         u.IsScam,
		CreatedAt:      time.Now(),
	}
	if common.GetClientProxy(ctx, client.ClientID) == models.ClientProxyStatusOn {
		if err := common.UpdateClientUserProxy(ctx,
			&models.ClientUser{
				ClientID: client.ClientID,
				UserID:   u.UserID,
			},
			true,
			u.FullName,
			u.AvatarURL); err != nil {
			tools.Println(err)
		}
	}
	if err := session.DB(ctx).Save(&user).Error; err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	clientUser := models.ClientUser{
		ClientID:        client.ClientID,
		UserID:          u.UserID,
		AccessToken:     accessToken,
		Priority:        models.ClientUserPriorityLow,
		Status:          0,
		AuthorizationID: store.AuthID,
		Scope:           store.Scope,
		PrivateKey:      store.PrivateKey,
		Ed25519:         store.VerifyKey,
	}
	status, err := common.GetClientUserStatusByClientUser(ctx, &clientUser)
	if err != nil && !errors.Is(err, session.ForbiddenError(ctx)) {
		return nil, err
	} else {
		clientUser.Status = status
		if status > 1 {
			clientUser.Priority = models.ClientUserPriorityHigh
		}
	}

	isNewUser, err := UpdateClientUser(ctx, clientUser, u.FullName)
	if err != nil {
		return nil, err
	}
	user.IsNew = isNewUser
	return &user, nil
}

func GenerateAuthenticationToken(ctx context.Context, userId, accessKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
	})
	sum := sha256.Sum256([]byte(accessKey))
	return token.SignedString(sum[:])
}
