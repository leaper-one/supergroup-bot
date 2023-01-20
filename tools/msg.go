package tools

import (
	"context"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/fox-one/mixin-sdk-go"
)

func SendMsgToDeveloper(msg string) {
	userID := config.Config.Dev
	if userID == "" {
		return
	}
	var client *mixin.Client
	if config.Config.Monitor.ClientID == "" {
		return
	}

	k := config.Config.Monitor
	client, _ = mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   k.ClientID,
		SessionID:  k.SessionID,
		PrivateKey: k.PrivateKey,
	})

	conversationID := mixin.UniqueConversationID(k.ClientID, userID)
	_ = client.SendMessage(context.Background(), &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           Base64Encode([]byte("super group log..." + msg)),
	})
}

func SendMonitorGroupMsg(msg string) {
	m := config.Config.Monitor
	if m.ConversationID == "" || m.ClientID == "" || m.SessionID == "" || m.PrivateKey == "" {
		return
	}
	msgClient, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   m.ClientID,
		SessionID:  m.SessionID,
		PrivateKey: m.PrivateKey,
	})
	if err != nil {
		Println("SendMonitorGroupMsg mixin.NewFromKeystore error %v", err)
		return
	}
	if err := msgClient.SendMessage(context.Background(), &mixin.MessageRequest{
		ConversationID: m.ConversationID,
		Data:           Base64Encode([]byte(msg)),
		Category:       mixin.MessageCategoryPlainText,
		MessageID:      GetUUID(),
	}); err != nil {
		Println(err)
	}
}
