package tools

import (
	"encoding/base64"
	"strings"
)

func Base64Decode(str string) []byte {
	decodeBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		decodeBytes, err := base64.RawURLEncoding.DecodeString(str)
		if err != nil {
			return make([]byte, 0)
		}
		return decodeBytes
	}
	return decodeBytes
}

func Base64Encode(str []byte) string {
	return base64.RawURLEncoding.EncodeToString(str)
}

func SafeBase64Encode(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "=", "")
	return s
}
