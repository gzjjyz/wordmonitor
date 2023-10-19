package wordmonitor

type Ret int8

const (
	Success Ret = 0 // 通过
	Suspect Ret = 1 // 嫌疑
	Failed  Ret = 2 // 不通过
)
