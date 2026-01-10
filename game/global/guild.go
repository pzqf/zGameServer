package global

import (
	"errors"
	"time"

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

func NewGuildService(logger *zap.Logger) *GuildService {
	gs := &GuildService{
		logger:       logger,
		guilds:       zMap.NewMap(),
		playerGuild:  zMap.NewMap(),
		guildNameMap: zMap.NewMap(),
		maxGuilds:    1000,
	}
	gs.BaseObject.Id = "GuildService"
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
func (gs *GuildService) CreateGuild(guildId int64, guildName string, leaderId int64, leaderName string) (*Guild, error) {
	// 检查公会名称是否已存在
	if _, exists := gs.guildNameMap.Get(guildName); exists {
		return nil, nil // 公会名称已存在
	}

	// 检查是否达到最大公会数量
	if gs.guilds.Len() >= int64(gs.maxGuilds) {
		return nil, nil // 已达到最大公会数量
	}

	// 检查玩家是否已加入其他公会
	if _, exists := gs.playerGuild.Get(leaderId); exists {
		return nil, nil // 玩家已加入其他公会
	}

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 初始化权限配置
	permissionConfig := make(map[int]int64)
	// 会长拥有所有权限
	permissionConfig[GuildPositionLeader] =
		GuildPermissionKickMember | GuildPermissionInviteMember |
			GuildPermissionUpdateNotice | GuildPermissionManageApply |
			GuildPermissionSetPosition | GuildPermissionTransferLeader |
			GuildPermissionUpgradeGuild | GuildPermissionManageTerritory |
			GuildPermissionManageWarehouse
	// 副会长拥有大部分权限
	permissionConfig[GuildPositionVice] =
		GuildPermissionKickMember | GuildPermissionInviteMember |
			GuildPermissionUpdateNotice | GuildPermissionManageApply |
			GuildPermissionSetPosition
	// 精英成员只有邀请权限
	permissionConfig[GuildPositionElite] =
		GuildPermissionInviteMember
	// 普通成员没有管理权限
	permissionConfig[GuildPositionMember] = 0

	// 创建新公会
	guild := &Guild{
		guildId:          guildId,
		name:             guildName,
		level:            1,
		exp:              0,
		leaderId:         leaderId,
		memberCount:      1,
		maxMembers:       20,
		members:          zMap.NewMap(),
		applies:          zMap.NewMap(),
		createTime:       currentTime,
		notice:           "欢迎加入公会！",
		warScore:         0,
		territories:      []int{},
		permissionConfig: permissionConfig,
	}

	// 添加会长到公会成员
	guildMember := &GuildMember{
		playerId:     leaderId,
		name:         leaderName,
		position:     GuildPositionLeader,
		contribution: 0,
		joinTime:     currentTime,
		online:       true,
		lastOnline:   currentTime,
	}
	guild.members.Store(leaderId, guildMember)

	// 存储公会信息
	gs.guilds.Store(guildId, guild)
	gs.playerGuild.Store(leaderId, guildId)
	gs.guildNameMap.Store(guildName, guildId)

	gs.logger.Info("Guild created", zap.Int64("guildId", guildId), zap.String("guildName", guildName), zap.Int64("leaderId", leaderId))
	return guild, nil
}

// JoinGuild 加入公会
func (gs *GuildService) JoinGuild(playerId int64, playerName string, guildId int64) error {
	// 检查公会是否存在
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil // 公会不存在
	}
	guild := guildInterface.(*Guild)

	// 检查玩家是否已加入其他公会
	if _, exists := gs.playerGuild.Get(playerId); exists {
		return nil // 玩家已加入其他公会
	}

	// 检查公会是否已满
	if guild.memberCount >= guild.maxMembers {
		return nil // 公会已满
	}

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 添加玩家到公会成员
	guildMember := &GuildMember{
		playerId:     playerId,
		name:         playerName,
		position:     GuildPositionMember,
		contribution: 0,
		joinTime:     currentTime,
		online:       true,
		lastOnline:   currentTime,
	}
	guild.members.Store(playerId, guildMember)
	guild.memberCount++

	// 存储玩家公会关系
	gs.playerGuild.Store(playerId, guildId)

	gs.logger.Info("Player joined guild", zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
	return nil
}

// LeaveGuild 离开公会
func (gs *GuildService) LeaveGuild(playerId int64) error {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return nil // 玩家未加入公会
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil // 公会不存在
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return nil // 玩家不在公会中
	}
	member := memberInterface.(*GuildMember)

	// 如果是会长，需要处理会长转让
	if member.position == GuildPositionLeader {
		// 转让会长职位
		gs.transferGuildLeader(guild, playerId)
	}

	// 从公会成员中移除
	guild.members.Delete(playerId)
	guild.memberCount--

	// 如果公会成员为0，解散公会
	if guild.memberCount <= 0 {
		gs.DisbandGuild(guildId)
		return nil
	}

	// 移除玩家公会关系
	gs.playerGuild.Delete(playerId)

	gs.logger.Info("Player left guild", zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
	return nil
}

// DisbandGuild 解散公会
func (gs *GuildService) DisbandGuild(guildId int64) error {
	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil // 公会不存在
	}
	guild := guildInterface.(*Guild)

	// 移除所有成员的公会关系
	guild.members.Range(func(key, value interface{}) bool {
		playerId := key.(int64)
		gs.playerGuild.Delete(playerId)
		return true
	})

	// 移除公会名称映射
	gs.guildNameMap.Delete(guild.name)

	// 从公会列表中移除
	gs.guilds.Delete(guildId)

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

// GetGuildByPlayer 获取玩家所在的公会
func (gs *GuildService) GetGuildByPlayer(playerId int64) (*Guild, bool) {
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return nil, false
	}
	guildId := guildIdInterface.(int64)
	return gs.GetGuild(guildId)
}

// ApplyGuild 申请加入公会
func (gs *GuildService) ApplyGuild(applyId int64, playerId int64, playerName string, guildId int64, remark string) error {
	// 检查玩家是否已加入其他公会
	if _, exists := gs.playerGuild.Get(playerId); exists {
		return nil // 玩家已加入其他公会
	}

	// 检查公会是否存在
	if _, exists := gs.guilds.Get(guildId); !exists {
		return nil // 公会不存在
	}

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 创建公会申请
	apply := &GuildApply{
		applyId:    applyId,
		playerId:   playerId,
		playerName: playerName,
		guildId:    guildId,
		applyTime:  currentTime,
		status:     GuildApplyStatusPending,
		remark:     remark,
	}

	// 获取公会
	guildInterface, _ := gs.guilds.Get(guildId)
	guild := guildInterface.(*Guild)

	// 存储申请
	guild.applies.Store(applyId, apply)

	gs.logger.Info("Guild application submitted", zap.Int64("applyId", applyId), zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
	return nil
}

// transferGuildLeader 转让会长职位
func (gs *GuildService) transferGuildLeader(guild *Guild, oldLeaderId int64) {
	// 寻找新的会长（副会长 -> 精英 -> 普通成员）
	var newLeader *GuildMember
	var newLeaderId int64

	// 先找副会长
	guild.members.Range(func(key, value interface{}) bool {
		member := value.(*GuildMember)
		if member.position == GuildPositionVice {
			newLeader = member
			newLeaderId = member.playerId
			return false
		}
		return true
	})

	// 如果没有副会长，找精英成员
	if newLeader == nil {
		guild.members.Range(func(key, value interface{}) bool {
			member := value.(*GuildMember)
			if member.position == GuildPositionElite {
				newLeader = member
				newLeaderId = member.playerId
				return false
			}
			return true
		})
	}

	// 如果没有精英成员，找普通成员
	if newLeader == nil {
		guild.members.Range(func(key, value interface{}) bool {
			member := value.(*GuildMember)
			if member.playerId != oldLeaderId {
				newLeader = member
				newLeaderId = member.playerId
				return false
			}
			return true
		})
	}

	// 如果找到新会长，更新职位
	if newLeader != nil {
		// 更新新会长的职位
		newLeader.position = GuildPositionLeader
		guild.members.Store(newLeaderId, newLeader)
		guild.leaderId = newLeaderId

		// 更新旧会长的职位为普通成员
		if memberInterface, exists := guild.members.Get(oldLeaderId); exists {
			oldLeader := memberInterface.(*GuildMember)
			oldLeader.position = GuildPositionMember
			guild.members.Store(oldLeaderId, oldLeader)
		}

		gs.logger.Info("Guild leader transferred", zap.Int64("guildId", guild.guildId), zap.Int64("oldLeaderId", oldLeaderId), zap.Int64("newLeaderId", newLeaderId))
	}
}

// SetGuildMemberPosition 设置公会成员职位
func (gs *GuildService) SetGuildMemberPosition(operatorId int64, targetPlayerId int64, newPosition int) error {
	// 检查操作者是否有权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionSetPosition); err != nil {
		return err
	}

	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(targetPlayerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取目标玩家在公会中的信息
	memberInterface, exists := guild.members.Get(targetPlayerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 如果设置为会长，需要处理旧会长
	if newPosition == GuildPositionLeader {
		// 如果目标玩家已经是会长，不需要处理
		if member.position == GuildPositionLeader {
			return nil
		}

		// 获取旧会长
		oldLeaderInterface, exists := guild.members.Get(guild.leaderId)
		if exists {
			oldLeader := oldLeaderInterface.(*GuildMember)
			// 将旧会长职位改为普通成员
			oldLeader.position = GuildPositionMember
			guild.members.Store(guild.leaderId, oldLeader)
		}

		// 更新公会的leaderId
		guild.leaderId = targetPlayerId
	}

	// 更新目标玩家职位
	member.position = newPosition
	guild.members.Store(targetPlayerId, member)

	gs.logger.Info("Guild member position updated", zap.Int64("guildId", guildId), zap.Int64("playerId", targetPlayerId), zap.Int("newPosition", newPosition))
	return nil
}

// KickGuildMember 踢出公会成员
func (gs *GuildService) KickGuildMember(operatorId int64, targetPlayerId int64, reason string) error {
	// 检查操作者是否有权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionKickMember); err != nil {
		return err
	}

	// 调用LeaveGuild实现踢出功能
	return gs.LeaveGuild(targetPlayerId)
}

// UpdateGuildNotice 更新公会公告
func (gs *GuildService) UpdateGuildNotice(operatorId int64, notice string) error {
	// 检查操作者是否有权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionUpdateNotice); err != nil {
		return err
	}

	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(operatorId)
	if !exists {
		return nil // 玩家未加入公会
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil // 公会不存在
	}
	guild := guildInterface.(*Guild)

	// 更新公告
	guild.notice = notice
	gs.guilds.Store(guildId, guild)

	gs.logger.Info("Guild notice updated", zap.Int64("guildId", guildId), zap.String("notice", notice))
	return nil
}

// 自定义错误类型
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrPlayerNotInGuild = errors.New("player not in guild")
	ErrGuildNotExists   = errors.New("guild not exists")
)

