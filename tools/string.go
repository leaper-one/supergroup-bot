package tools

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var colors = []string{
	"#0x7983C2", "#0x8F7AC5", "#0xC5595A", "#0xC97B46", "#0x76A048", "#0x3D98D0",
	"#0x5979F0", "#0x8A64D0", "#0xB76753", "#0xAA8A46", "#0x9CAD23", "#0x6BC0CE",
	"#0x6C89D3", "#0xAA66C3", "#0xC8697D", "#0xC49B4B", "#0x5FB05F", "#0x52A98B",
	"#0x75A2CB", "#0xA75C96", "#0x9B6D77", "#0xA49373", "#0x6AB48F", "#0x93B289",
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
