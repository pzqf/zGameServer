package guild

import (
	"errors"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zService"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// GuildService 公会服务
// 管理所有公会的创建、成员管理、权限控制等功能
type GuildService struct {
	zService.BaseService
	guilds       *zMap.TypedMap[common.GuildIdType, *Guild]                     // 公会映射表（GuildId -> Guild）
	playerGuild  *zMap.TypedShardedMap[common.PlayerIdType, common.GuildIdType] // 玩家公会映射表（PlayerId -> GuildId）
	guildNameMap *zMap.TypedMap[string, common.GuildIdType]                     // 公会名称映射表（Name -> GuildId）
	maxGuilds    int                                                            // 最大公会数量限制
}

// NewGuildService 创建公会服务
// 返回: 新创建的GuildService实例
func NewGuildService() *GuildService {
	gs := &GuildService{
		BaseService:  *zService.NewBaseService(common.ServiceIdGuild),
		guilds:       zMap.NewTypedMap[common.GuildIdType, *Guild](),
		playerGuild:  zMap.NewTypedShardedMap32[common.PlayerIdType, common.GuildIdType](),
		guildNameMap: zMap.NewTypedMap[string, common.GuildIdType](),
		maxGuilds:    1000,
	}
	return gs
}

// Init 初始化公会服务
// 返回: 初始化错误（如果有）
func (gs *GuildService) Init() error {
	gs.SetState(zService.ServiceStateInit)
	zLog.Info("Initializing guild service...", zap.String("serviceId", gs.ServiceId()))
	return nil
}

// Close 关闭公会服务
// 清理所有公会数据
// 返回: 关闭错误（如果有）
func (gs *GuildService) Close() error {
	gs.SetState(zService.ServiceStateStopping)
	zLog.Info("Closing guild service...", zap.String("serviceId", gs.ServiceId()))
	gs.guilds.Clear()
	gs.playerGuild.Clear()
	gs.guildNameMap.Clear()
	gs.SetState(zService.ServiceStateStopped)
	return nil
}

// Serve 启动服务
// 将服务状态设置为Running
func (gs *GuildService) Serve() {
	gs.SetState(zService.ServiceStateRunning)
}

// CreateGuild 创建公会
// 参数:
//   - guildId: 公会ID
//   - guildName: 公会名称
//   - leaderId: 会长玩家ID
//   - leaderName: 会长名称
//
// 返回:
//   - *Guild: 新创建的公会对象
//   - error: 创建错误（名称重复、数量超限、玩家已有公会等）
func (gs *GuildService) CreateGuild(guildId common.GuildIdType, guildName string, leaderId common.PlayerIdType, leaderName string) (*Guild, error) {
	// 检查公会名称是否已存在
	if _, exists := gs.guildNameMap.Load(guildName); exists {
		return nil, nil
	}

	// 检查公会数量上限
	if gs.guilds.Len() >= int64(gs.maxGuilds) {
		return nil, nil
	}

	// 检查创建者是否已有公会
	if _, exists := gs.playerGuild.Load(leaderId); exists {
		return nil, nil
	}

	currentTime := time.Now().UnixMilli()

	// 初始化权限配置
	permissionConfig := make(map[int]int64)
	// 会长拥有所有权限
	permissionConfig[GuildPositionLeader] =
		GuildPermissionKickMember |
			GuildPermissionInviteMember |
			GuildPermissionUpdateNotice |
			GuildPermissionManageApply |
			GuildPermissionSetPosition |
			GuildPermissionTransferLeader |
			GuildPermissionUpgradeGuild |
			GuildPermissionManageTerritory |
			GuildPermissionManageWarehouse
	// 副会长拥有大部分权限
	permissionConfig[GuildPositionVice] =
		GuildPermissionKickMember |
			GuildPermissionInviteMember |
			GuildPermissionUpdateNotice |
			GuildPermissionManageApply |
			GuildPermissionSetPosition
	// 精英成员只有邀请权限
	permissionConfig[GuildPositionElite] =
		GuildPermissionInviteMember
	// 普通成员没有管理权限
	permissionConfig[GuildPositionMember] = 0

	// 创建公会对象
	guild := &Guild{
		GuildId:          guildId,
		Name:             guildName,
		Level:            1,
		Exp:              0,
		LeaderId:         leaderId,
		MemberCount:      1,
		MaxMembers:       20,
		Members:          zMap.NewTypedShardedMap32[common.PlayerIdType, *GuildMember](),
		Applies:          zMap.NewTypedShardedMap32[int64, *GuildApply](),
		CreateTime:       currentTime,
		Notice:           "欢迎加入公会！",
		WarScore:         0,
		Territories:      []int{},
		PermissionConfig: permissionConfig,
	}

	// 创建会长成员
	guildMember := &GuildMember{
		PlayerId:     leaderId,
		Name:         leaderName,
		Position:     GuildPositionLeader,
		Contribution: 0,
		JoinTime:     currentTime,
		Online:       true,
		LastOnline:   currentTime,
	}
	guild.Members.Store(leaderId, guildMember)

	// 注册到映射表
	gs.guilds.Store(guildId, guild)
	gs.playerGuild.Store(leaderId, guildId)
	gs.guildNameMap.Store(guildName, guildId)

	zLog.Info("Guild created", zap.Int64("guildId", int64(guildId)), zap.String("guildName", guildName), zap.Int64("leaderId", int64(leaderId)))
	return guild, nil
}

// JoinGuild 加入公会
// 参数:
//   - playerId: 玩家ID
//   - playerName: 玩家名称
//   - guildId: 公会ID
//
// 返回:
//   - error: 加入错误
func (gs *GuildService) JoinGuild(playerId common.PlayerIdType, playerName string, guildId common.GuildIdType) error {
	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil
	}

	// 检查玩家是否已有公会
	if _, exists := gs.playerGuild.Load(playerId); exists {
		return nil
	}

	// 添加成员
	if err := guild.AddMember(playerId, playerName, GuildPositionMember); err != nil {
		return err
	}

	gs.playerGuild.Store(playerId, guildId)

	zLog.Info("Player joined guild", zap.Int64("playerId", int64(playerId)), zap.Int64("guildId", int64(guildId)))
	return nil
}

