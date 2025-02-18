package message

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/fox-one/mixin-sdk-go"
)

type EncryptedMessageResp struct {
	MessageID   string `json:"message_id"`
	RecipientID string `json:"recipient_id"`
	State       string `json:"state"`
	Sessions    []struct {
		SessionID string `json:"session_id"`
		PublicKey string `json:"public_key"`
	} `json:"sessions"`
}

func SendEncryptedMessage(ctx context.Context, pk string, client *mixin.Client, msgs []*mixin.MessageRequest) ([]*EncryptedMessageResp, error) {
	var userIDs []string
	for _, m := range msgs {
		userIDs = append(userIDs, m.RecipientID)
	}
	sessionSet, err := ReadSessionSetByUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	var body []map[string]interface{}
	for _, message := range msgs {
		if message.RepresentativeID == client.ClientID {
			message.RepresentativeID = ""
		}
		if message.Category == mixin.MessageCategoryMessageRecall {
			message.RepresentativeID = ""
		}
		m := map[string]interface{}{
			"conversation_id":   message.ConversationID,
			"recipient_id":      message.RecipientID,
			"message_id":        message.MessageID,
			"quote_message_id":  message.QuoteMessageID,
			"category":          message.Category,
			"data_base64":       message.Data,
			"silent":            false,
			"representative_id": message.RepresentativeID,
		}
		recipient := sessionSet[message.RecipientID]
		category := readEncryptCategory(message.Category, recipient)
		m["category"] = category
		if recipient != nil {
			m["checksum"] = GenerateUserChecksum(recipient.Sessions)
			var sessions []map[string]string
			for _, s := range recipient.Sessions {
				sessions = append(sessions, map[string]string{"session_id": s.SessionID})
			}
			m["recipient_sessions"] = sessions
			if strings.Contains(category, "ENCRYPTED") {
				data, err := encryptMessageData(message.Data, pk, recipient.Sessions)
				if err != nil {
					return nil, err
				}
				m["data_base64"] = data
			}
		}
		body = append(body, m)
	}
	return SendEncryptMessages(client, body)
}

func SendEncryptMessages(client *mixin.Client, body []map[string]interface{}) ([]*EncryptedMessageResp, error) {
	var resp []*EncryptedMessageResp
	err := client.Post(context.Background(), "/encrypted_messages", body, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			return nil, nil
		}
		if common.LogWithNotNetworkError(err) {
			data, _ := json.Marshal(body)
			log.Println("encrypt msg 3...", string(data))
		}
		log.Println("encrypt msg 4...", err)
		time.Sleep(time.Millisecond * 100)
		return SendEncryptMessages(client, body)
	}
	return resp, nil
}

func readEncryptCategory(category string, user *SimpleUser) string {
	if user == nil {
		return strings.Replace(category, "ENCRYPTED_", "PLAIN_", -1)
	}
	switch user.Category {
	case UserCategoryPlain:
		return strings.Replace(category, "ENCRYPTED_", "PLAIN_", -1)
	case UserCategoryEncrypted:
		return strings.Replace(category, "PLAIN_", "ENCRYPTED_", -1)
	default:
		return category
	}
}
