package models

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
)

const users_DDL = `
CREATE TABLE IF NOT EXISTS users (
	user_id	          VARCHAR(36) PRIMARY KEY,
  identity_number   VARCHAR NOT NULL UNIQUE,
	access_token      VARCHAR(512),
	full_name         VARCHAR(512),
	avatar_url        VARCHAR(1024),
	is_scam           BOOLEAN,
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type User struct {
	UserID              string    `json:"user_id,omitempty"`
	IdentityNumber      string    `json:"identity_number,omitempty"`
	AccessToken         string    `json:"access_token,omitempty"`
	FullName            string    `json:"full_name,omitempty"`
	AvatarURL           string    `json:"avatar_url,omitempty"`
	IsScam              bool      `json:"is_scam,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	AuthenticationToken string    `json:"authentication_token,omitempty"`

	IsNew bool `json:"is_new,omitempty"`
}

const (
	DefaultAvatar = "https://images.mixin.one/E2y0BnTopFK9qey0YI-8xV3M82kudNnTaGw0U5SU065864SsewNUo6fe9kDF1HIzVYhXqzws4lBZnLj1lPsjk-0=s128"
)

func AuthenticateUserByOAuth(ctx context.Context, host, authCode, inviteCode string) (*User, error) {
	client, err := GetMixinClientByIDOrHost(ctx, host)
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
	go hanldeUserInvite(inviteCode, client.ClientID, user.UserID)
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

func GenerateAuthenticationToken(ctx context.Context, userId, accessKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
	})
	sum := sha256.Sum256([]byte(accessKey))
	return token.SignedString(sum[:])
}

func AuthenticateUserByToken(ctx context.Context, host, authenticationToken string) (*ClientUser, error) {
	client, err := GetMixinClientByIDOrHost(ctx, host)
	if err != nil {
		return nil, err
	}
	if client.ClientID == "" {
		return nil, session.BadDataError(ctx)
	}
	var user ClientUser
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
		user, queryErr = GetClientUserByClientIDAndUserID(ctx, client.ClientID, fmt.Sprint(claims["jti"]))
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

type UserMeResp struct {
	*ClientUser
	FullName string `json:"full_name"`
	IsClaim  bool   `json:"is_claim"`
	IsBlock  bool   `json:"is_block"`
	IsProxy  bool   `json:"is_proxy"`
}

func GetMe(ctx context.Context, u *ClientUser) UserMeResp {
	req := session.Request(ctx)
	go createLoginLog(u, req.RemoteAddr, req.Header.Get("User-Agent"))
	proxy, _ := getClientUserProxyByProxyID(ctx, u.ClientID, u.UserID)
	me := UserMeResp{
		ClientUser: u,
		IsClaim:    checkIsClaim(ctx, u.UserID),
		IsBlock:    checkIsBlockUser(ctx, u.ClientID, u.UserID),
		IsProxy:    proxy.Status == ClientUserProxyStatusActive,
		FullName:   proxy.FullName,
	}
	return me
}

func checkAndWriteUser(ctx context.Context, client *MixinClient, accessToken string, store *mixin.OauthKeystore) (*User, error) {
	var u *mixin.User
	var err error
	if store != nil && store.AuthID != "" {
		client, err1 := mixin.NewFromOauthKeystore(store)
		if err1 != nil {
			return nil, err1
		}
		u, err = client.UserMe(ctx)
	} else if accessToken != "" {
		u, err = mixin.UserMe(ctx, accessToken)
	} else {
		return nil, session.BadDataError(ctx)
	}
	if err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(u.UserID); err != nil {
		return nil, session.BadDataError(ctx)
	}
	if u.AvatarURL == "" {
		u.AvatarURL = DefaultAvatar
	}
	user := &User{
		UserID:         u.UserID,
		FullName:       u.FullName,
		IdentityNumber: u.IdentityNumber,
		AvatarURL:      u.AvatarURL,
		IsScam:         u.IsScam,
		CreatedAt:      time.Now(),
	}
	if err := WriteUser(ctx, user); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	clientUser := ClientUser{
		ClientID:        client.ClientID,
		UserID:          u.UserID,
		AccessToken:     accessToken,
		Priority:        ClientUserPriorityLow,
		Status:          0,
		AssetID:         client.C.AssetID,
		AuthorizationID: store.AuthID,
		Scope:           store.Scope,
		PrivateKey:      store.PrivateKey,
		Ed25519:         store.VerifyKey,
	}
	status, err := GetClientUserStatusByClientUser(ctx, &clientUser)
	if err != nil && !errors.Is(err, session.ForbiddenError(ctx)) {
		return nil, err
	} else {
		clientUser.Status = status
		if status > 1 {
			clientUser.Priority = ClientUserPriorityHigh
		}
	}

	isNewUser, err := UpdateClientUser(ctx, &clientUser, u.FullName)
	if err != nil {
		return nil, err
	}
	user.IsNew = isNewUser
	return user, nil
}

func WriteUser(ctx context.Context, user *User) error {
	query := durable.InsertQueryOrUpdate("users", "user_id", "identity_number, full_name, avatar_url, is_scam")
	_, err := session.Database(ctx).Exec(ctx, query, user.UserID, user.IdentityNumber, user.FullName, user.AvatarURL, user.IsScam)
	if err != nil {
		return err
	}
	return nil
}

func SendMsgToDeveloper(ctx context.Context, clientID, msg string) {
	userID := config.Config.Dev
	if userID == "" {
		return
	}
	var client *mixin.Client

	if clientID == "" && config.Config.Monitor.ClientID != "" {
		k := config.Config.Monitor
		clientID = k.ClientID
		client, _ = mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   k.ClientID,
			SessionID:  k.SessionID,
			PrivateKey: k.PrivateKey,
		})
	} else {
		clientID = GetFirstClient(ctx).ClientID
		c, err := GetMixinClientByIDOrHost(ctx, clientID)
		if err != nil {
			return
		}
		client = c.Client
	}

	conversationID := mixin.UniqueConversationID(clientID, userID)
	_ = client.SendMessage(ctx, &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte("super group log..." + msg)),
	})
}

