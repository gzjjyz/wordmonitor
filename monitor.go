package wordmonitor

type Monitor interface {
	CheckName(content string) (Ret, error)
	CheckChat(content string) (Ret, error)
	SetNameBusinessId(string)
	SetChatBusinessId(string)
	ClearCache()
}
