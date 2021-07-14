package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

const (
	BuildVersion = "BUILD_VERSION"
)

const (
	MessageShardSize           = int64(20)
	CacheTime                  = 15 * time.Minute
	AssetsCheckTime            = 12 * time.Hour
	NotActiveCheckTime         = 7 * 24.0
	NotOpenAssetsCheckMsgLimit = 10
)

type config struct {
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Name     string `json:"name"`
	} `json:"database"`

	Monitor struct {
		ClientID       string `json:"client_id"`
		SessionID      string `json:"session_id"`
		PrivateKey     string `json:"private_key"`
		ConversationID string `json:"conversation_id"`
	} `json:"monitor"`

	Qiniu struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		Bucket    string `json:"bucket"`
	} `json:"qiniu"`

	RedisAddr string `json:"redis_addr"`

	ClientList     []string `json:"client_list"`
	ShowClientList []string `json:"show_client_list"`
	LuckCoinAppID  string   `json:"luck_coin_app_id"`

	Text struct {
		Desc            string            `json:"desc"`
		Join            string            `json:"join"`
		Home            string            `json:"home"`
		News            string            `json:"news"`
		Transfer        string            `json:"transfer"`
		Activity        string            `json:"activity"`
		Auth            string            `json:"auth"`
		Forward         string            `json:"forward"`
		Mute            string            `json:"mute"`
		Block           string            `json:"block"`
		JoinMsg         string            `json:"join_msg"`
		AuthSuccess     string            `json:"auth_success"`
		PrefixLeaveMsg  string            `json:"prefix_leave_msg"`
		LeaveGroup      string            `json:"leave_group"`
		OpenChatStatus  string            `json:"open_chat_status"`
		CloseChatStatus string            `json:"close_chat_status"`
		MuteOpen        string            `json:"mute_open"`
		MuteClose       string            `json:"mute_close"`
		Muting          string            `json:"muting"`
		VideoLiving     string            `json:"video_living"`
		VideoLiveEnd    string            `json:"video_live_end"`
		Living          string            `json:"living"`
		LiveEnd         string            `json:"live_end"`
		Category        map[string]string `json:"category"`
		WelcomeUpdate   string            `json:"welcome_update"`
		StopMessage     string            `json:"stop_message"`
		StopClose       string            `json:"stop_close"`
		StopBroadcast   string            `json:"stop_broadcast"`
	} `json:"text"`
}

var Config config

func init() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("config.json open fail...", err)
		return
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Println("config.json parse err...", err)
	}
	log.Println("config.json load success...")
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
