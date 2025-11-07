package wordmonitor

import (
	"strings"
)

type Ret int8

const (
	Success Ret = 0 // 通过
	Suspect Ret = 1 // 嫌疑
	Failed  Ret = 2 // 不通过
)

func GetPlatformUid(account string) (uid string) {
	split := strings.Split(account, "_")
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	return
}