// LeaveGuild 离开公会
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - error: 离开错误
func (gs *GuildService) LeaveGuild(playerId common.PlayerIdType) error {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return nil
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil
	}

	// 移除成员
	if err := guild.RemoveMember(playerId); err != nil {
		return err
	}

	// 如果公会没有成员，解散公会
	if guild.MemberCount <= 0 {
		gs.DisbandGuild(guildId)
		return nil
	}

	gs.playerGuild.Delete(playerId)

	zLog.Info("Player left guild", zap.Int64("playerId", int64(playerId)), zap.Int64("guildId", int64(guildId)))
	return nil
}

// DisbandGuild 解散公会
// 参数:
//   - guildId: 公会ID
//
// 返回:
//   - error: 解散错误
func (gs *GuildService) DisbandGuild(guildId common.GuildIdType) error {
	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil
	}

	// 清理所有成员的公会映射
	guild.Members.Range(func(key common.PlayerIdType, value *GuildMember) bool {
		gs.playerGuild.Delete(key)
		return true
	})

	// 清理名称映射
	gs.guildNameMap.Delete(guild.Name)

	// 删除公会
	gs.guilds.Delete(guildId)

	zLog.Info("Guild disbanded", zap.Int64("guildId", int64(guildId)), zap.String("guildName", guild.Name))
	return nil
}

// GetGuild 获取公会信息
// 参数:
//   - guildId: 公会ID
//
// 返回:
//   - *Guild: 公会对象
//   - bool: 是否存在
func (gs *GuildService) GetGuild(guildId common.GuildIdType) (*Guild, bool) {
	return gs.guilds.Load(guildId)
}

