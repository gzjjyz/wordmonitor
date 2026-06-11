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

func GetPlatformUid(platformUniquePlayerId string) string {
	split := strings.SplitN(platformUniquePlayerId, "_", 2)
	if len(split) > 1 {
		return split[1]
	}
	return platformUniquePlayerId
}
