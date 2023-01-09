package message

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"mvdan.cc/xurls"
)

func checkIsMutedUser(user *models.ClientUser) bool {
	now := time.Now()
	if user.MutedAt.After(now) {
		duration := decimal.NewFromFloat(user.MutedAt.Sub(now).Hours())
		hour := duration.IntPart()
		minute := duration.Sub(decimal.NewFromInt(hour)).Mul(decimal.NewFromInt(60)).IntPart()
		go sendMutedMsg(user.ClientID, user.UserID, user.MutedTime, int(hour), int(minute))
		return true
	}
	return false
}

// 检查管理员的消息 是否 quote 了 留言消息，如果是的话，就在这个函数里处理 return true
func checkIsQuoteLeaveMessage(ctx context.Context, u *models.ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	dm, err := common.GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if dm.Status != models.DistributeMessageStatusLeaveMessage {
		return false, nil
	}
	// 确定是 quote 的留言信息了
	// 1. 看是不是 mute 和 block
	data := string(tools.Base64Decode(msg.Data))
	if strings.HasPrefix(data, "/mute") {
		muteTime := "12"
		tmp := strings.Split(data, " ")
		if len(tmp) > 1 {
			t, err := strconv.Atoi(tmp[1])
			if err == nil && t >= 0 {
				muteTime = tmp[1]
			}
		}
		if err := common.MuteClientUser(ctx, u.ClientID, dm.RepresentativeID, muteTime); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if data == "/block" {
		if err := common.BlockClientUser(ctx, u.ClientID, u.UserID, dm.RepresentativeID, false); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	// 2. 转发给其他管理员和该用户
	go handleLeaveMsg(u.ClientID, u.UserID, dm.OriginMessageID, msg)
	return true, nil
}

// 检测是否是刚刚入群5分钟内
func checkIsJustJoinGroup(u *models.ClientUser) bool {
	return u.CreatedAt.Add(time.Minute * 5).After(time.Now())
}

// 检测是否含有链接
func checkHasURLMsg(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	if msg.Category == mixin.MessageCategoryPlainImage ||
		msg.Category == "ENCRYPTED_IMAGE" {
		client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
		if err != nil {
			return false
		}
		if url, err := tools.MessageQRFilter(ctx, client.Client, msg); err == nil {
			if url != "" && !CheckUrlIsWhiteURL(ctx, clientID, url) {
				return true
			}
		} else {
			tools.Println(err)
		}
	} else if msg.Category == mixin.MessageCategoryPlainText ||
		msg.Category == "ENCRYPTED_TEXT" {
		data := string(tools.Base64Decode(msg.Data))
		// if checkHasBotID(data) {
		// 	return true
		// }
		urls := xurls.Relaxed.FindAllString(data, -1)
		for _, url := range urls {
			if !CheckUrlIsWhiteURL(ctx, clientID, url) {
				return true
			}
		}
	}
	return false
}

func CheckUrlIsWhiteURL(ctx context.Context, clientID, targetURL string) bool {
	ws, err := GetClientWhiteURLByClientID(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return false
	}
	if strings.HasPrefix(targetURL, "http") {
		targetURLObj, err := url.Parse(targetURL)
		if err != nil {
			return false
		}
		for _, w := range ws {
			if targetURLObj.Host == w {
				return true
			}
		}
	} else {
		for _, w := range ws {
			if strings.HasPrefix(targetURL, w) {
				return true
			}
		}
	}
	return false
}
func GetClientWhiteURLByClientID(ctx context.Context, clientID string) ([]string, error) {
	var result []string
	err := session.DB(ctx).Table("client_white_url").Where("client_id = ?", clientID).Pluck("white_url", &result).Error
	return result, err
}

// 检测是否达到贴纸消息的限制
func checkStickerLimit(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	var count int64
	session.DB(ctx).Table("messages").Where("client_id = ? AND user_id = ? AND category in ? AND now()-created_at<interval '5 seconds'", clientID, msg.UserID, []string{mixin.MessageCategoryPlainSticker, "ENCRYPTED_STICKER"}).Count(&count)
	if count == 2 {
		go SendStickerLimitMsg(clientID, msg.UserID)
	}
	return count >= 5
}

func SendStickerLimitMsg(clientID, userID string) {
	go common.SendClientUserTextMsg(clientID, userID, config.Text.StickerWarning, "")
}

var ignoreMsgList = []string{"Hi", "你好"}

// 检测是否是忽略的消息
func checkIsIgnoreLeaveMsg(msg *mixin.MessageView) bool {
	data := string(tools.Base64Decode(msg.Data))
	for _, s := range ignoreMsgList {
		if data == s {
			return true
		}
	}
	return false
}

var emojiRx = regexp.MustCompile(`[#*0-9]\x{FE0F}?\x{20E3}|\x{A9}\x{FE0F}?|[\x{AE}\x{203C}\x{2049}\x{2122}\x{2139}\x{2194}-\x{2199}\x{21A9}\x{21AA}]\x{FE0F}?|[\x{231A}\x{231B}]|[\x{2328}\x{23CF}]\x{FE0F}?|[\x{23E9}-\x{23EC}]|[\x{23ED}-\x{23EF}]\x{FE0F}?|\x{23F0}|[\x{23F1}\x{23F2}]\x{FE0F}?|\x{23F3}|[\x{23F8}-\x{23FA}\x{24C2}\x{25AA}\x{25AB}\x{25B6}\x{25C0}\x{25FB}\x{25FC}]\x{FE0F}?|[\x{25FD}\x{25FE}]|[\x{2600}-\x{2604}\x{260E}\x{2611}]\x{FE0F}?|[\x{2614}\x{2615}]|\x{2618}\x{FE0F}?|\x{261D}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{2620}\x{2622}\x{2623}\x{2626}\x{262A}\x{262E}\x{262F}\x{2638}-\x{263A}\x{2640}\x{2642}]\x{FE0F}?|[\x{2648}-\x{2653}]|[\x{265F}\x{2660}\x{2663}\x{2665}\x{2666}\x{2668}\x{267B}\x{267E}]\x{FE0F}?|\x{267F}|\x{2692}\x{FE0F}?|\x{2693}|[\x{2694}-\x{2697}\x{2699}\x{269B}\x{269C}\x{26A0}]\x{FE0F}?|\x{26A1}|\x{26A7}\x{FE0F}?|[\x{26AA}\x{26AB}]|[\x{26B0}\x{26B1}]\x{FE0F}?|[\x{26BD}\x{26BE}\x{26C4}\x{26C5}]|\x{26C8}\x{FE0F}?|\x{26CE}|[\x{26CF}\x{26D1}\x{26D3}]\x{FE0F}?|\x{26D4}|\x{26E9}\x{FE0F}?|\x{26EA}|[\x{26F0}\x{26F1}]\x{FE0F}?|[\x{26F2}\x{26F3}]|\x{26F4}\x{FE0F}?|\x{26F5}|[\x{26F7}\x{26F8}]\x{FE0F}?|\x{26F9}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{26FA}\x{26FD}]|\x{2702}\x{FE0F}?|\x{2705}|[\x{2708}\x{2709}]\x{FE0F}?|[\x{270A}\x{270B}][\x{1F3FB}-\x{1F3FF}]?|[\x{270C}\x{270D}][\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{270F}\x{FE0F}?|[\x{2712}\x{2714}\x{2716}\x{271D}\x{2721}]\x{FE0F}?|\x{2728}|[\x{2733}\x{2734}\x{2744}\x{2747}]\x{FE0F}?|[\x{274C}\x{274E}\x{2753}-\x{2755}\x{2757}]|\x{2763}\x{FE0F}?|\x{2764}(?:\x{200D}[\x{1F525}\x{1FA79}]|\x{FE0F}(?:\x{200D}[\x{1F525}\x{1FA79}])?)?|[\x{2795}-\x{2797}]|\x{27A1}\x{FE0F}?|[\x{27B0}\x{27BF}]|[\x{2934}\x{2935}\x{2B05}-\x{2B07}]\x{FE0F}?|[\x{2B1B}\x{2B1C}\x{2B50}\x{2B55}]|[\x{3030}\x{303D}\x{3297}\x{3299}]\x{FE0F}?|[\x{1F004}\x{1F0CF}]|[\x{1F170}\x{1F171}\x{1F17E}\x{1F17F}]\x{FE0F}?|[\x{1F18E}\x{1F191}-\x{1F19A}]|\x{1F1E6}[\x{1F1E8}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F2}\x{1F1F4}\x{1F1F6}-\x{1F1FA}\x{1F1FC}\x{1F1FD}\x{1F1FF}]|\x{1F1E7}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EF}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1E8}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1EE}\x{1F1F0}-\x{1F1F5}\x{1F1F7}\x{1F1FA}-\x{1F1FF}]|\x{1F1E9}[\x{1F1EA}\x{1F1EC}\x{1F1EF}\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1FF}]|\x{1F1EA}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1ED}\x{1F1F7}-\x{1F1FA}]|\x{1F1EB}[\x{1F1EE}-\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1F7}]|\x{1F1EC}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EE}\x{1F1F1}-\x{1F1F3}\x{1F1F5}-\x{1F1FA}\x{1F1FC}\x{1F1FE}]|\x{1F1ED}[\x{1F1F0}\x{1F1F2}\x{1F1F3}\x{1F1F7}\x{1F1F9}\x{1F1FA}]|\x{1F1EE}[\x{1F1E8}-\x{1F1EA}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}]|\x{1F1EF}[\x{1F1EA}\x{1F1F2}\x{1F1F4}\x{1F1F5}]|\x{1F1F0}[\x{1F1EA}\x{1F1EC}-\x{1F1EE}\x{1F1F2}\x{1F1F3}\x{1F1F5}\x{1F1F7}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1F1}[\x{1F1E6}-\x{1F1E8}\x{1F1EE}\x{1F1F0}\x{1F1F7}-\x{1F1FB}\x{1F1FE}]|\x{1F1F2}[\x{1F1E6}\x{1F1E8}-\x{1F1ED}\x{1F1F0}-\x{1F1FF}]|\x{1F1F3}[\x{1F1E6}\x{1F1E8}\x{1F1EA}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F4}\x{1F1F5}\x{1F1F7}\x{1F1FA}\x{1F1FF}]|\x{1F1F4}\x{1F1F2}|\x{1F1F5}[\x{1F1E6}\x{1F1EA}-\x{1F1ED}\x{1F1F0}-\x{1F1F3}\x{1F1F7}-\x{1F1F9}\x{1F1FC}\x{1F1FE}]|\x{1F1F6}\x{1F1E6}|\x{1F1F7}[\x{1F1EA}\x{1F1F4}\x{1F1F8}\x{1F1FA}\x{1F1FC}]|\x{1F1F8}[\x{1F1E6}-\x{1F1EA}\x{1F1EC}-\x{1F1F4}\x{1F1F7}-\x{1F1F9}\x{1F1FB}\x{1F1FD}-\x{1F1FF}]|\x{1F1F9}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1ED}\x{1F1EF}-\x{1F1F4}\x{1F1F7}\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FF}]|\x{1F1FA}[\x{1F1E6}\x{1F1EC}\x{1F1F2}\x{1F1F3}\x{1F1F8}\x{1F1FE}\x{1F1FF}]|\x{1F1FB}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1EE}\x{1F1F3}\x{1F1FA}]|\x{1F1FC}[\x{1F1EB}\x{1F1F8}]|\x{1F1FD}\x{1F1F0}|\x{1F1FE}[\x{1F1EA}\x{1F1F9}]|\x{1F1FF}[\x{1F1E6}\x{1F1F2}\x{1F1FC}]|\x{1F201}|\x{1F202}\x{FE0F}?|[\x{1F21A}\x{1F22F}\x{1F232}-\x{1F236}]|\x{1F237}\x{FE0F}?|[\x{1F238}-\x{1F23A}\x{1F250}\x{1F251}\x{1F300}-\x{1F320}]|[\x{1F321}\x{1F324}-\x{1F32C}]\x{FE0F}?|[\x{1F32D}-\x{1F335}]|\x{1F336}\x{FE0F}?|[\x{1F337}-\x{1F37C}]|\x{1F37D}\x{FE0F}?|[\x{1F37E}-\x{1F384}]|\x{1F385}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F386}-\x{1F393}]|[\x{1F396}\x{1F397}\x{1F399}-\x{1F39B}\x{1F39E}\x{1F39F}]\x{FE0F}?|[\x{1F3A0}-\x{1F3C1}]|\x{1F3C2}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C3}\x{1F3C4}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3C5}\x{1F3C6}]|\x{1F3C7}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C8}\x{1F3C9}]|\x{1F3CA}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CB}\x{1F3CC}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CD}\x{1F3CE}]\x{FE0F}?|[\x{1F3CF}-\x{1F3D3}]|[\x{1F3D4}-\x{1F3DF}]\x{FE0F}?|[\x{1F3E0}-\x{1F3F0}]|\x{1F3F3}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308})|\x{FE0F}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308}))?)?|\x{1F3F4}(?:\x{200D}\x{2620}\x{FE0F}?|\x{E0067}\x{E0062}(?:\x{E0065}\x{E006E}\x{E0067}|\x{E0073}\x{E0063}\x{E0074}|\x{E0077}\x{E006C}\x{E0073})\x{E007F})?|[\x{1F3F5}\x{1F3F7}]\x{FE0F}?|[\x{1F3F8}-\x{1F407}]|\x{1F408}(?:\x{200D}\x{2B1B})?|[\x{1F409}-\x{1F414}]|\x{1F415}(?:\x{200D}\x{1F9BA})?|[\x{1F416}-\x{1F43A}]|\x{1F43B}(?:\x{200D}\x{2744}\x{FE0F}?)?|[\x{1F43C}-\x{1F43E}]|\x{1F43F}\x{FE0F}?|\x{1F440}|\x{1F441}(?:\x{200D}\x{1F5E8}\x{FE0F}?|\x{FE0F}(?:\x{200D}\x{1F5E8}\x{FE0F}?)?)?|[\x{1F442}\x{1F443}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F444}\x{1F445}]|[\x{1F446}-\x{1F450}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F451}-\x{1F465}]|[\x{1F466}\x{1F467}][\x{1F3FB}-\x{1F3FF}]?|\x{1F468}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|[\x{1F468}\x{1F469}]\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F469}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?[\x{1F468}\x{1F469}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|\x{1F469}\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F46A}|[\x{1F46B}-\x{1F46D}][\x{1F3FB}-\x{1F3FF}]?|\x{1F46E}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F46F}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F470}\x{1F471}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F472}[\x{1F3FB}-\x{1F3FF}]?|\x{1F473}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F474}-\x{1F476}][\x{1F3FB}-\x{1F3FF}]?|\x{1F477}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F478}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F479}-\x{1F47B}]|\x{1F47C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F47D}-\x{1F480}]|[\x{1F481}\x{1F482}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F483}[\x{1F3FB}-\x{1F3FF}]?|\x{1F484}|\x{1F485}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F486}\x{1F487}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F488}-\x{1F48E}]|\x{1F48F}[\x{1F3FB}-\x{1F3FF}]?|\x{1F490}|\x{1F491}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F492}-\x{1F4A9}]|\x{1F4AA}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F4AB}-\x{1F4FC}]|\x{1F4FD}\x{FE0F}?|[\x{1F4FF}-\x{1F53D}]|[\x{1F549}\x{1F54A}]\x{FE0F}?|[\x{1F54B}-\x{1F54E}\x{1F550}-\x{1F567}]|[\x{1F56F}\x{1F570}\x{1F573}]\x{FE0F}?|\x{1F574}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{1F575}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F576}-\x{1F579}]\x{FE0F}?|\x{1F57A}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F587}\x{1F58A}-\x{1F58D}]\x{FE0F}?|\x{1F590}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{1F595}\x{1F596}][\x{1F3FB}-\x{1F3FF}]?|\x{1F5A4}|[\x{1F5A5}\x{1F5A8}\x{1F5B1}\x{1F5B2}\x{1F5BC}\x{1F5C2}-\x{1F5C4}\x{1F5D1}-\x{1F5D3}\x{1F5DC}-\x{1F5DE}\x{1F5E1}\x{1F5E3}\x{1F5E8}\x{1F5EF}\x{1F5F3}\x{1F5FA}]\x{FE0F}?|[\x{1F5FB}-\x{1F62D}]|\x{1F62E}(?:\x{200D}\x{1F4A8})?|[\x{1F62F}-\x{1F634}]|\x{1F635}(?:\x{200D}\x{1F4AB})?|\x{1F636}(?:\x{200D}\x{1F32B}\x{FE0F}?)?|[\x{1F637}-\x{1F644}]|[\x{1F645}-\x{1F647}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F648}-\x{1F64A}]|\x{1F64B}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F64D}\x{1F64E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F680}-\x{1F6A2}]|\x{1F6A3}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6A4}-\x{1F6B3}]|[\x{1F6B4}-\x{1F6B6}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6B7}-\x{1F6BF}]|\x{1F6C0}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6C1}-\x{1F6C5}]|\x{1F6CB}\x{FE0F}?|\x{1F6CC}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6CD}-\x{1F6CF}]\x{FE0F}?|[\x{1F6D0}-\x{1F6D2}\x{1F6D5}-\x{1F6D7}]|[\x{1F6E0}-\x{1F6E5}\x{1F6E9}]\x{FE0F}?|[\x{1F6EB}\x{1F6EC}]|[\x{1F6F0}\x{1F6F3}]\x{FE0F}?|[\x{1F6F4}-\x{1F6FC}\x{1F7E0}-\x{1F7EB}]|\x{1F90C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F90D}\x{1F90E}]|\x{1F90F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F910}-\x{1F917}]|[\x{1F918}-\x{1F91C}][\x{1F3FB}-\x{1F3FF}]?|\x{1F91D}|[\x{1F91E}\x{1F91F}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F920}-\x{1F925}]|\x{1F926}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F927}-\x{1F92F}]|[\x{1F930}-\x{1F934}][\x{1F3FB}-\x{1F3FF}]?|\x{1F935}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F936}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F937}-\x{1F939}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F93A}|\x{1F93C}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F93D}\x{1F93E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F93F}-\x{1F945}\x{1F947}-\x{1F976}]|\x{1F977}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F978}\x{1F97A}-\x{1F9B4}]|[\x{1F9B5}\x{1F9B6}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9B7}|[\x{1F9B8}\x{1F9B9}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9BA}|\x{1F9BB}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9BC}-\x{1F9CB}]|[\x{1F9CD}-\x{1F9CF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D0}|\x{1F9D1}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FC}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FE}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|[\x{1F9D2}\x{1F9D3}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9D4}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D5}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9D6}-\x{1F9DD}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F9DE}\x{1F9DF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F9E0}-\x{1F9FF}\x{1FA70}-\x{1FA74}\x{1FA78}-\x{1FA7A}\x{1FA80}-\x{1FA86}\x{1FA90}-\x{1FAA8}\x{1FAB0}-\x{1FAB6}\x{1FAC0}-\x{1FAC2}\x{1FAD0}-\x{1FAD6}]|\,|\.|\?|\<|\>|\/|\;|\:|\'|\"|\[|\{|\]|\}|\!|\@|\#|\$|\%|\^|\&|\*|\(|\)|\_|\+|\-|\=|\~|\ |，|《|。|》|？|；|：|、|！|¥|…|（|）|—|【|】|｜|｛|｝|～|1|2|3|4|5|6|7|8|9|0`)

// 语言检测
func checkMsgLanguage(msg *mixin.MessageView, clientID string) bool {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false
	}
	lang := config.Config.Lang
	if lang == "zh" {
		return false
	}
	c, err := common.GetClientByIDOrHost(models.Ctx, clientID)
	if err != nil {
		tools.Println(err)
		return false
	}
	if c.Lang == "zh" {
		return false
	}
	data := string(emojiRx.ReplaceAllString(string(tools.Base64Decode(msg.Data)), ``))
	if len(data) == 0 {
		return false
	}
	return languageRateCheck(data, lang)
}

func languageRateCheck(data, lang string) bool {
	var t *unicode.RangeTable
	switch lang {
	case "en":
		t = nil
	case "zh":
		t = new(unicode.RangeTable)
	}
	langPer := tools.LanguageCount(data, t)
	return langPer.LessThan(config.LangCheckPer)
}

var forbiddenMsgCategory = map[string]bool{
	mixin.MessageCategoryPlainAudio:     true,
	"ENCRYPTED_AUDIO":                   true,
	mixin.MessageCategoryPlainLocation:  true,
	"ENCRYPTED_LOCATION":                true,
	mixin.MessageCategoryAppButtonGroup: true,
}

// 单独检测 禁止发的消息类型 这三种消息不能发。
func checkMsgIsForbid(u *models.ClientUser, msg *mixin.MessageView) bool {
	ctx := models.Ctx
	if forbiddenMsgCategory[msg.Category] {
		// 发送禁止消息
		go common.SendForbidMsg(u.ClientID, u.UserID, msg.Category)
		return true
	}

	if msg.Category == mixin.MessageCategoryPlainContact ||
		msg.Category == "ENCRYPTED_CONTACT" {
		data := tools.Base64Decode(msg.Data)
		var c mixin.ContactMessage
		if err := json.Unmarshal(data, &c); err != nil {
			return true
		}
		contactUser, err := common.SearchUser(ctx, u.ClientID, c.UserID)
		if err != nil {
			return true
		}
		id, _ := strconv.Atoi(contactUser.IdentityNumber)
		if id < 7000000000 {
			// 联系人卡片消息
			go sendForbidMsg(u.ClientID, u.UserID, msg.Category)
			return true
		}
	}

	return false
}

// 检查消息频率
func checkMessageCountLimit(ctx context.Context, clientID, userID string, status int) bool {
	var count int64
	if err := session.DB(ctx).Table("messages").Where("client_id=? AND user_id=? AND now()-created_at<interval '1 minutes'", clientID, userID).Count(&count).Error; err != nil {
		return false
	}
	limit := statusLimitMap[status]
	return count < limit
}

// 检查用户是否可以发送目标的消息类型
func checkCategory(ctx context.Context, clientID, category string, status int) bool {
	if category == mixin.MessageCategoryMessageRecall ||
		status == models.ClientUserStatusAdmin ||
		status == models.ClientUserStatusGuest {
		return true
	}
	return common.CheckHasClientMemberAuth(ctx, clientID, category, status)
}

var cacheSendJoinMsg = tools.NewMutex()

// 检测用户是否5分钟内发过消息
func checkIsSendJoinMsg(userID string) bool {
	t := cacheSendJoinMsg.Read(userID)
	if t == nil {
		cacheSendJoinMsg.Write(userID, time.Now())
		return false
	}
	if t.(time.Time).Add(time.Minute * 5).Before(time.Now()) {
		cacheSendJoinMsg.Write(userID, time.Now())
		return false
	}
	return true
}

// 检查 conversation 是否是会话
func checkIsContact(ctx context.Context, clientID, conversationID string) bool {
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return false
	}
	c, err := client.ReadConversation(ctx, conversationID)
	if err != nil {
		tools.Println(err)
		return false
	}
	return c.Category == mixin.ConversationCategoryContact
}