// GetGuildByPlayer 获取玩家所在的公会
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *Guild: 公会对象
//   - bool: 是否存在
func (gs *GuildService) GetGuildByPlayer(playerId common.PlayerIdType) (*Guild, bool) {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return nil, false
	}
	return gs.GetGuild(guildId)
}

// ApplyGuild 申请加入公会
// 参数:
//   - applyId: 申请ID
//   - playerId: 申请人ID
//   - playerName: 申请人名称
//   - guildId: 目标公会ID
//   - remark: 申请备注
//
// 返回:
//   - error: 申请错误
func (gs *GuildService) ApplyGuild(applyId int64, playerId common.PlayerIdType, playerName string, guildId common.GuildIdType, remark string) error {
	// 检查玩家是否已有公会
	if _, exists := gs.playerGuild.Load(playerId); exists {
		return nil
	}

	// 检查公会是否存在
	if _, exists := gs.guilds.Load(guildId); !exists {
		return nil
	}

	currentTime := time.Now().UnixMilli()

	// 创建申请记录
	apply := &GuildApply{
		ApplyId:    applyId,
		PlayerId:   playerId,
		PlayerName: playerName,
		GuildId:    guildId,
		ApplyTime:  currentTime,
		Status:     GuildApplyStatusPending,
		Remark:     remark,
	}

	guild, _ := gs.guilds.Load(guildId)
	guild.Applies.Store(applyId, apply)

	zLog.Info("Guild application submitted", zap.Int64("applyId", applyId), zap.Int64("playerId", int64(playerId)), zap.Int64("guildId", int64(guildId)))
	return nil
}

// transferGuildLeader 转让会长职位
// 当会长离开公会时自动调用，按优先级选择新会长：副会长 -> 精英 -> 普通成员
// 参数:
//   - guild: 公会对象
//   - oldLeaderId: 旧会长ID
func (gs *GuildService) transferGuildLeader(guild *Guild, oldLeaderId common.PlayerIdType) {
	var newLeader *GuildMember
	var newLeaderId common.PlayerIdType

	// 先找副会长
	guild.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
		if member.Position == GuildPositionVice {
			newLeader = member
			newLeaderId = common.PlayerIdType(member.PlayerId)
			return false
		}
		return true
	})

	// 如果没有副会长，找精英成员
	if newLeader == nil {
		guild.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
			if member.Position == GuildPositionElite {
				newLeader = member
				newLeaderId = common.PlayerIdType(member.PlayerId)
				return false
			}
			return true
		})
	}

	// 如果没有精英成员，找普通成员
	if newLeader == nil {
		guild.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
			if common.PlayerIdType(member.PlayerId) != oldLeaderId {
				newLeader = member
				newLeaderId = common.PlayerIdType(member.PlayerId)
				return false
			}
			return true
		})
	}

	// 如果找到新会长，更新职位
	if newLeader != nil {
		newLeader.Position = GuildPositionLeader
		guild.Members.Store(newLeaderId, newLeader)
		guild.LeaderId = newLeaderId

		// 更新旧会长的职位为普通成员
		if oldLeader, exists := guild.Members.Load(oldLeaderId); exists {
			oldLeader.Position = GuildPositionMember
			guild.Members.Store(oldLeaderId, oldLeader)
		}

		zLog.Info("Guild leader transferred", zap.Int64("guildId", int64(guild.GuildId)), zap.Int64("oldLeaderId", int64(oldLeaderId)), zap.Int64("newLeaderId", int64(newLeaderId)))
	}
}

