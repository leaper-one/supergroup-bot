package tools

import (
	"encoding/base64"
)

func Base64Decode(str string) []byte {
	decodeBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return make([]byte, 0)
	}
	return decodeBytes
}

func Base64Encode(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}
