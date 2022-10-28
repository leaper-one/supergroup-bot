package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

const (
	BuildVersion = "BUILD_VERSION"
)

const (
	MessageShardSize           = int64(16)
	CacheTime                  = 15 * time.Minute
	AssetsCheckTime            = 12 * time.Hour
	NotActiveCheckTime         = 14 * 24.0
	NotOpenAssetsCheckMsgLimit = 10
	NoticeLotteryTimes         = 5
	UpdateUserDeliverTime      = 30 * time.Minute

	QuoteMsgSavedTime = 48 * time.Hour
)

var LangCheckPer = decimal.NewFromInt(2).Div(decimal.NewFromInt(3))

type Lottery struct {
	LotteryID string          `json:"lottery_id"`
	AssetID   string          `json:"asset_id"`
	Amount    decimal.Decimal `json:"amount"`
	IconURL   string          `json:"icon_url"`
	ClientID  string          `json:"client_id"`

	SupplyID  string `json:"supply_id,omitempty"`
	Inventory int    `json:"inventory,omitempty"`
}
type config struct {
	Lang      string `json:"lang"`
	Port      int    `json:"port"`
	Dev       string `json:"dev"`
	Encrypted bool   `json:"encrypted"`
	Database  struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Name     string `json:"name"`
		MaxConn  int    `json:"max_conn"`
		MinConn  int    `json:"min_conn"`
	} `json:"database"`

	Monitor struct {
		ClientID       string `json:"client_id"`
		SessionID      string `json:"session_id"`
		PrivateKey     string `json:"private_key"`
		ConversationID string `json:"conversation_id"`
	} `json:"monitor"`

	Lottery struct {
		ClientID   string                     `json:"client_id"`
		SessionID  string                     `json:"session_id"`
		PrivateKey string                     `json:"private_key"`
		PinToken   string                     `json:"pin_token"`
		PIN        string                     `json:"pin"`
		List       []Lottery                  `json:"list"`
		Rate       map[string]decimal.Decimal `json:"rate"`
	} `json:"lottery"`

	Qiniu struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		Bucket    string `json:"bucket"`
		Region    string `json:"region"`
	} `json:"qiniu,omitempty"`

	RedisAddr        string `json:"redis_addr"`
	RedisAddrReplica string `json:"redis_addr_replica"`

	ClientList     []string `json:"client_list"`
	ShowClientList []string `json:"show_client_list"`
	LuckCoinAppID  string   `json:"luck_coin_app_id"`

	FoxToken     string `json:"fox_token"`
	ExinToken    string `json:"exin_token"`
	ExinLocalKey string `json:"exin_local_key"`

	SuperManager []string `json:"super_manager"`

	Pprof map[string]string `json:"pprof"`
}

type text struct {
	Desc            string
	Join            string
	Home            string
	News            string
	Transfer        string
	Activity        string
	Auth            string
	Forward         string
	Mute            string
	Block           string
	JoinMsg         string
	AuthSuccess     string
	PrefixLeaveMsg  string
	LeaveGroup      string
	OpenChatStatus  string
	CloseChatStatus string
	MuteOpen        string
	MuteClose       string
	Muting          string
	VideoLiving     string
	VideoLiveEnd    string
	Living          string
	LiveEnd         string
	WelcomeUpdate   string
	StopMessage     string
	StopClose       string
	StopBroadcast   string
	StickerWarning  string
	StatusSet       string
	StatusCancel    string
	StatusAdmin     string
	StatusGuest     string
	Reward          string
	From            string
	MemberCentre    string
	PayForFresh     string
	PayForLarge     string
	AuthForFresh    string
	AuthForLarge    string
	LimitReject     string
	MutedReject     string
	URLReject       string
	QrcodeReject    string
	URLAdmin        string
	LanguageReject  string
	LanguageAdmin   string
	BalanceReject   string
	CategoryReject  string
	Forbid          string
	BotCard         string
	MemberTips      string
	JoinMsgInfo     string
	PINMessageError string
	Category        map[string]string
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
	} else {
		Text = en_Text
	}
	if Config.Database.MaxConn == 0 {
		Config.Database.MaxConn = 256
	}
	if Config.Database.MinConn == 0 {
		Config.Database.MinConn = 5
	}
}
