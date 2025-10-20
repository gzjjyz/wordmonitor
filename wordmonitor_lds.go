/**
 * @Author:
 * @Date:
 * @Desc: 鲁大师
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

// doc:https://www.showdoc.com.cn/ludashi/10791204051230107
// lds
const (
	_ldsApiUrl = "https://zhushou-l.ludashi.com/wan/chat/ban/"

	ChatTypeLdsByPrivate   = 1 // 1 私聊；
	ChatTypeLdsByBroadcast = 2 // 2 喇叭；
	ChatTypeLdsByWorld     = 4 // 4 世界；
	ChatTypeLdsByGuild     = 6 // 6 工会/帮会；
	ChatTypeLdsByTeam      = 7 // 7 队伍；
	ChatTypeLdsByNear      = 8 // 8 交易；
	ChatTypeLdsByOther     = 9 // 9 区域;
)

type _ldsMonitor struct {
	Gid        string `json:"gid"`
	Key        string
	ChannelMap map[int]int
}

type _ldsMonitorReq struct {
	Gid     string `json:"gid"`
	Sid     string `json:"sid"`
	Uid     string `json:"uid"`
	URole   string `json:"urole"`
	URoleId string `json:"urole_id"`
	Channel string `json:"channel"`
	Tid     string `json:"tid"`
	TRole   string `json:"trole"`
	Body    string `json:"body"`
	Key     string `json:"-"`
}

func NewLdsMonitor(gid, key string, channelMap map[int]int) *_ldsMonitor {
	return &_ldsMonitor{
		Gid:        gid,
		Key:        key,
		ChannelMap: channelMap,
	}
}

func (r *_ldsMonitorReq) ToFormData(timeSec int64) map[string]string {
	return map[string]string{
		"gid":      r.Gid,
		"sid":      r.Sid,
		"uid":      r.Uid,
		"urole":    r.URole,
		"urole_id": r.URoleId,
		"channel":  r.Channel,
		"tid":      r.Tid,
		"trole":    r.TRole,
		"body":     r.Body,
		"time":     fmt.Sprintf("%d", timeSec),
		"sign":     r.MakeSign(timeSec),
	}
}

func (r *_ldsMonitorReq) MakeSign(timeSec int64) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s%s%d%s", r.Gid, r.Sid, r.Uid, timeSec, r.Key))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_ldsMonitor) check(req *_ldsMonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post(_ldsApiUrl)
	if err != nil {
		return
	}
	body := response.Body()
	retJson, err := simplejson.NewJson(body)
	if err != nil {
		return
	}
	errCode, _ := retJson.Get("errno").Int()
	if errCode != 0 {
		err = fmt.Errorf("检测不通过 %s", string(body))
	} else {
		result = Success
	}
	return
}

func (m *_ldsMonitor) CheckName(data *CommonData) (Ret, error) {
	// 不提供校验取名 默认不通过
	return Failed, nil
}

func (m *_ldsMonitor) CheckChat(data *CommonData) (Ret, error) {
	var platformUniquePlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	var chatType = uint32(ChatTypeLdsByWorld)
	if m.ChannelMap != nil {
		val, ok := m.ChannelMap[int(data.ChatChannel)]
		if ok {
			chatType = uint32(val)
		}
	}
	ret, err := m.check(&_ldsMonitorReq{
		Gid:     m.Gid,
		Sid:     fmt.Sprintf("%d", data.SrvId),
		Uid:     fmt.Sprintf("%d", platformUniquePlayerId),
		URole:   data.ActorName,
		URoleId: fmt.Sprintf("%d", data.ActorId),
		Channel: fmt.Sprintf("%d", chatType),
		Tid:     fmt.Sprintf("%d", data.TargetActorId),
		TRole:   data.TargetActorName,
		Body:    data.Content,
		Key:     m.Key,
	})
	return ret, err
}

func (m *_ldsMonitor) SetNameBusinessId(id string) {
}

func (m *_ldsMonitor) SetChatBusinessId(id string) {
}

func (m *_ldsMonitor) ClearCache() {
}