// SetGuildMemberPosition 设置公会成员职位
// 参数:
//   - operatorId: 操作者ID
//   - targetPlayerId: 目标玩家ID
//   - newPosition: 新职位
//
// 返回:
//   - error: 设置错误
func (gs *GuildService) SetGuildMemberPosition(operatorId common.PlayerIdType, targetPlayerId common.PlayerIdType, newPosition int) error {
	// 检查权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionSetPosition); err != nil {
		return err
	}

	guildId, exists := gs.playerGuild.Load(targetPlayerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return ErrGuildNotExists
	}

	if err := guild.SetMemberPosition(int64(targetPlayerId), newPosition); err != nil {
		return err
	}

	zLog.Info("Guild member position updated", zap.Int64("guildId", int64(guildId)), zap.Int64("playerId", int64(targetPlayerId)), zap.Int("newPosition", newPosition))
	return nil
}

// KickGuildMember 踢出公会成员
// 参数:
//   - operatorId: 操作者ID
//   - targetPlayerId: 目标玩家ID
//   - reason: 踢出原因
//
// 返回:
//   - error: 踢出错误
func (gs *GuildService) KickGuildMember(operatorId common.PlayerIdType, targetPlayerId common.PlayerIdType, reason string) error {
	// 检查权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionKickMember); err != nil {
		return err
	}

	return gs.LeaveGuild(targetPlayerId)
}

// UpdateGuildNotice 更新公会公告
// 参数:
//   - operatorId: 操作者ID
//   - notice: 新公告内容
//
// 返回:
//   - error: 更新错误
func (gs *GuildService) UpdateGuildNotice(operatorId common.PlayerIdType, notice string) error {
	// 检查权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionUpdateNotice); err != nil {
		return err
	}

	guildId, exists := gs.playerGuild.Load(operatorId)
	if !exists {
		return nil
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil
	}

	guild.UpdateNotice(notice)

	zLog.Info("Guild notice updated", zap.Int64("guildId", int64(guildId)), zap.String("notice", notice))
	return nil
}

// checkGuildPermission 检查公会权限（基于职位的权限检查）
// 这是旧版检查方式，仅比较职位等级
// 参数:
//   - playerId: 玩家ID
//   - requiredPosition: 要求的最低职位
//
// 返回:
//   - error: 权限不足错误
func (gs *GuildService) checkGuildPermission(playerId common.PlayerIdType, requiredPosition int) error {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return ErrGuildNotExists
	}

	member, exists := guild.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	// 职位数值越小，权限越高
	if member.Position > requiredPosition {
		return ErrPermissionDenied
	}

	return nil
}

// CheckGuildPermission 检查公会权限（基于权限掩码的细粒度检查）
// 使用位运算检查玩家是否拥有特定权限
// 参数:
//   - playerId: 玩家ID
//   - permission: 要求的权限位
//
// 返回:
//   - error: 权限不足错误
func (gs *GuildService) CheckGuildPermission(playerId common.PlayerIdType, permission int64) error {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return ErrGuildNotExists
	}

	return guild.CheckPermission(playerId, permission)
}

// ProcessGuildApply 处理公会申请
// 参数:
//   - operatorId: 操作者ID
//   - applyId: 申请ID
//   - accept: 是否接受
//   - remark: 处理备注
//
// 返回:
//   - error: 处理错误
func (gs *GuildService) ProcessGuildApply(operatorId common.PlayerIdType, applyId int64, accept bool, remark string) error {
	// 检查权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionManageApply); err != nil {
		return err
	}

	var apply *GuildApply
	var guild *Guild
	var found bool

	// 遍历所有公会查找申请
	gs.guilds.Range(func(key common.GuildIdType, g *Guild) bool {
		if a, exists := g.Applies.Load(applyId); exists {
			apply = a
			guild = g
			found = true
			return false
		}
		return true
	})

	if !found {
		return errors.New("application not exists")
	}

	// 处理申请
	apply.Status = GuildApplyStatusRejected
	if accept {
		// 检查公会是否已满
		if guild.MemberCount >= guild.MaxMembers {
			return errors.New("guild is full")
		}

		gs.JoinGuild(common.PlayerIdType(apply.PlayerId), apply.PlayerName, common.GuildIdType(apply.GuildId))
		apply.Status = GuildApplyStatusAccepted
	}

	guild.Applies.Store(applyId, apply)

	zLog.Info("Guild application processed", zap.Int64("applyId", applyId), zap.Bool("accepted", accept), zap.Int64("playerId", int64(apply.PlayerId)), zap.Int64("guildId", int64(apply.GuildId)))
	return nil
}

