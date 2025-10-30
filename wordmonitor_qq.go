/**
 * @Author: lzp
 * @Date: 2025/10/28
 * @Desc:
**/

package wordmonitor

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// 1:名称 2:消息 3:标题 4:评论 5:签名 6:搜索 7:其他
const (
	ContentTypeQQWanByName    = 1
	ContentTypeQQWanByMessage = 2
	ContentTypeQQWanByTitle   = 3
	ContentTypeQQWanByComment = 4
	ContentTypeQQWanBySign    = 5
	ContentTypeQQWanBySearch  = 6
	ContentTypeQQWanByOther   = 7
)

const _qqApiUrl = "https://openapi.minigame.qq.com/v3/user/uic_filter"

type _qqWanMonitor struct {
	AppId   string
	AppKey  string
	QZonePF string
	Format  string
}

type _qqWanMonitorReq struct {
	OpenId      string `json:"openid"`
	OpenKey     string `json:"openkey"`
	AppId       string `json:"appid"`
	Sig         string `json:"sig"`
	Content     string `json:"content"`
	ContentType int    `json:"contenttype"`
}

func genTencentSig(method, uriPath string, params map[string]string, appKey string) string {
	// 1. 按 key 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接参数字符串
	var paramStrs []string
	for _, k := range keys {
		paramStrs = append(paramStrs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramStr := strings.Join(paramStrs, "&")

	// 3. 构造 base_string
	baseString := fmt.Sprintf("%s&%s&%s",
		strings.ToUpper(method),
		url.QueryEscape(uriPath),
		url.QueryEscape(paramStr),
	)

	// 4. HMAC-SHA1 计算
	mac := hmac.New(sha1.New, []byte(appKey+"&"))
	mac.Write([]byte(baseString))
	sign := mac.Sum(nil)

	// 5. base64 编码
	return base64.StdEncoding.EncodeToString(sign)
}

func NewQQWanMonitor(appId, appKey, qZonePF, formatJson string) *_qqWanMonitor {
	return &_qqWanMonitor{
		AppId:   appId,
		AppKey:  appKey,
		QZonePF: qZonePF,
		Format:  formatJson,
	}
}
func (r *_qqWanMonitorReq) buildParams() map[string]string {
	return map[string]string{
		"openid":  r.OpenId,
		"openkey": r.OpenKey,
		"appid":   r.AppId,
		"sig":     r.Sig,
	}
}

func (r *_qqWanMonitorReq) buildBody() map[string]interface{} {
	return map[string]interface{}{
		"text_list_": []map[string]interface{}{
			{
				"text_":       r.Content,
				"text_scene_": r.ContentType,
			},
		},
	}
}

func (m *_qqWanMonitor) check(req *_qqWanMonitorReq) (result Ret, err error) {
	result = Failed
	params := req.buildParams()
	body := req.buildBody()

	response, err := resty.New().R().
		SetQueryParams(params).
		SetHeader("Content-Type", "application/json; charset=utf-8").
		SetBody(body).
		Post(_qqApiUrl)
	if err != nil {
		return
	}

	retJson, err := simplejson.NewJson(response.Body())
	if err != nil {
		return
	}

	ret := retJson.Get("ret").MustInt(-1)
	if ret == 0 {
		result = Success
	} else {
		err = fmt.Errorf("检测不通过")
	}
	return
}

func (m *_qqWanMonitor) CheckChat(data *CommonData) (Ret, error) {
	var platformUniquePlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	params := map[string]string{
		"openid":  fmt.Sprintf("%d", platformUniquePlayerId),
		"openkey": data.OpenKey,
		"userip":  data.ActorIP,
		"appid":   m.AppId,
		"pf":      m.QZonePF,
		"format":  m.Format,
	}

	sig := genTencentSig("GET", "/v3/user/get_info", params, m.AppKey)

	ret, err := m.check(&_qqWanMonitorReq{
		OpenId:      fmt.Sprintf("%d", data.ActorId),
		OpenKey:     data.OpenKey,
		AppId:       m.AppId,
		Sig:         sig,
		Content:     data.Content,
		ContentType: ContentTypeQQWanByMessage,
	})
	return ret, err
}

func (m *_qqWanMonitor) CheckName(data *CommonData) (Ret, error) {
	var platformUniquePlayerId int
	split := strings.Split(data.PlatformUniquePlayerId, "_")
	if len(split) > 0 {
		platformUniquePlayerId, _ = strconv.Atoi(split[len(split)-1])
	}

	params := map[string]string{
		"openid":  fmt.Sprintf("%d", platformUniquePlayerId),
		"openkey": data.OpenKey,
		"userip":  data.ActorIP,
		"appid":   m.AppId,
		"pf":      m.QZonePF,
		"format":  m.Format,
	}

	sig := genTencentSig("GET", "/v3/user/get_info", params, m.AppKey)

	ret, err := m.check(&_qqWanMonitorReq{
		OpenId:      fmt.Sprintf("%d", data.ActorId),
		OpenKey:     data.OpenKey,
		AppId:       m.AppId,
		Sig:         sig,
		Content:     data.Content,
		ContentType: ContentTypeQQWanByName,
	})
	return ret, err

}

func (m *_qqWanMonitor) SetNameBusinessId(id string) {
}

func (m *_qqWanMonitor) SetChatBusinessId(id string) {
}

func (m *_qqWanMonitor) ClearCache() {
}