// checkGuildPermission 检查公会权限（基于职位的权限检查）
func (gs *GuildService) checkGuildPermission(playerId int64, requiredPosition int) error {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 检查权限（职位数字越小，权限越高）
	if member.position > requiredPosition {
		return ErrPermissionDenied
	}

	return nil
}

// CheckGuildPermission 检查公会权限（基于权限掩码的细粒度检查）
func (gs *GuildService) CheckGuildPermission(playerId int64, permission int64) error {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 检查权限（基于权限掩码）
	if guild.permissionConfig[member.position]&permission == 0 {
		return ErrPermissionDenied
	}

	return nil
}

// ProcessGuildApply 处理公会申请
func (gs *GuildService) ProcessGuildApply(operatorId int64, applyId int64, accept bool, remark string) error {
	// 检查操作者是否有权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionManageApply); err != nil {
		return err
	}

	// 遍历所有公会查找申请
	var apply *GuildApply
	var guild *Guild
	var found bool

	gs.guilds.Range(func(key, value interface{}) bool {
		g := value.(*Guild)
		if a, exists := g.applies.Get(applyId); exists {
			apply = a.(*GuildApply)
			guild = g
			found = true
			return false
		}
		return true
	})

	if !found {
		return errors.New("application not exists")
	}

	// 更新申请状态
	apply.status = GuildApplyStatusRejected
	if accept {
		// 检查公会是否已满
		if guild.memberCount >= guild.maxMembers {
			return errors.New("guild is full")
		}

		// 接受申请，加入公会
		gs.JoinGuild(apply.playerId, apply.playerName, apply.guildId)
		apply.status = GuildApplyStatusAccepted
	}

	// 保存更新后的申请状态
	guild.applies.Store(applyId, apply)

	// 更新申请
	guild.applies.Store(applyId, apply)

	gs.logger.Info("Guild application processed", zap.Int64("applyId", applyId), zap.Bool("accepted", accept), zap.Int64("playerId", apply.playerId), zap.Int64("guildId", apply.guildId))
	return nil
}

