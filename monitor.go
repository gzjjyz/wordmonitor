package wordmonitor

type Monitor interface {
	CheckName(data *CommonData) (Ret, error)
	CheckChat(data *CommonData) (Ret, error)
	SetNameBusinessId(string)
	SetChatBusinessId(string)
	ClearCache()
}

// 名称来源，CheckTypeName 时用于区分是哪种取名场景
const (
	NameSourceRole        = 1 // 角色名
	NameSourceGuild       = 2 // 仙盟名
	NameSourceGuildNotice = 3 // 仙盟公告
)

type CommonData struct {
	NameSource uint32

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
	OpenKey     string // 平台登录 key
	Platform    string // 平台渠道标识
}
