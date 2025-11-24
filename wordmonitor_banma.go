/**
 * @Author:
 * @Date:
 * @Desc:
**/

package wordmonitor

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
)

// 斑马
const (
	_BanMaApiUrl = "https://h5.jzsyvip.cn/checkword/4399/2072"
)

type _BanMaMonitor struct {
	AppId string
}

type _BanMaMonitorReq struct {
	ToCheck string `json:"toCheck"`
}

type _BanMaMonitorResp struct {
	Code string                     `json:"code"`
	Msg  string                     `json:"msg"`
	Data map[string]*_BanMa4399Resp `json:"data"`
}

type _BanMaMonitorRespV2 struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data []*_BanMa4399Resp `json:"data"`
}

type _BanMa4399Resp struct {
	App      string `json:"app"`
	Level    uint32 `json:"level"`
	StartPos uint32 `json:"startPos"`
	EndPos   uint32 `json:"endPos"`
	MaskWork string `json:"mask_work"`
}

func NewBanMaMonitor(appId string) *_BanMaMonitor {
	return &_BanMaMonitor{
		AppId: appId,
	}
}

func (r *_BanMaMonitorReq) MakeSign(secret string) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", secret, r.ToCheck))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_BanMaMonitor) check(req *_BanMaMonitorReq) (result Ret, err error) {
	result = Failed
	response, err := resty.New().R().
		SetQueryParams(map[string]string{
			"toCheck":  req.ToCheck,
			"appid":    m.AppId,
			"byPinyin": fmt.Sprintf("%v", true),
		}).
		Post(_BanMaApiUrl)
	if err != nil {
		return
	}
	body := response.Body()
	var resp _BanMaMonitorResp
	err = json.Unmarshal(body, &resp)
	if err == nil {
		if len(resp.Data) == 0 {
			result = Success
		} else {
			err = fmt.Errorf("检测不通过 %s, err:%v", string(body), err)
		}
		return
	}
	var respV2 _BanMaMonitorRespV2
	err = json.Unmarshal(body, &respV2)
	if err == nil {
		if len(resp.Data) == 0 {
			result = Success
		} else {
			err = fmt.Errorf("检测不通过 %s, err:%v", string(body), err)
		}
		return
	}
	return
}

func (m *_BanMaMonitor) CheckName(data *CommonData) (Ret, error) {
	ret, err := m.check(&_BanMaMonitorReq{
		ToCheck: data.Content,
	})
	return ret, err
}

func (m *_BanMaMonitor) CheckChat(data *CommonData) (Ret, error) {
	ret, err := m.check(&_BanMaMonitorReq{
		ToCheck: data.Content,
	})
	return ret, err
}

func (m *_BanMaMonitor) SetNameBusinessId(id string) {
}

func (m *_BanMaMonitor) SetChatBusinessId(id string) {
}

func (m *_BanMaMonitor) ClearCache() {
}
