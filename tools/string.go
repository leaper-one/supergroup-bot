package tools

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var colors = []string{
	"7983C2", "8F7AC5", "C5595A", "C97B46", "76A048", "3D98D0",
	"5979F0", "8A64D0", "B76753", "AA8A46", "9CAD23", "6BC0CE",
	"6C89D3", "AA66C3", "C8697D", "C49B4B", "5FB05F", "52A98B",
	"75A2CB", "A75C96", "9B6D77", "A49373", "6AB48F", "93B289",
}

func RandomColor() string {
	return "#" + colors[rand.Intn(len(colors))]
}

func SplitString(s string, length int) string {
	res := []rune(s)
	if len(res) > length {
		s = string(res[:length-3]) + "..."
	}
	return s
}

func RandomString(letter []rune, n int) string {
	rand.Seed(time.Now().UnixNano())
	if len(letter) == 0 {
		letter = letters
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomNumber(n int) string {
	var s string
	for i := 0; i < n; i++ {
		s += strconv.Itoa(rand.Intn(10))
	}
	return s
}

func PrintJson(d interface{}) {

	s, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
		return
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, s, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		return
	}
	log.Println(prettyJSON.String())
}

func WriteDataToFile(fileName string, data interface{}) {
	s, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	err = ioutil.WriteFile(fileName, s, 0644)
	if err != nil {
		log.Println(err)
		return
	}
}

var inviteCode = []rune("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz")

func GetRandomInvitedCode() string {
	return RandomString(inviteCode, 6)
}

var voucherCode = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ123456789")

func GetRandomVoucherCode() string {
	return RandomString(voucherCode, 6)
}
