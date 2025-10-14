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
)

// doc:https://my.4399.com/game/autoJoin-doc-ifs-pbc
// 4399
const (
	_4399ApiUrl = "https://wo.webgame138.com/test/matchService.do"
)

type _4399Monitor struct {
	App    string
	Secret string
}

type _4399MonitorReq struct {
	ToCheck string `json:"toCheck"`
}

func New4399Monitor(app, secret string) *_4399Monitor {
	return &_4399Monitor{
		App:    app,
		Secret: secret,
	}
}

func (r *_4399MonitorReq) MakeSign(secret string) string {
	// 8810
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", secret, r.ToCheck))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *_4399Monitor) check(req *_4399MonitorReq) (result Ret, err error) {
	result = Failed
	response, err := resty.New().R().
		SetQueryParams(map[string]string{
			"toCheck":  req.ToCheck,
			"app":      m.App,
			"secret":   m.Secret,
			"byPinyin": fmt.Sprintf("%v", true),
			"sig":      req.MakeSign(m.Secret),
		}).
		Post(_4399ApiUrl)
	if err != nil {
		return
	}
	body := response.Body()
	if string(body) == "{}" {
		result = Success
	} else {
		err = fmt.Errorf("检测不通过 %s", string(body))
	}
	return
}

func (m *_4399Monitor) CheckName(data *CommonData) (Ret, error) {
	ret, err := m.check(&_4399MonitorReq{
		ToCheck: data.Content,
	})
	return ret, err
}

func (m *_4399Monitor) CheckChat(data *CommonData) (Ret, error) {
	ret, err := m.check(&_4399MonitorReq{
		ToCheck: data.Content,
	})
	return ret, err
}

func (m *_4399Monitor) SetNameBusinessId(id string) {
}

func (m *_4399Monitor) SetChatBusinessId(id string) {
}

func (m *_4399Monitor) ClearCache() {
}