// UpdateGuildMemberOnlineStatus 更新公会成员在线状态
// 参数:
//   - playerId: 玩家ID
//   - online: 是否在线
//
// 返回:
//   - error: 更新错误
func (gs *GuildService) UpdateGuildMemberOnlineStatus(playerId common.PlayerIdType, online bool) error {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return nil
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil
	}

	member, exists := guild.Members.Load(playerId)
	if !exists {
		return nil
	}

	member.Online = online
	if !online {
		member.LastOnline = time.Now().UnixMilli()
	}

	guild.Members.Store(playerId, member)

	zLog.Debug("Guild member online status updated", zap.Int64("playerId", int64(playerId)), zap.Bool("online", online))
	return nil
}

// GetGuildMembers 获取公会成员列表
// 参数:
//   - guildId: 公会ID
//
// 返回:
//   - []*GuildMember: 成员列表
//   - error: 获取错误
func (gs *GuildService) GetGuildMembers(guildId common.GuildIdType) ([]*GuildMember, error) {
	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return nil, ErrGuildNotExists
	}

	members := make([]*GuildMember, 0, guild.MemberCount)
	guild.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
		members = append(members, member)
		return true
	})

	return members, nil
}

// UpdateGuildMemberContribution 更新公会成员贡献
// 同时增加公会经验值
// 参数:
//   - playerId: 玩家ID
//   - amount: 贡献变化量
//
// 返回:
//   - error: 更新错误
func (gs *GuildService) UpdateGuildMemberContribution(playerId common.PlayerIdType, amount int64) error {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return ErrGuildNotExists
	}

	member, exists := guild.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	// 更新成员贡献和公会经验
	member.Contribution += amount
	guild.Exp += amount

	guild.Members.Store(playerId, member)
	gs.guilds.Store(guildId, guild)

	zLog.Info("Guild member contribution updated", zap.Int64("playerId", int64(playerId)), zap.Int64("amount", amount), zap.Int64("total", member.Contribution))
	return nil
}

// GetGuildMemberContribution 获取公会成员贡献
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - int64: 贡献值
//   - error: 获取错误
func (gs *GuildService) GetGuildMemberContribution(playerId common.PlayerIdType) (int64, error) {
	guildId, exists := gs.playerGuild.Load(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return 0, ErrGuildNotExists
	}

	member, exists := guild.Members.Load(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}

	return member.Contribution, nil
}

// UpgradeGuild 升级公会
// 消耗公会经验值提升等级，增加成员上限
// 参数:
//   - operatorId: 操作者ID
//
// 返回:
//   - error: 升级错误
func (gs *GuildService) UpgradeGuild(operatorId common.PlayerIdType) error {
	// 检查权限
	if err := gs.CheckGuildPermission(operatorId, GuildPermissionUpgradeGuild); err != nil {
		return err
	}

	guildId, exists := gs.playerGuild.Load(operatorId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	guild, exists := gs.guilds.Load(guildId)
	if !exists {
		return ErrGuildNotExists
	}

	// 检查公会经验是否足够
	requiredExp := int64(guild.Level * 10000)
	if guild.Exp < requiredExp {
		return errors.New("guild exp not enough")
	}

	// 扣除经验并升级
	guild.Exp -= requiredExp
	guild.Level++
	guild.MaxMembers += 10

	gs.guilds.Store(guildId, guild)

	zLog.Info("Guild upgraded", zap.Int64("guildId", int64(guildId)), zap.Int("oldLevel", guild.Level-1), zap.Int("newLevel", guild.Level))
	return nil
}