// UpdateGuildMemberOnlineStatus 更新公会成员在线状态
func (gs *GuildService) UpdateGuildMemberOnlineStatus(playerId int64, online bool) error {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return nil // 玩家未加入公会
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil // 公会不存在
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return nil // 玩家不在公会中
	}
	member := memberInterface.(*GuildMember)

	// 更新在线状态
	member.online = online
	if !online {
		member.lastOnline = time.Now().UnixMilli()
	}

	// 存储更新后的成员信息
	guild.members.Store(playerId, member)

	gs.logger.Debug("Guild member online status updated", zap.Int64("playerId", playerId), zap.Bool("online", online))
	return nil
}

// GetGuildMembers 获取公会成员列表
func (gs *GuildService) GetGuildMembers(guildId int64) ([]*GuildMember, error) {
	// 检查公会是否存在
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return nil, ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取所有成员
	members := make([]*GuildMember, 0, guild.memberCount)
	guild.members.Range(func(key, value interface{}) bool {
		member := value.(*GuildMember)
		members = append(members, member)
		return true
	})

	return members, nil
}

// UpdateGuildMemberContribution 更新公会成员贡献
func (gs *GuildService) UpdateGuildMemberContribution(playerId int64, amount int64) error {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 更新贡献
	member.contribution += amount
	guild.exp += amount // 公会经验也增加相应的贡献值

	// 存储更新后的成员信息
	guild.members.Store(playerId, member)
	gs.guilds.Store(guildId, guild)

	gs.logger.Info("Guild member contribution updated", zap.Int64("playerId", playerId), zap.Int64("amount", amount), zap.Int64("total", member.contribution))
	return nil
}

