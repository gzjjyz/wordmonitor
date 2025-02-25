package wordmonitor

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gzjjyz/random"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// doc:https://support.dun.163.com/documents/588434200783982592?docId=589310433773625344

const (
	apiUrl  = "http://as.dun.163.com/v5/text/check"
	version = "v5.2"
)

type YDunMonitor struct {
	secretId  string //产品密钥ID，产品标识
	secretKey string //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露

	nameBusinessId string
	chatBusinessId string

	cache *wordCache
}

func NewYDunMonitor(ak, sk string) *YDunMonitor {
	return &YDunMonitor{
		secretId:  ak,
		secretKey: sk,
		cache:     newWordCache(),
	}
}

func (m *YDunMonitor) signature(params url.Values) string {
	var paramStr string
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		paramStr += key + params[key][0]
	}
	paramStr += m.secretKey
	md5Reader := md5.New()
	md5Reader.Write([]byte(paramStr))
	return hex.EncodeToString(md5Reader.Sum(nil))
}

func (m *YDunMonitor) check(businessId string, content string, dataId string) (int, error) {
	exit := m.cache.Exit(content)
	if exit {
		return int(Failed), nil
	}
	params := url.Values{}
	params.Set("businessId", businessId)
	params.Set("secretId", m.secretId)
	params.Set("timestamp", fmt.Sprintf("%v", time.Now().Unix()))
	params.Set("nonce", fmt.Sprintf("%v", random.UintU(10000000)))
	params.Set("dataId", dataId)
	params.Set("version", version)
	params.Set("content", content)
	params.Set("signature", m.signature(params))

	resp, err := http.Post(apiUrl, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if nil != err {
		return 0, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if nil != err {
		return 0, err
	}

	retJson, err := simplejson.NewJson(body)
	if err != nil {
		return 0, err
	}

	code, _ := retJson.Get("code").Int()
	if code == 200 {
		result := retJson.Get("result")
		antispam := result.Get("antispam")
		suggestion, _ := antispam.Get("suggestion").Int() // 0：通过，1：嫌疑，2：不通过
		if suggestion != 0 {
			m.cache.Add(content)
		}
		return suggestion, nil
	} else {
		msg, err := retJson.Get("msg").String()
		if nil != err {
			return 0, err
		}
		return 0, errors.New(msg)
	}
}

func (m *YDunMonitor) CheckName(data *CommonData) (Ret, error) {
	ret, err := m.check(m.nameBusinessId, data.Content, random.GenerateKey(10))
	return Ret(ret), err
}

func (m *YDunMonitor) CheckChat(data *CommonData) (Ret, error) {
	ret, err := m.check(m.chatBusinessId, data.Content, random.GenerateKey(10))
	return Ret(ret), err
}

func (m *YDunMonitor) SetNameBusinessId(id string) {
	m.nameBusinessId = id
}

func (m *YDunMonitor) SetChatBusinessId(id string) {
	m.chatBusinessId = id
}

func (m *YDunMonitor) ClearCache() {
	m.cache.Set = make(map[string]struct{})
}
