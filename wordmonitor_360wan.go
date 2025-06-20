package wordmonitor

import (
	"crypto/md5"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// doc:https://easydoc.soft.360.cn/doc?project=94c0311bf73c33c7f41fd0b6b205d0f5&doc=9a5638a8e8e186f233e6540f55c4eb3f&config=title_menu_toc#h1-%E4%B8%89.%20%E9%98%B2%E6%8B%89%E4%BA%BA[%E5%BF%85%E5%81%9A]
// 360 戮仙战纪
const (
	_360ApiUrl = "http://game.api.1360.com/realchat"

	ChatType360ByPrivate   = 1  // 1 私聊；
	ChatType360ByBroadcast = 2  // 2 喇叭；
	ChatType360ByMail      = 3  // 3 邮件；
	ChatType360ByWorld     = 4  // 4 世界；
	ChatType360ByNation    = 5  // 5 国家；
	ChatType360ByGuild     = 6  // 6 工会/帮会；
	ChatType360ByTeam      = 7  // 7 队伍；
	ChatType360ByNear      = 8  // 8 附近；
	ChatType360ByOther     = 9  // 9 其他;
	ChatType360ByName      = 10 // 10 昵称(需要玩家在创建角色的时候，检测昵称是否合规)；
	ChatType360ByNotice    = 11 // 11 公告

)

type _360WanMonitor struct {
	GKey       string
	ChannelMap map[int]int
	LoginKey   string
}

type _360WanMonitorReq struct {
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

func New360WanMonitor(gkey, loginKey string, channelMap map[int]int) *_360WanMonitor {
	return &_360WanMonitor{
		GKey:       gkey,
		ChannelMap: channelMap,
		LoginKey:   loginKey,
	}
}

func (r *_360WanMonitorReq) ToFormData(timeSec int64) map[string]string {
	r.Name = url.QueryEscape(r.Name)
	r.ToName = url.QueryEscape(r.ToName)
	r.Content = url.QueryEscape(r.Content)

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

func (r *_360WanMonitorReq) MakeSign(timeSec int64) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s%d%s%d%d%s%s%d%s%s", r.GKey, r.ServerId, r.QId, r.Name, r.Type, r.ToQid, r.ToName, r.Content, timeSec, r.IP, r.LoginKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_360WanMonitor) check(req *_360WanMonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post(_360ApiUrl)
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

func (m *_360WanMonitor) CheckName(data *CommonData) (Ret, error) {
	var platformUniquePlayerId, platformUniqueTargetPlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	split = strings.Split(data.PlatformUniqueTargetPlayerId, "_")
	if len(split) > 0 {
		platformUniqueTargetPlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	ret, err := m.check(&_360WanMonitorReq{
		GKey:     m.GKey,
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      uint64(platformUniquePlayerId),
		Name:     data.ActorName,
		Type:     ChatType360ByName,
		ToQid:    uint64(platformUniqueTargetPlayerId),
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.TargetActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: m.LoginKey,
	})
	return ret, err
}

func (m *_360WanMonitor) CheckChat(data *CommonData) (Ret, error) {
	var platformUniquePlayerId, platformUniqueTargetPlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}
	split = strings.Split(data.PlatformUniqueTargetPlayerId, "_")
	if len(split) > 0 {
		platformUniqueTargetPlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	var chatType = uint32(ChatType360ByWorld)
	if m.ChannelMap != nil {
		val, ok := m.ChannelMap[int(data.ChatChannel)]
		if ok {
			chatType = uint32(val)
		}
	}
	ret, err := m.check(&_360WanMonitorReq{
		GKey:     m.GKey,
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      uint64(platformUniquePlayerId),
		Name:     data.ActorName,
		Type:     chatType,
		ToQid:    uint64(platformUniqueTargetPlayerId),
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.TargetActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: m.LoginKey,
	})
	return ret, err
}

func (m *_360WanMonitor) SetNameBusinessId(id string) {
}

func (m *_360WanMonitor) SetChatBusinessId(id string) {
}

func (m *_360WanMonitor) ClearCache() {
}
