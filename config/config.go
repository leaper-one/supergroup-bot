package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

const (
	BuildVersion = "BUILD_VERSION"
)

var (
	MessageShardSize     = int64(16)
	MessageShardModifier = "SHARD"
	CacheTime            = 15 * time.Minute
	DebounceTime         = 1 * time.Minute
	AssetsCheckTime      = 12 * time.Hour
	NotActiveCheckTime   = 7 * 24.0
	//CacheTime          = 1 * time.Minute
	//DebounceTime       = 1 * time.Second
	//AssetsCheckTime    = 5 * time.Minute
	//NotActiveCheckTime = 3.0
)

var (
	DatabaseUser     string
	DatabasePassword string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
)

var (
	ClientList      []string
	AvoidClientList []string
)

var (
	JoinBtn1       string
	JoinBtn2       string
	WelBtn1        string
	WelBtn2        string
	TransferBtn    string
	WelBtn4        string
	AuthBtn        string
	Forward        string
	Mute           string
	Block          string
	JoinMsg        string
	AuthSuccess    string
	PrefixLeaveMsg string
	LuckCoinAppID  string

	LeaveGroup      string
	OpenChatStatus  string
	CloseChatStatus string

	Category = make(map[string]string)
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DatabaseUser = os.Getenv("DatabaseUser")
	DatabasePassword = os.Getenv("DatabasePassword")
	DatabaseHost = os.Getenv("DatabaseHost")
	DatabasePort = os.Getenv("DatabasePort")
	DatabaseName = os.Getenv("DatabaseName")

	clientList := os.Getenv("ClientList")
	ClientList = strings.Split(clientList, ",")

	avoidClientList := os.Getenv("AvoidClientList")
	AvoidClientList = strings.Split(avoidClientList, ",")

	JoinBtn1 = os.Getenv("JoinBtn1")
	JoinBtn2 = os.Getenv("JoinBtn2")
	WelBtn1 = os.Getenv("WelBtn1")
	WelBtn2 = os.Getenv("WelBtn2")
	TransferBtn = os.Getenv("TransferBtn")
	WelBtn4 = os.Getenv("WelBtn4")
	AuthBtn = os.Getenv("AuthBtn")
	Forward = os.Getenv("Forward")
	Mute = os.Getenv("Mute")
	Block = os.Getenv("Block")
	JoinMsg = os.Getenv("JoinMsg")
	AuthSuccess = os.Getenv("AuthSuccess")
	PrefixLeaveMsg = os.Getenv("PrefixLeaveMsg")
	LuckCoinAppID = os.Getenv("LuckCoinAppID")
	LeaveGroup = os.Getenv("LeaveGroup")
	OpenChatStatus = os.Getenv("OpenChatStatus")
	CloseChatStatus = os.Getenv("CloseChatStatus")

	Category["PLAIN_TEXT"] = os.Getenv("PLAIN_TEXT")
	Category["PLAIN_POST"] = os.Getenv("PLAIN_POST")
	Category["PLAIN_IMAGE"] = os.Getenv("PLAIN_IMAGE")
	Category["PLAIN_STICKER"] = os.Getenv("PLAIN_STICKER")
	Category["PLAIN_LIVE"] = os.Getenv("PLAIN_LIVE")
	Category["PLAIN_VIDEO"] = os.Getenv("PLAIN_VIDEO")
	Category["APP_CARD"] = os.Getenv("APP_CARD")
	Category["PLAIN_LOCATION"] = os.Getenv("PLAIN_LOCATION")
	Category["PLAIN_DATA"] = os.Getenv("PLAIN_DATA")
	Category["PLAIN_CONTACT"] = os.Getenv("PLAIN_CONTACT")
	Category["PLAIN_VIDEO"] = os.Getenv("PLAIN_VIDEO")
	Category["PLAIN_AUDIO"] = os.Getenv("PLAIN_AUDIO")
}
