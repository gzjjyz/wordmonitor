/**
 * @Desc: 超神觉醒（WebGame SDK 2.0）屏蔽字检测 + 聊天监控
 **/

package wordmonitor

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/bitly/go-simplejson"
)

// doc:WebGame SDK 2.0 集成文档
// 3.聊天监控 /api/v2/webgame/chat-monitor
// 6.屏蔽词检测 /api/v2/webgame/text-check
const (
	_csjxDomain = "https://sdkapi.qkyx.online"

	_csjxTextCheckApiUrl   = _csjxDomain + "/api/v2/webgame/text-check"
	_csjxChatMonitorApiUrl = _csjxDomain + "/api/v2/webgame/chat-monitor"
)

// 屏蔽词检测场景 scene
const (
	SceneCsjxByNickname    = "nickname"     // 角色取名
	SceneCsjxByGuildName   = "guild_name"   // 公会取名
	SceneCsjxByGuildNotice = "guild_notice" // 公会公告
	SceneCsjxByChat        = "chat"         // 聊天发言（默认）
)

// 聊天监控频道 channel
const (
	ChannelCsjxByWorld     = "世界"
	ChannelCsjxByNation    = "国家"
	ChannelCsjxByGuild     = "公会"
	ChannelCsjxByFaction   = "帮会"
	ChannelCsjxByTeam      = "队伍"
	ChannelCsjxByPrivate   = "私聊"
	ChannelCsjxByBroadcast = "喇叭"
	ChannelCsjxByMail      = "邮件"
)

type _CsjxMonitor struct {
	GameId     string
	LoginKey   string
	ChannelMap map[uint32]string
}

func NewCsjxMonitor(gameId, loginKey string, channelMap map[uint32]string) *_CsjxMonitor {
	return &_CsjxMonitor{
		GameId:     gameId,
		LoginKey:   loginKey,
		ChannelMap: channelMap,
	}
}

func (m *_CsjxMonitor) makeSign(source string) string {
	h := md5.New()
	io.WriteString(h, source)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// textCheck 屏蔽词检测，platform 取本次校验数据里的平台标识，sign = md5(game_id + platform + scene + text + timestamp + login_key)
func (m *_CsjxMonitor) textCheck(platform, scene, content string) (Ret, error) {
	nowSec := time.Now().Unix()
	formData := map[string]string{
		"platform":  platform,
		"game_id":   m.GameId,
		"text":      content,
		"scene":     scene,
		"timestamp": fmt.Sprintf("%d", nowSec),
		"sign": m.makeSign(fmt.Sprintf("%s%s%s%s%d%s", m.GameId, platform, scene, content,
			nowSec, m.LoginKey)),
	}

	response, err := GetRestyClient().R().
		SetFormData(formData).
		Post(_csjxTextCheckApiUrl)
	if err != nil {
		return Failed, err
	}

	retJson, err := simplejson.NewJson(response.Body())
	if err != nil {
		return Failed, err
	}
	if code := retJson.Get("code").MustInt(0); code != 1 {
		return Failed, fmt.Errorf("屏蔽字检测请求失败:code:%d,msg:%s", code, retJson.Get("msg").MustString())
	}

	data := retJson.Get("data")
	if data.Get("passed").MustBool() {
		return Success, nil
	}
	return Failed, fmt.Errorf("检测不通过:content:%s,msg:%v", content,
		retJson.Get("msg").MustString())
}

// chatMonitor 聊天监控，sign = md5(user_id + game_id + timestamp + login_key)
func (m *_CsjxMonitor) chatMonitor(data *CommonData) (Ret, error) {
	nowSec := time.Now().Unix()
	userId := GetPlatformUid(data.PlatformUniquePlayerId)

	channel := ChannelCsjxByWorld
	if m.ChannelMap != nil {
		if val, ok := m.ChannelMap[data.ChatChannel]; ok && val != "" {
			channel = val
		}
	}

	formData := map[string]string{
		"user_id":      userId,
		"game_id":      m.GameId,
		"timestamp":    fmt.Sprintf("%d", nowSec),
		"sign":         m.makeSign(fmt.Sprintf("%s%s%d%s", userId, m.GameId, nowSec, m.LoginKey)),
		"chat_content": data.Content,
		"channel":      channel,
		"role_id":      fmt.Sprintf("%d", data.ActorId),
		"role_name":    data.ActorName,
		"server_id":    fmt.Sprintf("%d", data.SrvId),
		"account_name": userId,
		"chat_time":    fmt.Sprintf("%d", nowSec),
		"ip_addr":      data.ActorIP,
	}
	if data.TargetActorId != 0 || data.PlatformUniqueTargetPlayerId != "" {
		formData["sec_chat_user_id"] = GetPlatformUid(data.PlatformUniqueTargetPlayerId)
		formData["sec_chat_role_id"] = fmt.Sprintf("%d", data.TargetActorId)
		formData["sec_chat_nickname"] = data.TargetActorName
	}

	response, err := GetRestyClient().R().
		SetFormData(formData).
		Post(_csjxChatMonitorApiUrl)
	if err != nil {
		return Failed, err
	}

	retJson, err := simplejson.NewJson(response.Body())
	if err != nil {
		return Failed, err
	}
	if code := retJson.Get("code").MustInt(0); code != 1 {
		return Failed, fmt.Errorf("聊天监控请求失败:code:%d,msg:%s", code, retJson.Get("msg").MustString())
	}

	result := retJson.Get("data").Get("platform_result")
	if result.Get("pass").MustBool() {
		return Success, nil
	}
	return Failed, fmt.Errorf("检测不通过:content:%s,message:%s", data.Content,
		result.Get("message").MustString())
}

var _csjxNameSceneRef = map[uint32]string{
	NameSourceRole:        SceneCsjxByNickname,
	NameSourceGuild:       SceneCsjxByGuildName,
	NameSourceGuildNotice: SceneCsjxByGuildNotice,
}

func (m *_CsjxMonitor) CheckName(data *CommonData) (Ret, error) {
	scene := SceneCsjxByNickname
	if val, ok := _csjxNameSceneRef[data.NameSource]; ok {
		scene = val
	}
	return m.textCheck(data.Platform, scene, data.Content)
}

func (m *_CsjxMonitor) CheckChat(data *CommonData) (Ret, error) {
	ret, err := m.textCheck(data.Platform, SceneCsjxByChat, data.Content)
	if err != nil || ret != Success {
		return ret, err
	}
	return m.chatMonitor(data)
}

func (m *_CsjxMonitor) SetNameBusinessId(id string) {
}

func (m *_CsjxMonitor) SetChatBusinessId(id string) {
}

func (m *_CsjxMonitor) ClearCache() {
}