// 检测是否能够发送红包
func checkCanNotSendLuckyCoin(ctx context.Context, clientID, data, status string) bool {
	var m mixin.AppCardMessage
	err := json.Unmarshal(tools.Base64Decode(data), &m)
	if err != nil {
		tools.Println(err)
		return true
	}
	u, err := url.Parse(m.Action)
	if err != nil {
		tools.Println(err)
		return true
	}
	query, _ := url.ParseQuery(u.RawQuery)
	if len(query["uid"]) == 0 {
		return true
	}
	uid := query["uid"][0]
	if common.CheckIsBlockUser(ctx, clientID, uid) {
		return true
	}
	user, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, uid)
	if err != nil || user.UserID == "" {
		tools.Println(err, user)
		return true
	}
	if !common.CheckHasClientMemberAuth(ctx, clientID, "lucky_coin", user.Status) {
		return true
	}
	if (status == models.ClientConversationStatusMute ||
		status == models.ClientConversationStatusAudioLive) &&
		!common.CheckIsAdmin(ctx, clientID, uid) {
		return true
	}

	return false
}

// 检查 是否是 帮转/禁言/拉黑 的消息
func checkIsButtonOperation(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if !strings.HasPrefix(data, "---operation") {
		return false, nil
	}
	// 确定是操作的内容了
	operationAction := strings.Split(data, ",")
	if len(operationAction) != 3 {
		return true, nil
	}
	originMsg, err := getMsgByClientIDAndMessageID(ctx, clientID, operationAction[2])
	if err != nil {
		return true, err
	}
	switch operationAction[1] {
	// 1. 帮转发
	case "forward":
		if err := common.CreateMessage(ctx, clientID, &mixin.MessageView{
			ConversationID: originMsg.ConversationID,
			UserID:         originMsg.UserID,
			MessageID:      originMsg.MessageID,
			Category:       originMsg.Category,
			Data:           originMsg.Data,
			CreatedAt:      msg.CreatedAt,
		}, models.MessageStatusPending); err != nil {
			return true, err
		}
	// 2. 禁言
	case "mute":
		if err := common.MuteClientUser(ctx, clientID, originMsg.UserID, "12"); err != nil {
			tools.Println(err)
		}
	// 3. 拉黑
	case "block":
		if err := common.BlockClientUser(ctx, clientID, msg.UserID, originMsg.UserID, false); err != nil {
			tools.Println(err)
		}
	}

	return true, nil
}

