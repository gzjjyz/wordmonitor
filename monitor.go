package wordmonitor

type Monitor interface {
	CheckName(data *CommonData) (Ret, error)
	CheckChat(data *CommonData) (Ret, error)
	SetNameBusinessId(string)
	SetChatBusinessId(string)
	ClearCache()
}

type CommonData struct {
	ActorId                      uint64
	ActorName                    string
	ActorIP                      string
	PlatformUniquePlayerId       uint64 // 平台帐号
	TargetActorId                uint64
	TargetActorName              string
	Content                      string
	PlatformUniqueTargetPlayerId uint64 // 平台帐号

	SrvId       uint64 // 服务器 id
	ChatChannel uint32 // 聊天频道

	PlatformLoginKey string // 平台登陆 key
}
