/**
 * @Author: lzp
 * @Date: 2025/9/29
 * @Desc:
**/

package wordmonitor

import (
	"crypto/md5"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"strconv"
	"strings"
	"time"
)

// doc:https://docs-game.flash.cn/flashcenter_webgame_doc/webgame_push_doc
// 2144
const (
	_2144ApiUrl = "https://trackcp-game.flash.cn/vendor/check-chat-v2/"

	ChatType2144ByPrivate   = 1  // 1 私聊；
	ChatType2144ByBroadcast = 2  // 2 喇叭；
	ChatType2144ByMail      = 3  // 3 邮件；
	ChatType2144ByWorld     = 4  // 4 世界；
	ChatType2144ByNation    = 5  // 5 国家；
	ChatType2144ByGuild     = 6  // 6 工会/帮会；
	ChatType2144ByTeam      = 7  // 7 队伍；
	ChatType2144ByNear      = 8  // 8 附近；
	ChatType2144ByOther     = 9  // 9 其他;
	ChatType2144ByName      = 10 // 10 昵称(需要玩家在创建角色的时候，检测昵称是否合规)；
	ChatType2144ByNotice    = 11 // 11 公告

)

type _2144WanMonitor struct {
	GKey       string
	ChannelMap map[int]int
	LoginKey   string
}

type _2144WanMonitorReq struct {
	GKey     string `json:"gkey"`
	ServerId string `json:"server_id"`
	QId      uint64 `json:"qid"`
	Name     string `json:"name"`
	Type     uint32 `json:"type"`
	ToQid    uint64 `json:"toqid"`
	ToName   string `json:"toname"`
	RoleId   string `json:"roleid"`
	Content  string `json:"content"`
	IP       string `json:"ip"`
	LoginKey string `json:"login_key"`
}

func New2144WanMonitor(gkey, loginKey string, channelMap map[int]int) *_2144WanMonitor {
	return &_2144WanMonitor{
		GKey:       gkey,
		ChannelMap: channelMap,
		LoginKey:   loginKey,
	}
}

func (r *_2144WanMonitorReq) ToFormData(timeSec int64) map[string]string {
	return map[string]string{
		"gkey":      r.GKey,
		"server_id": r.ServerId,
		"qid":       fmt.Sprintf("%d", r.QId),
		"name":      r.Name,
		"type":      fmt.Sprintf("%d", r.Type),
		"toqid":     fmt.Sprintf("%d", r.ToQid),
		"toname":    r.ToName,
		"roleid":    r.RoleId,
		"content":   r.Content,
		"time":      fmt.Sprintf("%d", timeSec),
		"ip":        r.IP,
		"retint":    fmt.Sprintf("%d", 1),
		"sign":      r.MakeSign(timeSec),
	}
}

func (r *_2144WanMonitorReq) MakeSign(timeSec int64) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s%d%s%d%d%s%s%d%s%s", r.GKey, r.ServerId, r.QId, r.Name, r.Type, r.ToQid, r.ToName, r.Content, timeSec, r.IP, r.LoginKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_2144WanMonitor) check(req *_2144WanMonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post(_2144ApiUrl)
	if err != nil {
		return
	}
	code, _ := strconv.Atoi(string(response.Body()))
	switch code {
	case 1:
		result = Success
	case 4:
		err = fmt.Errorf("签名错误")
	case 5:
		err = fmt.Errorf("其他错误")
	default:
		err = fmt.Errorf("检测不通过")
	}
	return
}

func (m *_2144WanMonitor) CheckName(data *CommonData) (Ret, error) {
	var platformUniquePlayerId, platformUniqueTargetPlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	split = strings.Split(data.PlatformUniqueTargetPlayerId, "_")
	if len(split) > 0 {
		platformUniqueTargetPlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	ret, err := m.check(&_2144WanMonitorReq{
		GKey:     m.GKey,
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      uint64(platformUniquePlayerId),
		Name:     data.ActorName,
		Type:     ChatType2144ByName,
		ToQid:    uint64(platformUniqueTargetPlayerId),
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.ActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: m.LoginKey,
	})
	return ret, err
}

func (m *_2144WanMonitor) CheckChat(data *CommonData) (Ret, error) {
	var platformUniquePlayerId, platformUniqueTargetPlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	split = strings.Split(data.PlatformUniqueTargetPlayerId, "_")
	if len(split) > 0 {
		platformUniqueTargetPlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	var chatType = uint32(ChatType2144ByWorld)
	if m.ChannelMap != nil {
		val, ok := m.ChannelMap[int(data.ChatChannel)]
		if ok {
			chatType = uint32(val)
		}
	}
	ret, err := m.check(&_2144WanMonitorReq{
		GKey:     m.GKey,
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      uint64(platformUniquePlayerId),
		Name:     data.ActorName,
		Type:     chatType,
		ToQid:    uint64(platformUniqueTargetPlayerId),
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.ActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: m.LoginKey,
	})
	return ret, err
}

func (m *_2144WanMonitor) SetNameBusinessId(id string) {
}

func (m *_2144WanMonitor) SetChatBusinessId(id string) {
}

func (m *_2144WanMonitor) ClearCache() {
}
