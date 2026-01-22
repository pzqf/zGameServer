package guild

import (
	"errors"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// GuildService 公会服务
type GuildService struct {
	zObject.BaseObject
	guilds       *zMap.Map // key: int64(guildId), value: *Guild
	playerGuild  *zMap.Map // key: int64(playerId), value: int64(guildId)
	guildNameMap *zMap.Map // key: string(guildName), value: int64(guildId)
	maxGuilds    int
}

func NewGuildService() *GuildService {
	gs := &GuildService{
		guilds:       zMap.NewMap(),
		playerGuild:  zMap.NewMap(),
		guildNameMap: zMap.NewMap(),
		maxGuilds:    1000,
	}
	gs.BaseObject.Id = "GuildService"
	return gs
}

func (gs *GuildService) Init() error {
	zLog.Info("Initializing guild service...")
	// 初始化公会服务相关资源
	return nil
}

func (gs *GuildService) Close() error {
	zLog.Info("Closing guild service...")
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

	// 创建新公会
	guild := &Guild{
		GuildId:          guildId,
		Name:             guildName,
		Level:            1,
		Exp:              0,
		LeaderId:         leaderId,
		MemberCount:      1,
		MaxMembers:       20,
		Members:          zMap.NewMap(),
		Applies:          zMap.NewMap(),
		CreateTime:       currentTime,
		Notice:           "欢迎加入公会！",
		WarScore:         0,
		Territories:      []int{},
		PermissionConfig: permissionConfig,
	}

	// 添加会长到公会成员
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

	// 存储公会信息
	gs.guilds.Store(guildId, guild)
	gs.playerGuild.Store(leaderId, guildId)
	gs.guildNameMap.Store(guildName, guildId)

	zLog.Info("Guild created", zap.Int64("guildId", guildId), zap.String("guildName", guildName), zap.Int64("leaderId", leaderId))
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

	// 调用公会的AddMember方法添加成员
	if err := guild.AddMember(playerId, playerName, GuildPositionMember); err != nil {
		return err
	}

	// 存储玩家公会关系
	gs.playerGuild.Store(playerId, guildId)

	zLog.Info("Player joined guild", zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
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

	// 调用公会的RemoveMember方法移除成员
	if err := guild.RemoveMember(playerId); err != nil {
		return err
	}

	// 如果公会成员为0，解散公会
	if guild.MemberCount <= 0 {
		gs.DisbandGuild(guildId)
		return nil
	}

	// 移除玩家公会关系
	gs.playerGuild.Delete(playerId)

	zLog.Info("Player left guild", zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
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
	guild.Members.Range(func(key, value interface{}) bool {
		playerId := key.(int64)
		gs.playerGuild.Delete(playerId)
		return true
	})

	// 移除公会名称映射
	gs.guildNameMap.Delete(guild.Name)

	// 从公会列表中移除
	gs.guilds.Delete(guildId)

	zLog.Info("Guild disbanded", zap.Int64("guildId", guildId), zap.String("guildName", guild.Name))
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
		ApplyId:    applyId,
		PlayerId:   playerId,
		PlayerName: playerName,
		GuildId:    guildId,
		ApplyTime:  currentTime,
		Status:     GuildApplyStatusPending,
		Remark:     remark,
	}

	// 获取公会
	guildInterface, _ := gs.guilds.Get(guildId)
	guild := guildInterface.(*Guild)

	// 存储申请
	guild.Applies.Store(applyId, apply)

	zLog.Info("Guild application submitted", zap.Int64("applyId", applyId), zap.Int64("playerId", playerId), zap.Int64("guildId", guildId))
	return nil
}

// transferGuildLeader 转让会长职位
func (gs *GuildService) transferGuildLeader(guild *Guild, oldLeaderId int64) {
	// 寻找新的会长（副会长 -> 精英 -> 普通成员）
	var newLeader *GuildMember
	var newLeaderId int64

	// 先找副会长
	guild.Members.Range(func(key, value interface{}) bool {
		member := value.(*GuildMember)
		if member.Position == GuildPositionVice {
			newLeader = member
			newLeaderId = member.PlayerId
			return false
		}
		return true
	})

	// 如果没有副会长，找精英成员
	if newLeader == nil {
		guild.Members.Range(func(key, value interface{}) bool {
			member := value.(*GuildMember)
			if member.Position == GuildPositionElite {
				newLeader = member
				newLeaderId = member.PlayerId
				return false
			}
			return true
		})
	}

	// 如果没有精英成员，找普通成员
	if newLeader == nil {
		guild.Members.Range(func(key, value interface{}) bool {
			member := value.(*GuildMember)
			if member.PlayerId != oldLeaderId {
				newLeader = member
				newLeaderId = member.PlayerId
				return false
			}
			return true
		})
	}

	// 如果找到新会长，更新职位
	if newLeader != nil {
		// 更新新会长的职位
		newLeader.Position = GuildPositionLeader
		guild.Members.Store(newLeaderId, newLeader)
		guild.LeaderId = newLeaderId

		// 更新旧会长的职位为普通成员
		if memberInterface, exists := guild.Members.Get(oldLeaderId); exists {
			oldLeader := memberInterface.(*GuildMember)
			oldLeader.Position = GuildPositionMember
			guild.Members.Store(oldLeaderId, oldLeader)
		}

		zLog.Info("Guild leader transferred", zap.Int64("guildId", guild.GuildId), zap.Int64("oldLeaderId", oldLeaderId), zap.Int64("newLeaderId", newLeaderId))
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

	// 调用公会的SetMemberPosition方法设置职位
	if err := guild.SetMemberPosition(targetPlayerId, newPosition); err != nil {
		return err
	}

	zLog.Info("Guild member position updated", zap.Int64("guildId", guildId), zap.Int64("playerId", targetPlayerId), zap.Int("newPosition", newPosition))
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

	// 调用公会的UpdateNotice方法更新公告
	guild.UpdateNotice(notice)

	zLog.Info("Guild notice updated", zap.Int64("guildId", guildId), zap.String("notice", notice))
	return nil
}

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
	memberInterface, exists := guild.Members.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 检查权限（职位数字越小，权限越高）
	if member.Position > requiredPosition {
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

	// 调用公会的CheckPermission方法检查权限
	return guild.CheckPermission(playerId, permission)
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
		if a, exists := g.Applies.Get(applyId); exists {
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
	apply.Status = GuildApplyStatusRejected
	if accept {
		// 检查公会是否已满
		if guild.MemberCount >= guild.MaxMembers {
			return errors.New("guild is full")
		}

		// 接受申请，加入公会
		gs.JoinGuild(apply.PlayerId, apply.PlayerName, apply.GuildId)
		apply.Status = GuildApplyStatusAccepted
	}

	// 保存更新后的申请状态
	guild.Applies.Store(applyId, apply)

	// 更新申请
	guild.Applies.Store(applyId, apply)

	zLog.Info("Guild application processed", zap.Int64("applyId", applyId), zap.Bool("accepted", accept), zap.Int64("playerId", apply.PlayerId), zap.Int64("guildId", apply.GuildId))
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
	memberInterface, exists := guild.Members.Get(playerId)
	if !exists {
		return nil // 玩家不在公会中
	}
	member := memberInterface.(*GuildMember)

	// 更新在线状态
	member.Online = online
	if !online {
		member.LastOnline = time.Now().UnixMilli()
	}

	// 存储更新后的成员信息
	guild.Members.Store(playerId, member)

	zLog.Debug("Guild member online status updated", zap.Int64("playerId", playerId), zap.Bool("online", online))
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
	members := make([]*GuildMember, 0, guild.MemberCount)
	guild.Members.Range(func(key, value interface{}) bool {
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
	memberInterface, exists := guild.Members.Get(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 更新贡献
	member.Contribution += amount
	guild.Exp += amount // 公会经验也增加相应的贡献值

	// 存储更新后的成员信息
	guild.Members.Store(playerId, member)
	gs.guilds.Store(guildId, guild)

	zLog.Info("Guild member contribution updated", zap.Int64("playerId", playerId), zap.Int64("amount", amount), zap.Int64("total", member.Contribution))
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
	memberInterface, exists := guild.Members.Get(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	return member.Contribution, nil
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
	requiredExp := int64(guild.Level * 10000)
	if guild.Exp < requiredExp {
		return errors.New("guild exp not enough")
	}

	// 扣除经验
	guild.Exp -= requiredExp
	// 升级
	guild.Level++
	// 增加最大成员数
	guild.MaxMembers += 10

	// 存储更新后的公会信息
	gs.guilds.Store(guildId, guild)

	zLog.Info("Guild upgraded", zap.Int64("guildId", guildId), zap.Int("oldLevel", guild.Level-1), zap.Int("newLevel", guild.Level))
	return nil
}
