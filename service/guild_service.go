package service

import (
	"errors"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// 公会职位定义
const (
	GuildPositionLeader = 1 // 会长
	GuildPositionVice   = 2 // 副会长
	GuildPositionElite  = 3 // 精英
	GuildPositionMember = 4 // 普通成员
)

// 公会权限定义
const (
	GuildPermissionKickMember      = 1 << 0 // 踢出成员
	GuildPermissionInviteMember    = 1 << 1 // 邀请成员
	GuildPermissionUpdateNotice    = 1 << 2 // 更新公告
	GuildPermissionManageApply     = 1 << 3 // 管理申请
	GuildPermissionSetPosition     = 1 << 4 // 设置职位
	GuildPermissionTransferLeader  = 1 << 5 // 转让会长
	GuildPermissionUpgradeGuild    = 1 << 6 // 升级公会
	GuildPermissionManageTerritory = 1 << 7 // 管理领地
	GuildPermissionManageWarehouse = 1 << 8 // 管理仓库
)

// 公会申请状态定义
const (
	GuildApplyStatusPending  = 1 // 待处理
	GuildApplyStatusAccepted = 2 // 已接受
	GuildApplyStatusRejected = 3 // 已拒绝
)

// GuildMember 公会成员
type GuildMember struct {
	playerId     int64
	name         string
	position     int
	contribution int64
	joinTime     int64
	online       bool
	lastOnline   int64
}

// GuildApply 公会申请
type GuildApply struct {
	applyId    int64
	playerId   int64
	playerName string
	guildId    int64
	applyTime  int64
	status     int
	remark     string
}

// Guild 公会结构
type Guild struct {
	guildId          int64
	name             string
	level            int
	exp              int64
	leaderId         int64
	memberCount      int
	maxMembers       int
	members          *zMap.Map // key: int64(playerId), value: *GuildMember
	applies          *zMap.Map // key: int64(applyId), value: *GuildApply
	createTime       int64
	notice           string
	warScore         int64         // 公会战积分
	territories      []int         // 占领的领地
	permissionConfig map[int]int64 // key: position, value: permission mask
}

// GuildService 公会服务
type GuildService struct {
	zObject.BaseObject
	logger       *zap.Logger
	guilds       *zMap.Map // key: int64(guildId), value: *Guild
	playerGuild  *zMap.Map // key: int64(playerId), value: int64(guildId)
	guildNameMap *zMap.Map // key: string(guildName), value: int64(guildId)
	maxGuilds    int
}

func NewGuildService() *GuildService {
	gs := &GuildService{
		logger:       zLog.GetLogger(),
		guilds:       zMap.NewMap(),
		playerGuild:  zMap.NewMap(),
		guildNameMap: zMap.NewMap(),
		maxGuilds:    1000,
	}
	gs.SetId(ServiceIdGuildService)
	return gs
}

func (gs *GuildService) Init() error {
	gs.logger.Info("Initializing guild service...")
	// 初始化公会服务相关资源
	return nil
}

func (gs *GuildService) Close() error {
	gs.logger.Info("Closing guild service...")
	// 清理公会服务相关资源
	gs.guilds.Clear()
	gs.playerGuild.Clear()
	gs.guildNameMap.Clear()
	return nil
}

func (gs *GuildService) Serve() {
	// 公会服务不需要持续运行的协程
}

// CreateGuild 创建公会
func (gs *GuildService) CreateGuild(leaderId int64, leaderName string, guildName string) (*Guild, error) {
	// 检查公会名称是否已存在
	if _, exists := gs.guildNameMap.Get(guildName); exists {
		return nil, errors.New("guild name already exists")
	}

	// 检查玩家是否已有公会
	if _, exists := gs.playerGuild.Get(leaderId); exists {
		return nil, errors.New("player already in a guild")
	}

	// 检查是否达到最大公会数量
	if gs.guilds.Len() >= int64(gs.maxGuilds) {
		return nil, errors.New("max guild count reached")
	}

	// 生成公会ID (这里简化处理，实际应该从数据库获取或使用更复杂的生成逻辑)
	guildId := gs.guilds.Len() + 1

	// 创建公会成员
	leader := &GuildMember{
		playerId:     leaderId,
		name:         leaderName,
		position:     GuildPositionLeader,
		contribution: 0,
		joinTime:     time.Now().Unix(),
		online:       true,
		lastOnline:   time.Now().Unix(),
	}

	// 创建公会
	guild := &Guild{
		guildId:     guildId,
		name:        guildName,
		level:       1,
		exp:         0,
		leaderId:    leaderId,
		memberCount: 1,
		maxMembers:  20,
		members:     zMap.NewMap(),
		applies:     zMap.NewMap(),
		createTime:  time.Now().Unix(),
		notice:      "欢迎加入公会！",
		warScore:    0,
		territories: make([]int, 0),
		permissionConfig: map[int]int64{
			GuildPositionLeader: GuildPermissionKickMember | GuildPermissionInviteMember |
				GuildPermissionUpdateNotice | GuildPermissionManageApply |
				GuildPermissionSetPosition | GuildPermissionTransferLeader |
				GuildPermissionUpgradeGuild | GuildPermissionManageTerritory |
				GuildPermissionManageWarehouse,
			GuildPositionVice: GuildPermissionKickMember | GuildPermissionInviteMember |
				GuildPermissionUpdateNotice | GuildPermissionManageApply |
				GuildPermissionSetPosition,
			GuildPositionElite:  GuildPermissionInviteMember,
			GuildPositionMember: 0,
		},
	}

	// 添加会长到公会成员列表
	guild.members.Store(leaderId, leader)

	// 存储公会信息
	gs.guilds.Store(guildId, guild)
	gs.playerGuild.Store(leaderId, guildId)
	gs.guildNameMap.Store(guildName, guildId)

	gs.logger.Info("Created new guild", zap.Int64("guildId", guildId), zap.String("guildName", guildName), zap.Int64("leaderId", leaderId))
	return guild, nil
}

// JoinGuild 加入公会
func (gs *GuildService) JoinGuild(guildId int64, playerId int64, playerName string) error {
	// 检查公会是否存在
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return errors.New("guild not found")
	}
	guild := guildInterface.(*Guild)

	// 检查玩家是否已有公会
	if _, exists := gs.playerGuild.Get(playerId); exists {
		return errors.New("player already in a guild")
	}

	// 检查公会是否已满
	if guild.memberCount >= guild.maxMembers {
		return errors.New("guild is full")
	}

	// 创建公会成员
	member := &GuildMember{
		playerId:     playerId,
		name:         playerName,
		position:     GuildPositionMember,
		contribution: 0,
		joinTime:     time.Now().Unix(),
		online:       true,
		lastOnline:   time.Now().Unix(),
	}

	// 添加成员到公会
	guild.members.Store(playerId, member)
	guild.memberCount++

	// 记录玩家与公会的映射关系
	gs.playerGuild.Store(playerId, guildId)

	gs.logger.Info("Player joined guild", zap.Int64("playerId", playerId), zap.String("playerName", playerName), zap.Int64("guildId", guildId), zap.String("guildName", guild.name))
	return nil
}