func getUserByID(ctx context.Context, userID string) (*mixin.User, error) {
	var u mixin.User
	err := session.Database(ctx).QueryRow(ctx, "SELECT user_id, identity_number, full_name, avatar_url, is_scam FROM users WHERE user_id = $1", userID).Scan(&u.UserID, &u.IdentityNumber, &u.FullName, &u.AvatarURL, &u.IsScam)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return SearchUser(ctx, userID)
	}
	return &u, err
}

func GetUserByIdentityNumber(ctx context.Context, identityNumber string) (*mixin.User, error) {
	var u mixin.User
	err := session.Database(ctx).
		QueryRow(ctx, "SELECT user_id, identity_number, full_name, avatar_url FROM users WHERE identity_number = $1", identityNumber).
		Scan(&u.UserID, &u.IdentityNumber, &u.FullName, &u.AvatarURL)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return SearchUser(ctx, identityNumber)
	}
	return &u, err
}

func SearchUser(ctx context.Context, userIDOrIdentityNumber string) (*mixin.User, error) {
	u, err := GetFirstClient(ctx).ReadUser(ctx, userIDOrIdentityNumber)
	if err != nil {
		return nil, err
	}
	_ = WriteUser(ctx, &User{
		UserID:         u.UserID,
		IdentityNumber: u.IdentityNumber,
		FullName:       u.FullName,
		AvatarURL:      u.AvatarURL,
		IsScam:         u.IsScam,
	})
	return u, err
}

func checkUserIsScam(ctx context.Context, userID string) bool {
	u, err := getUserByID(ctx, userID)
	if err != nil {
		session.Logger(ctx).Println(err, userID)
		return false
	}
	return u.IsScam
}