// GetGuildMemberContribution 获取公会成员贡献
func (gs *GuildService) GetGuildMemberContribution(playerId int64) (int64, error) {
	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return 0, ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 获取玩家在公会中的信息
	memberInterface, exists := guild.members.Get(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	return member.contribution, nil
}

// UpgradeGuild 升级公会
func (gs *GuildService) UpgradeGuild(operatorId int64) error {
	// 检查操作者是否有权限（只有会长可以升级公会）
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionUpgradeGuild); err != nil {
		return err
	}

	// 检查玩家是否已加入公会
	guildIdInterface, exists := gs.playerGuild.Get(operatorId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	guildId := guildIdInterface.(int64)

	// 获取公会
	guildInterface, exists := gs.guilds.Get(guildId)
	if !exists {
		return ErrGuildNotExists
	}
	guild := guildInterface.(*Guild)

	// 计算升级所需经验
	requiredExp := int64(guild.level * 10000)
	if guild.exp < requiredExp {
		return errors.New("guild exp not enough")
	}

	// 扣除经验
	guild.exp -= requiredExp
	// 升级
	guild.level++
	// 增加最大成员数
	guild.maxMembers += 10

	// 存储更新后的公会信息
	gs.guilds.Store(guildId, guild)

	gs.logger.Info("Guild upgraded", zap.Int64("guildId", guildId), zap.Int("oldLevel", guild.level-1), zap.Int("newLevel", guild.level))
	return nil
}
