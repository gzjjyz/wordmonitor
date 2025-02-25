package wordmonitor

import (
	"crypto/md5"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"strconv"
	"time"
)

// doc:https://easydoc.soft.360.cn/doc?project=94c0311bf73c33c7f41fd0b6b205d0f5&doc=9a5638a8e8e186f233e6540f55c4eb3f&config=title_menu_toc#h1-%E4%B8%89.%20%E9%98%B2%E6%8B%89%E4%BA%BA[%E5%BF%85%E5%81%9A]
// 360 戮仙战纪
const (
	_360ApiUrl = "http://game.api.1360.com/realchat"

	_360ChatTypeByPrivate   = 1  // 1 私聊；
	_360ChatTypeByBroadcast = 2  // 2 喇叭；
	_360ChatTypeByMail      = 3  // 3 邮件；
	_360ChatTypeByWorld     = 4  // 4 世界；
	_360ChatTypeByNation    = 5  // 5 国家；
	_360ChatTypeByGuild     = 6  // 6 工会/帮会；
	_360ChatTypeByTeam      = 7  // 7 队伍；
	_360ChatTypeByNear      = 8  // 8 附近；
	_360ChatTypeByOther     = 9  // 9 其他;
	_360ChatTypeByName      = 10 // 10 昵称(需要玩家在创建角色的时候，检测昵称是否合规)；
	_360ChatTypeByNotice    = 11 // 11 公告

)

type _360Monitor struct {
	GKey       string
	ChannelMap map[int]int
}

type _360MonitorReq struct {
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

func New360Monitor(gkey string, channelMap map[int]int) *_360Monitor {
	return &_360Monitor{
		GKey:       gkey,
		ChannelMap: channelMap,
	}
}

func (r *_360MonitorReq) ToFormData(timeSec int64) map[string]string {
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

func (r *_360MonitorReq) MakeSign(timeSec int64) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s%d%s%d%d%s%s%d%s%s", "", r.ServerId, r.QId, r.Name, r.Type, r.ToQid, r.ToName, r.Content, timeSec, r.IP, r.LoginKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_360Monitor) build() {

}

func (m *_360Monitor) check(req *_360MonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post("http://game.api.1360.com/realchat")
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

func (m *_360Monitor) CheckName(data *CommonData) (Ret, error) {
	ret, err := m.check(&_360MonitorReq{
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      data.PlatformUniquePlayerId,
		Name:     data.ActorName,
		Type:     _360ChatTypeByName,
		ToQid:    data.PlatformUniqueTargetPlayerId,
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.TargetActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: data.PlatformLoginKey,
	})
	return ret, err
}

func (m *_360Monitor) CheckChat(data *CommonData) (Ret, error) {
	var chatType = uint32(_360ChatTypeByWorld)
	if m.ChannelMap != nil {
		val, ok := m.ChannelMap[int(data.ChatChannel)]
		if ok {
			chatType = uint32(val)
		}
	}
	ret, err := m.check(&_360MonitorReq{
		ServerId: fmt.Sprintf("S%d", data.SrvId),
		QId:      data.PlatformUniquePlayerId,
		Name:     data.ActorName,
		Type:     chatType,
		ToQid:    data.PlatformUniqueTargetPlayerId,
		ToName:   data.TargetActorName,
		RoleId:   fmt.Sprintf("%d", data.TargetActorId),
		Content:  data.Content,
		IP:       data.ActorIP,
		LoginKey: data.PlatformLoginKey,
	})
	return ret, err
}

func (m *_360Monitor) SetNameBusinessId(id string) {
}

func (m *_360Monitor) SetChatBusinessId(id string) {
}

func (m *_360Monitor) ClearCache() {
}
