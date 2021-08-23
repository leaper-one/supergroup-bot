package config

import (
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
	NotActiveCheckTime         = 14 * 24.0
	NotOpenAssetsCheckMsgLimit = 10
)

type config struct {
	Lang     string `json:"lang"`
	Port     int    `json:"port"`
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
}

type text struct {
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
	StickerWarning  string            `json:"sticker_warning"`
	StatusSet       string            `json:"status_set"`
	StatusCancel    string            `json:"status_cancel"`
	StatusAdmin     string            `json:"status_admin"`
	StatusGuest     string            `json:"status_guest"`
	Reward          string            `json:"reward"`
	From            string            `json:"from"`
	MemberCentre    string            `json:"member_centre"`
	PayForFresh     string            `json:"pay_for_fresh"`
	PayForLarge     string            `json:"pay_for_large"`
	AuthForFresh    string            `json:"auth_for_fresh"`
	AuthForLarge    string            `json:"auth_for_large"`

	LimitReject    string `json:"limit_reject"`
	MutedReject    string `json:"muted_reject"`
	URLReject      string `json:"url_reject"`
	URLAdmin       string `json:"url_admin"`
	LanguageReject string `json:"language_reject"`
	LanguageAdmin  string `json:"language_admin"`
	BalanceReject  string `json:"balance_reject"`
	CategoryReject string `json:"category_reject"`
	Forbid         string `json:"forbid"`
	BotCard        string `json:"bot_card"`
}

var Config config

var Text text

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
	if Config.Lang == "zh" {
		Text = zh_CN_Text
	} else if Config.Lang == "en" {
		Text = en_Text
	}

	log.Println("config.json load success...")
}
