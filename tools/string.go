package tools

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var color = []rune("0123456789abcdef")

func RandomColor() string {
	b := make([]rune, 6)

	for i := range b {
		b[i] = color[rand.Intn(len(color))]
	}

	return "#" + string(b)
}

func RandomString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
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
	log.Println(string(prettyJSON.Bytes()))
}
