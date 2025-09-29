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
	PlatformUniquePlayerId       string // 平台帐号
	TargetActorId                uint64
	TargetActorName              string
	Content                      string
	PlatformUniqueTargetPlayerId string // 平台帐号

	SrvId       uint32 // 服务器 id
	ChatChannel uint32 // 聊天频道
	GuildId     uint64 // 仙盟id
}