// LeaveGuild 离开公会
func (gs *GuildService) LeaveGuild(playerId int64) error {
	// 检查玩家是否在公会中
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return errors.New("player not in a guild")
	}
	guildId := guildIdInterface.(int64)

	// 检查公会是否存在
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return errors.New("guild not found")
	}
	guild := guildInterface.(*Guild)

	// 检查是否是会长
	if guild.leaderId == playerId {
		// 如果是会长，需要解散公会或转让会长
		// 这里简化处理，直接解散公会
		return gs.DisbandGuild(guildId)
	}

	// 从公会成员列表中移除玩家
	guild.members.Delete(playerId)
	guild.memberCount--

	// 移除玩家与公会的映射关系
	gs.playerGuild.Delete(playerId)

	gs.logger.Info("Player left guild", zap.Int64("playerId", playerId), zap.Int64("guildId", guildId), zap.String("guildName", guild.name))
	return nil
}

// DisbandGuild 解散公会
func (gs *GuildService) DisbandGuild(guildId int64) error {
	// 检查公会是否存在
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return errors.New("guild not found")
	}
	guild := guildInterface.(*Guild)

	// 移除所有玩家与公会的映射关系
	guild.members.Range(func(key, value interface{}) bool {
		playerId := key.(int64)
		gs.playerGuild.Delete(playerId)
		return true
	})

	// 移除公会信息
	gs.guilds.Delete(guildId)
	gs.guildNameMap.Delete(guild.name)

	gs.logger.Info("Guild disbanded", zap.Int64("guildId", guildId), zap.String("guildName", guild.name))
	return nil
}

// GetGuild 获取公会信息
func (gs *GuildService) GetGuild(guildId int64) (*Guild, bool) {
	guild, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil, false
	}
	return guild.(*Guild), true
}

// GetGuildByPlayer 获取玩家所在公会
func (gs *GuildService) GetGuildByPlayer(playerId int64) (*Guild, bool) {
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return nil, false
	}
	guildId := guildIdInterface.(int64)

	guild, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil, false
	}

	return guild.(*Guild), true
}

// GetPlayerGuildInfo 获取玩家公会信息
func (gs *GuildService) GetPlayerGuildInfo(playerId int64) (*Guild, *GuildMember, bool) {
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return nil, nil, false
	}
	guildId := guildIdInterface.(int64)

	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil, nil, false
	}
	guild := guildInterface.(*Guild)

	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return nil, nil, false
	}
	member := memberInterface.(*GuildMember)

	return guild, member, true
}
