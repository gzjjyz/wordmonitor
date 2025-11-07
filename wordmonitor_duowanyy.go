package wordmonitor

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"time"
)

const (
	_DuoWanYYApiUrl = "https://twapi-out.yy.com/txt/api"
)

type _DuoWanYYMonitor struct {
	appid    int
	secretId string
	srvId    uint32
	lastSec  int64
	sequence uint16
}

func NewDuoWanYYMonitor(appid int, secretId string, serverId uint32) *_DuoWanYYMonitor {
	return &_DuoWanYYMonitor{
		appid:    appid,
		secretId: secretId,
		srvId:    serverId,
		lastSec:  time.Now().Unix(),
		sequence: 0,
	}
}

func (r *_DuoWanYYMonitor) GenerateRandom() int32 {
	currentTime := time.Now().Unix()

	if currentTime == r.lastSec {
		r.sequence++
	} else {
		r.lastSec = currentTime
		r.sequence = 0
	}

	serverPart := uint32(r.srvId) & 0x7FFF
	timePart := uint32(currentTime) & 0xFF
	seqPart := uint32(r.sequence) & 0xFF

	randomValue := (serverPart << 16) | (timePart << 8) | seqPart
	return int32(randomValue)
}

func (r *_DuoWanYYMonitor) CheckName(data *CommonData) (Ret, error) {
	var (
		platformUniquePlayerId = GetPlatformUid(data.PlatformUniquePlayerId)
	)

	contentMap := map[string]string{
		"nick": data.Content,
	}

	jsonBytes, _ := json.Marshal(contentMap)

	random := r.GenerateRandom()
	ret, err := r.check(&_DuoWanYYMonitorReq{
		Sign:     "1",
		SecretId: r.secretId,
		Appid:    r.appid,
		Random:   random,
		Serial:   fmt.Sprintf("%d-%d-%d", r.srvId, r.lastSec, r.sequence),
		Content:  string(jsonBytes),
		DataType: "json",
		Account:  platformUniquePlayerId,
	})
	return ret, err
}

func (r *_DuoWanYYMonitor) check(req *_DuoWanYYMonitorReq) (result Ret, err error) {
	result = Failed
	unix := time.Now().Unix()
	formData := req.ToFormData(unix)
	response, err := resty.New().R().
		SetFormData(formData).
		Post(_DuoWanYYApiUrl)
	if err != nil {
		return
	}
	body := response.Body()

	retJson, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, _ := retJson.Get("code").Int()
	if code != 100 {
		err = fmt.Errorf("请求失败:%d", code)
		return
	}

	status, _ := retJson.Get("result").Get("status").Int()
	if status == 1 {
		result = Success
	} else {
		err = fmt.Errorf("检测不通过")
	}

	return
}

func (r *_DuoWanYYMonitor) CheckChat(data *CommonData) (Ret, error) {
	var (
		platformUniquePlayerId = GetPlatformUid(data.PlatformUniquePlayerId)
	)

	contentMap := map[string]string{
		"content": data.Content,
	}

	jsonBytes, _ := json.Marshal(contentMap)

	random := r.GenerateRandom()
	ret, err := r.check(&_DuoWanYYMonitorReq{
		Sign:     "1",
		SecretId: r.secretId,
		Appid:    r.appid,
		Random:   random,
		Serial:   fmt.Sprintf("%d-%d-%d", r.srvId, r.lastSec, r.sequence),
		Content:  string(jsonBytes),
		DataType: "json",
		Account:  platformUniquePlayerId,
	})
	return ret, err
}

func (r *_DuoWanYYMonitor) SetNameBusinessId(id string) {
}

func (r *_DuoWanYYMonitor) SetChatBusinessId(id string) {
}

func (r *_DuoWanYYMonitor) ClearCache() {
}

type _DuoWanYYMonitorReq struct {
	Sign     string `json:"sign"`
	SecretId string `json:"secretId"`
	Appid    int    `json:"appid"`
	Random   int32  `json:"random"`
	Serial   string `json:"serial"`
	Content  string `json:"content"`
	DataType string `json:"dataType"`
	Account  string `json:"account"`
}

func (r *_DuoWanYYMonitorReq) ToFormData(timeSec int64) map[string]string {
	return map[string]string{
		"sign":      r.Sign,
		"secretId":  r.SecretId,
		"appid":     fmt.Sprintf("%d", r.Appid),
		"timestamp": fmt.Sprintf("%d", timeSec),
		"random":    fmt.Sprintf("%d", r.Random),
		"serial":    r.Serial,
		"content":   r.Content,
		"dataType":  r.DataType,
		"account":   r.Account,
	}
}
