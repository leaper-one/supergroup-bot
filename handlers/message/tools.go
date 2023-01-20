package message

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"

	"github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup/models"
	"golang.org/x/crypto/curve25519"
)

func encryptMessageData(data, pk string, sessions []*models.Session) (string, error) {
	ss := make([]*bot.Session, len(sessions))
	for i, s := range sessions {
		ss[i] = &bot.Session{
			UserID:    s.UserID,
			SessionID: s.SessionID,
			PublicKey: s.PublicKey,
		}
	}
	return bot.EncryptMessageData(data, ss, pk)
}

func decryptMessageData(data string, client *models.Client) (string, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	size := 16 + 48 // session id bytes + encypted key bytes size
	total := len(bytes)
	if total < 1+2+32+size+12 {
		return "", nil
	}
	sessionLen := int(binary.LittleEndian.Uint16(bytes[1:3]))
	prefixSize := 35 + sessionLen*size
	var key []byte
	for i := 35; i < prefixSize; i += size {
		if uid, _ := bot.UuidFromBytes(bytes[i : i+16]); uid.String() == client.SessionID {
			private, err := base64.RawURLEncoding.DecodeString(client.PrivateKey)
			if err != nil {
				return "", err
			}
			var dst, priv, pub [32]byte
			copy(pub[:], bytes[3:35])
			bot.PrivateKeyToCurve25519(&priv, ed25519.PrivateKey(private))
			curve25519.ScalarMult(&dst, &priv, &pub)
			block, err := aes.NewCipher(dst[:])
			if err != nil {
				return "", err
			}
			iv := bytes[i+16 : i+16+aes.BlockSize]
			key = bytes[i+16+aes.BlockSize : i+size]
			mode := cipher.NewCBCDecrypter(block, iv)
			mode.CryptBlocks(key, key)
			key = key[:16]
			break
		}
	}
	if len(key) != 16 {
		return "", nil
	}
	nonce := bytes[prefixSize : prefixSize+12]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}
	plaintext, err := aesgcm.Open(nil, nonce, bytes[prefixSize+12:], nil)
	if err != nil {
		return "", nil
	}
	return base64.RawURLEncoding.EncodeToString(plaintext), nil
}
