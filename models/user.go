package models

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/tools"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/dgrijalva/jwt-go"
	"github.com/fox-one/mixin-sdk-go"
	uuid "github.com/satori/go.uuid"
)

const users_DDL = `
CREATE TABLE IF NOT EXISTS users (
	user_id	          VARCHAR(36) PRIMARY KEY,
  identity_number   VARCHAR NOT NULL UNIQUE,
	access_token      VARCHAR(512),
	full_name         VARCHAR(512),
	avatar_url        VARCHAR(1024),
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
  );
`

type User struct {
	UserID              string    `json:"user_id,omitempty"`
	IdentityNumber      string    `json:"identity_number,omitempty"`
	AccessToken         string    `json:"access_token,omitempty"`
	FullName            string    `json:"full_name,omitempty"`
	AvatarURL           string    `json:"avatar_url,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	AuthenticationToken string    `json:"authentication_token,omitempty"`
}

const (
	DefaultAvatar = "https://images.mixin.one/E2y0BnTopFK9qey0YI-8xV3M82kudNnTaGw0U5SU065864SsewNUo6fe9kDF1HIzVYhXqzws4lBZnLj1lPsjk-0=s128"
)

func AuthenticateUserByOAuth(ctx context.Context, host, authorizationCode string) (*User, error) {
	client := GetMixinClientByHost(ctx, host)
	if client.ClientID == "" {
		return nil, session.BadDataError(ctx)
	}
	accessToken, scope, err := mixin.AuthorizeToken(ctx, client.ClientID, client.Secret, authorizationCode, "")
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			return nil, session.ForbiddenError(ctx)
		}
		return nil, session.BadDataError(ctx)
	}
	if !strings.Contains(scope, "PROFILE:READ") {
		return nil, session.ForbiddenError(ctx)
	}
	if !strings.Contains(scope, "MESSAGES:REPRESEN") {
		return nil, session.ForbiddenError(ctx)
	}
	me, err := mixin.UserMe(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if me == nil {
		return nil, session.BadDataError(ctx)
	}

	user, err := checkAndWriteUser(ctx, client, me.UserID, accessToken, me.FullName, me.AvatarURL, me.IdentityNumber)
	if err != nil {
		return nil, err
	}
	authenticationToken, err := generateAuthenticationToken(ctx, user.UserID, accessToken)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	user.AuthenticationToken = authenticationToken
	return user, nil
}

func generateAuthenticationToken(ctx context.Context, userId, accessToken string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
	})
	sum := sha256.Sum256([]byte(accessToken))
	return token.SignedString(sum[:])
}

func AuthenticateUserByToken(ctx context.Context, host, authenticationToken string) (*ClientUser, error) {
	client := GetMixinClientByHost(ctx, host)
	if client.ClientID == "" {
		return nil, session.BadDataError(ctx)
	}
	var user *ClientUser
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
		user, queryErr = GetClientUserByClientIDAndUserID(ctx, GetMixinClientByHost(ctx, host).ClientID, fmt.Sprint(claims["jti"]))
		if queryErr != nil {
			return nil, queryErr
		}
		if user == nil {
			return nil, session.BadDataError(ctx)
		}
		sum := sha256.Sum256([]byte(user.AccessToken))
		return sum[:], nil
	})
	if queryErr != nil {
		return nil, queryErr
	}
	if err != nil || !token.Valid {
		return nil, nil
	}
	return user, nil
}

func checkAndWriteUser(ctx context.Context, client MixinClient, userId, accessToken, fullName, avatarURL, identityNumber string) (*User, error) {
	if _, err := uuid.FromString(userId); err != nil {
		return nil, session.BadDataError(ctx)
	}
	if avatarURL == "" {
		avatarURL = DefaultAvatar
	}
	user := &User{
		UserID:         userId,
		FullName:       fullName,
		AccessToken:    accessToken,
		IdentityNumber: identityNumber,
		AvatarURL:      avatarURL,
		CreatedAt:      time.Now(),
	}
	if err := WriteUser(ctx, user); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	clientUser := ClientUser{
		ClientID:    client.ClientID,
		UserID:      userId,
		AccessToken: accessToken,
		Priority:    ClientUserPriorityLow,
		Status:      0,
		AssetID:     client.AssetID,
	}
	if client.ClientID != "11efbb75-e7fe-44d7-a14f-698535289310" {
		status, err := GetClientUserStatusByClientUser(ctx, &clientUser)
		if err != nil && !errors.Is(err, session.ForbiddenError(ctx)) {
			return nil, err
		} else {
			clientUser.Status = status
			if status > 1 {
				clientUser.Priority = ClientUserPriorityHigh
			}
		}
	} else {
		clientUser.Status = ClientUserStatusLarge
		clientUser.Priority = ClientUserPriorityHigh
	}

	if err := UpdateClientUser(ctx, &clientUser, fullName); err != nil {
		return nil, err
	}
	return user, nil
}

func WriteUser(ctx context.Context, user *User) error {
	query := durable.InsertQueryOrUpdate("users", "user_id", "identity_number, full_name, avatar_url")
	_, err := session.Database(ctx).Exec(ctx, query, user.UserID, user.IdentityNumber, user.FullName, user.AvatarURL)
	if err != nil {
		return err
	}
	return nil
}

func SendMsgToManager(ctx context.Context, clientID, msg string) {
	userID := "e8e8cd79-cd40-4796-8c54-3a13cfe50115"
	if clientID == "" {
		clientID = GetFirstClient(ctx).ClientID
	}

	conversationID := mixin.UniqueConversationID(clientID, userID)
	client := GetMixinClientByID(ctx, clientID)
	_ = client.SendMessage(ctx, &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte(msg)),
	})
}

func SearchUser(ctx context.Context, userIDOrIdentityNumber string) (*mixin.User, error) {
	return GetFirstClient(ctx).ReadUser(ctx, userIDOrIdentityNumber)
}
