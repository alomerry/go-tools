package service

import (
	"context"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/components/kook/service/internal"
)

type ChannelService interface {
	List(ctx context.Context, guildId string, page, pageSize int, channelType int) (*model.ChannelListResp, error)
	View(ctx context.Context, targetId string, needChildren bool) (*model.ViewChannelResp, error)
	Create(ctx context.Context, req model.CreateChannelReq) (*model.Channel, error)
	Update(ctx context.Context, req model.UpdateChannelReq) (*model.Channel, error)
	Delete(ctx context.Context, channelId string) error
	UserList(ctx context.Context, channelId string) (*model.ChannelUserListResp, error)
	MoveUser(ctx context.Context, channelId string, userIds ...string) error
	Kickout(ctx context.Context, channelId string, userIds ...string) error
	GetJoinedChannel(ctx context.Context, guildId, userId string, page, pageSize int) (*model.ChannelListResp, error)
}

type GuildService interface {
	List(ctx context.Context, page, pageSize int, sort string) (*model.GuildListResp, error)
	View(ctx context.Context, guildId string) (*model.Guild, error)
	UserList(ctx context.Context, guildId string, page, pageSize int) (*model.GuildUserListResp, error)
	Leave(ctx context.Context, guildId string) error
	Kickout(ctx context.Context, guildId, targetId string) error
}

type MessageService interface {
	List(ctx context.Context, targetId string) (*model.MessageListResp, error)
	View(ctx context.Context, msgId string) (*model.Message, error)
	Create(context.Context, model.CreateMessageRequest) (*model.CreateMessageResp, error)
	Update(ctx context.Context, req model.MessageUpdateReq) error
	Delete(ctx context.Context, msgId string) error
	AddReaction(ctx context.Context, msgId, emoji string) error
	DeleteReaction(ctx context.Context, msgId, emoji, userId string) error
}

type DirectMessageService interface {
	List(ctx context.Context, chatCode, targetId string) (*model.DirectMessageListResp, error)
	View(ctx context.Context, msgId, chatCode string) (*model.DirectMessage, error)
	Create(ctx context.Context, req model.CreateDirectMessageReq) (*model.CreateDirectMessageResp, error)
	Update(ctx context.Context, req model.UpdateDirectMessageReq) error
	Delete(ctx context.Context, msgId string) error
	AddReaction(ctx context.Context, msgId, emoji string) error
	DeleteReaction(ctx context.Context, msgId, emoji string) error
}

type UserService interface {
	Me(ctx context.Context) (*model.UserMeResp, error)
	View(ctx context.Context, userId string, guildId string) (*model.UserViewResp, error)
	Offline(ctx context.Context) error
}

type RoleService interface {
	List(ctx context.Context, guildId string) (*model.GuildRoleListResp, error)
	Create(ctx context.Context, req model.CreateGuildRoleReq) (*model.GuildRole, error)
	Update(ctx context.Context, req model.UpdateGuildRoleReq) (*model.GuildRole, error)
	Delete(ctx context.Context, guildId string, roleId int) error
	Grant(ctx context.Context, guildId, userId string, roleId int) error
	Revoke(ctx context.Context, guildId, userId string, roleId int) error
}

type AssetService interface {
	Create(ctx context.Context, file []byte) (*model.AssetCreateResp, error)
}

type VoiceService interface {
	Join(ctx context.Context, channelId, password string) (*model.JoinVoiceResp, error)
	Leave(ctx context.Context, channelId string) error
	List(ctx context.Context) (*model.VoiceListResp, error)
	KeepAlive(ctx context.Context, channelId string) error
}

type EmojiService interface {
	List(ctx context.Context, guildId string) (*model.GuildEmojiListResp, error)
}

type InviteService interface {
	List(ctx context.Context, guildId, channelId string) (*model.InviteListResp, error)
}

type BlacklistService interface {
	List(ctx context.Context, guildId string) (*model.BlacklistListResp, error)
}

type IntimacyService interface {
	Update(ctx context.Context, req model.IntimacyUpdateReq) error
}

type BadgeService interface {
	List(ctx context.Context, guildId string) (*model.BadgeListResp, error)
}

type GameService interface {
	UpdateActivity(ctx context.Context, req model.GameActivityReq) error
}

type ThreadService interface {
	Create(ctx context.Context, req internal.CreateThreadReq) (*internal.CreateThreadResp, error)
}

// 工厂函数

func NewChannelService(client http.Client) ChannelService {
	return &internal.ChannelService{BaseService: internal.NewBaseService(client)}
}

func NewGuildService(client http.Client) GuildService {
	return &internal.GuildService{BaseService: internal.NewBaseService(client)}
}

func NewMessageService(client http.Client) MessageService {
	return &internal.MessageService{BaseService: internal.NewBaseService(client)}
}

func NewDirectMessageService(client http.Client) DirectMessageService {
	return &internal.DirectMessageService{BaseService: internal.NewBaseService(client)}
}

func NewUserService(client http.Client) UserService {
	return &internal.UserService{BaseService: internal.NewBaseService(client)}
}

func NewRoleService(client http.Client) RoleService {
	return &internal.RoleService{BaseService: internal.NewBaseService(client)}
}

func NewAssetService(client http.Client) AssetService {
	return &internal.AssetService{BaseService: internal.NewBaseService(client)}
}

func NewVoiceService(client http.Client) VoiceService {
	return &internal.VoiceService{BaseService: internal.NewBaseService(client)}
}

func NewEmojiService(client http.Client) EmojiService {
	return &internal.EmojiService{BaseService: internal.NewBaseService(client)}
}

func NewInviteService(client http.Client) InviteService {
	return &internal.InviteService{BaseService: internal.NewBaseService(client)}
}

func NewBlacklistService(client http.Client) BlacklistService {
	return &internal.BlacklistService{BaseService: internal.NewBaseService(client)}
}

func NewIntimacyService(client http.Client) IntimacyService {
	return &internal.IntimacyService{BaseService: internal.NewBaseService(client)}
}

func NewBadgeService(client http.Client) BadgeService {
	return &internal.BadgeService{BaseService: internal.NewBaseService(client)}
}

func NewGameService(client http.Client) GameService {
	return &internal.GameService{BaseService: internal.NewBaseService(client)}
}

func NewThreadService(client http.Client) ThreadService {
	return &internal.ThreadService{BaseService: internal.NewBaseService(client)}
}