func checkIsOperationMsg(ctx context.Context, u *models.ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if data == "/mute open" || data == "/mute close" {
		muteStatus := data == "/mute open"
		common.MuteClientOperation(muteStatus, u.ClientID)
		return true, nil
	}
	if isOperation, err := handleUnmuteAndUnblockMsg(ctx, data, u); err != nil {
		tools.Println(err)
	} else if isOperation {
		return true, nil
	}

	return handleRecallOrMuteOrBlockOrInfoMsg(ctx, data, u.ClientID, msg)
}
func handleUnmuteAndUnblockMsg(ctx context.Context, data string, u *models.ClientUser) (bool, error) {
	operation := strings.Split(data, " ")
	if len(operation) < 2 || len(operation[1]) <= 4 {
		return false, nil
	}
	if strings.HasPrefix(data, "/unmute") {
		_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := common.MuteClientUser(ctx, u.ClientID, _u.UserID, "0"); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/unblock") {
		_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := common.BlockClientUser(ctx, u.ClientID, u.UserID, _u.UserID, true); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/blockall") {
		if checkIsSuperManager(u.UserID) {
			_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
			if err != nil {
				tools.Println(err)
				return true, nil
			}
			memo := ""
			if len(operation) == 3 {
				memo = operation[2]
			}
			if err := common.AddBlockUser(ctx, u.UserID, u.ClientID, _u.UserID, memo); err != nil {
				tools.Println(err)
			}
			common.SendClientUserTextMsg(u.ClientID, u.UserID, "success", "")
		}
		return true, nil
	}
	return false, nil
}

