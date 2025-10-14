/**
 * @Author: lzp
 * @Date: 2025/9/29
 * @Desc:
**/

package wordmonitor

import (
	"crypto/md5"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"io"
	"strconv"
	"strings"
	"time"
)

// doc:https://api.9hou.com/e/member/mixed/doc/api_game_cp.php#chat_report_list
// 9hou
const (
	_9HouApiUrl = "http://api.9hou.com/api/chat_report/"

	ChatType9HouByPrivate     = 1 // 1 私聊；
	ChatType9HouByBroadcast   = 2 // 2 喇叭；
	ChatType9HouByFuBen       = 3 // 3 副本；
	ChatType9HouByWorld       = 4 // 4 世界；
	ChatType9HouByCross       = 5 // 5 跨服；
	ChatType9HouByGuild       = 6 // 6 工会/帮会；
	ChatType9HouByTeam        = 7 // 7 队伍；
	ChatType9HouByTransaction = 8 // 8 交易；
	ChatType9HouByOther       = 9 // 9 附近;
)

type _9HouMonitor struct {
	Gid        uint32
	Type       uint32
	LoginKey   string
	ChannelMap map[int]int
}

type _9HouMonitorReq struct {
	Type      uint64 `json:"type"`
	Uid       uint64 `json:"uid"`
	Gid       uint64 `json:"gid"`
	Sid       uint64 `json:"sid"`
	URoleName string `json:"urole_name"`
	URoleId   uint64 `json:"urole_id"`
	Channel   uint64 `json:"channel"`
	Tid       uint64 `json:"tid"`
	TRoleName string `json:"trole_name"`
	TRoleId   uint64 `json:"trole_id"`
	ChatBody  string `json:"chat_body"`
	LoginKey  string `json:"login_key"`
}

func New9HouMonitor(gid, typ uint32, loginKey string, channelMap map[int]int) *_9HouMonitor {
	return &_9HouMonitor{
		Gid:        gid,
		Type:       typ,
		LoginKey:   loginKey,
		ChannelMap: channelMap,
	}
}

func (r *_9HouMonitorReq) ToFormData(timeSec int64) map[string]string {
	return map[string]string{
		"type":       fmt.Sprintf("%d", r.Type),
		"uid":        fmt.Sprintf("%d", r.Uid),
		"gid":        fmt.Sprintf("%d", r.Gid),
		"sid":        fmt.Sprintf("%d", r.Sid),
		"urole_name": r.URoleName,
		"urole_id":   fmt.Sprintf("%d", r.URoleId),
		"channel":    fmt.Sprintf("%d", r.Channel),
		"tid":        fmt.Sprintf("%d", r.Tid),
		"trole_name": r.TRoleName,
		"trole_id":   fmt.Sprintf("%d", r.TRoleId),
		"chat_body":  r.ChatBody,
		"time":       fmt.Sprintf("%d", timeSec),
		"sign":       r.MakeSign(timeSec),
		"api":        "ban_keywords",
	}
}

func (r *_9HouMonitorReq) MakeSign(timeSec int64) string {
	// 8810
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%d%d%s%d%d%s", r.Uid, r.Gid, r.Sid, r.URoleName, r.Channel, timeSec, r.LoginKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_9HouMonitor) check(req *_9HouMonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post(_9HouApiUrl)
	if err != nil {
		return
	}
	body := response.Body()

	retJson, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, _ := retJson.Get("is_ban_words").String()
	content, _ := retJson.Get("msg").String()
	if strings.EqualFold(code, "1") {
		err = fmt.Errorf("检测不通过 %s", content)
	} else {
		result = Success
	}
	return
}

func (m *_9HouMonitor) CheckName(data *CommonData) (Ret, error) {
	// 不提供校验取名 默认不通过
	return Failed, nil
}

func (m *_9HouMonitor) CheckChat(data *CommonData) (Ret, error) {
	var platformUniquePlayerId, platformUniqueTargetPlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	split = strings.Split(data.PlatformUniqueTargetPlayerId, "_")
	if len(split) > 0 {
		platformUniqueTargetPlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	var chatType = uint32(ChatType9HouByWorld)
	if m.ChannelMap != nil {
		val, ok := m.ChannelMap[int(data.ChatChannel)]
		if ok {
			chatType = uint32(val)
		}
	}
	ret, err := m.check(&_9HouMonitorReq{
		Type:      uint64(m.Type),
		Uid:       uint64(platformUniquePlayerId),
		Gid:       uint64(m.Gid),
		Sid:       uint64(data.SrvId),
		URoleName: data.ActorName,
		URoleId:   data.ActorId,
		Channel:   uint64(chatType),
		Tid:       data.TargetActorId,
		TRoleName: data.TargetActorName,
		TRoleId:   uint64(platformUniqueTargetPlayerId),
		ChatBody:  data.Content,
		LoginKey:  m.LoginKey,
	})
	return ret, err
}

func (m *_9HouMonitor) SetNameBusinessId(id string) {
}

func (m *_9HouMonitor) SetChatBusinessId(id string) {
}

func (m *_9HouMonitor) ClearCache() {
}
