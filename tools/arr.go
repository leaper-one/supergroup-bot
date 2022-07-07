package tools

import (
	"crypto/aes"
	"crypto/cipher"
)

func Includes(arr []string, target string) bool {
	for _, s := range arr {
		if target == s {
			return true
		}
	}
	return false
}

func Reverse(arr []interface{}) []interface{} {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		temp := arr[length-1-i]
		arr[length-1-i] = arr[i]
		arr[i] = temp
	}
	return arr
}

func DecryptAttachment(data, keys, digest []byte) ([]byte, error) {
	aesKey := keys[:32]
	iv := data[:16]
	ciphertext := data[16 : len(data)-32]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	return ciphertext, nil
}