func checkIsSuperManager(userID string) bool {
	for _, v := range config.Config.SuperManager {
		if v == userID {
			return true
		}
	}
	return false
}
func handleRecallOrMuteOrBlockOrInfoMsg(ctx context.Context, data, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	if data != "/info" && data != "ban" && data != "kick" && data != "delete" && data != "/recall" && data != "/block" && !strings.HasPrefix(data, "/mute") {
		return false, nil
	}
	dm, err := common.GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
	if err != nil {
		return true, err
	}
	m, err := getMsgByClientIDAndMessageID(ctx, clientID, dm.OriginMessageID)
	if err != nil {
		tools.Println(err)
		return true, err
	}
	if data == "/recall" || data == "delete" {
		if err := common.CreatedManagerRecallMsg(ctx, clientID, dm.OriginMessageID, m.UserID); err != nil {
			return true, err
		}
	}
	// 针对用户的操作
	if data == "/info" {
		common.CheckAndReplaceProxyUser(ctx, clientID, &m.UserID)
		objData := map[string]string{"user_id": m.UserID}
		byteData, _ := json.Marshal(objData)
		client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
		if err != nil {
			return true, err
		}
		go common.SendMessage(models.Ctx, client.Client, &mixin.MessageRequest{
			ConversationID: msg.ConversationID,
			RecipientID:    msg.RepresentativeID,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryPlainContact,
			Data:           tools.Base64Encode(byteData),
		}, false)
		return true, nil
	}
	if strings.HasPrefix(data, "/mute") || data == "kick" {
		muteTime := "12"
		tmp := strings.Split(data, " ")
		if len(tmp) > 1 {
			t, err := strconv.Atoi(tmp[1])
			if err == nil && t >= 0 {
				muteTime = tmp[1]
			}
		}
		if err := common.MuteClientUser(ctx, clientID, m.UserID, muteTime); err != nil {
			return true, err
		}
	}
	if data == "/block" || data == "ban" {
		if err := common.BlockClientUser(ctx, clientID, msg.UserID, m.UserID, false); err != nil {
			return true, err
		}
	}
	return true, nil
}
