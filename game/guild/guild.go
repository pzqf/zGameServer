package guild

import (
	"errors"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zUtil/zMap"
	"go.uber.org/zap"
)

// 公会职位定义
// 职位数值越小，权限等级越高
const (
	GuildPositionLeader = 1 // 会长（最高权限）
	GuildPositionVice   = 2 // 副会长
	GuildPositionElite  = 3 // 精英成员
	GuildPositionMember = 4 // 普通成员
)

// 公会权限定义
// 使用位掩码实现细粒度权限控制
const (
	GuildPermissionKickMember      = 1 << 0 // 踢出成员权限
	GuildPermissionInviteMember    = 1 << 1 // 邀请成员权限
	GuildPermissionUpdateNotice    = 1 << 2 // 更新公告权限
	GuildPermissionManageApply     = 1 << 3 // 管理申请权限
	GuildPermissionSetPosition     = 1 << 4 // 设置职位权限
	GuildPermissionTransferLeader  = 1 << 5 // 转让会长权限
	GuildPermissionUpgradeGuild    = 1 << 6 // 升级公会权限
	GuildPermissionManageTerritory = 1 << 7 // 管理领地权限
	GuildPermissionManageWarehouse = 1 << 8 // 管理仓库权限
)

// 公会申请状态定义
const (
	GuildApplyStatusPending  = 1 // 待处理（等待审核）
	GuildApplyStatusAccepted = 2 // 已接受（申请通过）
	GuildApplyStatusRejected = 3 // 已拒绝（申请未通过）
)

// GuildMember 公会成员
// 表示公会中的一个成员信息
type GuildMember struct {
	PlayerId     common.PlayerIdType // 成员玩家ID
	Name         string              // 成员名称
	Position     int                 // 成员职位（GuildPosition*）
	Contribution int64               // 贡献值
	JoinTime     int64               // 加入时间戳（毫秒）
	Online       bool                // 是否在线
	LastOnline   int64               // 最后在线时间戳（毫秒）
}

// GuildApply 公会申请
// 表示玩家申请加入公会的记录
type GuildApply struct {
	ApplyId    int64                // 申请记录ID
	PlayerId   common.PlayerIdType  // 申请人ID
	PlayerName string               // 申请人名称
	GuildId    common.GuildIdType   // 目标公会ID
	ApplyTime  int64                // 申请时间戳（毫秒）
	Status     int                  // 申请状态（GuildApplyStatus*）
	Remark     string               // 申请备注
}

// Guild 公会结构
// 表示一个完整的公会实体
type Guild struct {
	GuildId          common.GuildIdType                                   // 公会ID
	Name             string                                               // 公会名称
	Level            int                                                  // 公会等级
	Exp              int64                                                // 公会经验值
	LeaderId         common.PlayerIdType                                  // 会长ID
	MemberCount      int                                                  // 当前成员数
	MaxMembers       int                                                  // 最大成员数
	Members          *zMap.TypedShardedMap[common.PlayerIdType, *GuildMember] // 成员映射表（PlayerId -> GuildMember）
	Applies          *zMap.TypedShardedMap[int64, *GuildApply]            // 申请映射表（ApplyId -> GuildApply）
	CreateTime       int64                                                // 创建时间戳（毫秒）
	Notice           string                                               // 公会公告
	WarScore         int64                                                // 公会战积分
	Territories      []int                                                // 拥有的领地列表
	PermissionConfig map[int]int64                                        // 权限配置（Position -> PermissionMask）
}

// 自定义错误类型
var (
	ErrPermissionDenied = errors.New("permission denied")    // 权限不足
	ErrPlayerNotInGuild = errors.New("player not in guild")  // 玩家不在公会中
	ErrGuildNotExists   = errors.New("guild not exists")     // 公会不存在
)

// AddMember 添加成员到公会
// 参数:
//   - playerId: 玩家ID
//   - playerName: 玩家名称
//   - position: 成员职位
//
// 返回:
//   - error: 添加错误
func (g *Guild) AddMember(playerId common.PlayerIdType, playerName string, position int) error {
	// 检查成员是否已存在
	if _, exists := g.Members.Load(playerId); exists {
		return nil
	}

	// 检查成员数上限
	if g.MemberCount >= g.MaxMembers {
		return nil
	}

	currentTime := time.Now().UnixMilli()

	guildMember := &GuildMember{
		PlayerId:     playerId,
		Name:         playerName,
		Position:     position,
		Contribution: 0,
		JoinTime:     currentTime,
		Online:       true,
		LastOnline:   currentTime,
	}

	g.Members.Store(playerId, guildMember)
	g.MemberCount++

	return nil
}

// RemoveMember 从公会移除成员
// 如果移除的是会长，自动转让会长职位
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - error: 移除错误
func (g *Guild) RemoveMember(playerId common.PlayerIdType) error {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	// 如果移除的是会长，先转让会长职位
	if member.Position == GuildPositionLeader {
		g.transferGuildLeader(playerId)
	}

	g.Members.Delete(playerId)
	g.MemberCount--

	return nil
}

// transferGuildLeader 转让会长职位
// 按优先级选择新会长：副会长 -> 精英成员 -> 普通成员
// 参数:
//   - oldLeaderId: 旧会长ID
func (g *Guild) transferGuildLeader(oldLeaderId common.PlayerIdType) {
	var newLeader *GuildMember
	var newLeaderId common.PlayerIdType

	// 先找副会长
	g.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
		if member.Position == GuildPositionVice {
			newLeader = member
			newLeaderId = member.PlayerId
			return false
		}
		return true
	})

	// 如果没有副会长，找精英成员
	if newLeader == nil {
		g.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
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
		g.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
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
		newLeader.Position = GuildPositionLeader
		g.Members.Store(common.PlayerIdType(newLeaderId), newLeader)
		g.LeaderId = common.PlayerIdType(newLeaderId)

		// 更新旧会长的职位为普通成员
		if oldLeader, exists := g.Members.Load(common.PlayerIdType(oldLeaderId)); exists {
			oldLeader.Position = GuildPositionMember
			g.Members.Store(common.PlayerIdType(oldLeaderId), oldLeader)
		}

		zLog.Info("Guild leader transferred", zap.Int64("guildId", int64(g.GuildId)), zap.Int64("oldLeaderId", int64(oldLeaderId)), zap.Int64("newLeaderId", int64(newLeaderId)))
	}
}

// SetMemberPosition 设置公会成员职位
// 参数:
//   - playerId: 玩家ID
//   - newPosition: 新职位
//
// 返回:
//   - error: 设置错误
func (g *Guild) SetMemberPosition(playerId int64, newPosition int) error {
	pId := common.PlayerIdType(playerId)
	member, exists := g.Members.Load(pId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	// 如果设置为会长，需要处理原会长
	if newPosition == GuildPositionLeader {
		if member.Position == GuildPositionLeader {
			return nil
		}

		// 将原会长降为普通成员
		if oldLeader, exists := g.Members.Load(g.LeaderId); exists {
			oldLeader.Position = GuildPositionMember
			g.Members.Store(g.LeaderId, oldLeader)
		}

		g.LeaderId = pId
	}

	member.Position = newPosition
	g.Members.Store(pId, member)

	return nil
}

// UpdateNotice 更新公会公告
// 参数:
//   - notice: 新公告内容
func (g *Guild) UpdateNotice(notice string) {
	g.Notice = notice
}

// GetMember 获取公会成员信息
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - *GuildMember: 成员信息
//   - bool: 是否存在
func (g *Guild) GetMember(playerId int64) (*GuildMember, bool) {
	return g.Members.Load(common.PlayerIdType(playerId))
}

// GetMembers 获取公会所有成员
// 返回:
//   - []*GuildMember: 成员列表
func (g *Guild) GetMembers() []*GuildMember {
	members := make([]*GuildMember, 0, g.MemberCount)
	g.Members.Range(func(key common.PlayerIdType, member *GuildMember) bool {
		members = append(members, member)
		return true
	})
	return members
}

// UpdateMemberContribution 更新公会成员贡献
// 同时增加公会经验值
// 参数:
//   - playerId: 玩家ID
//   - amount: 贡献变化量
//
// 返回:
//   - error: 更新错误
func (g *Guild) UpdateMemberContribution(playerId common.PlayerIdType, amount int64) error {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	member.Contribution += amount
	g.Exp += amount

	g.Members.Store(playerId, member)

	return nil
}

// GetMemberContribution 获取公会成员贡献
// 参数:
//   - playerId: 玩家ID
//
// 返回:
//   - int64: 贡献值
//   - error: 获取错误
func (g *Guild) GetMemberContribution(playerId common.PlayerIdType) (int64, error) {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}

	return member.Contribution, nil
}

// Upgrade 升级公会
// 消耗公会经验值提升等级，增加成员上限
// 返回:
//   - error: 升级错误
func (g *Guild) Upgrade() error {
	// 计算升级所需经验（等级 * 10000）
	requiredExp := int64(g.Level * 10000)
	if g.Exp < requiredExp {
		return errors.New("guild exp not enough")
	}

	// 扣除经验
	g.Exp -= requiredExp
	// 升级
	g.Level++
	// 增加最大成员数（每级+10）
	g.MaxMembers += 10

	return nil
}

// CheckPermission 检查玩家是否有指定权限
// 使用位运算检查权限掩码
// 参数:
//   - playerId: 玩家ID
//   - permission: 要求的权限位
//
// 返回:
//   - error: 权限不足错误
func (g *Guild) CheckPermission(playerId common.PlayerIdType, permission int64) error {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	// 检查权限位
	if g.PermissionConfig[member.Position]&permission == 0 {
		return ErrPermissionDenied
	}

	return nil
}

// AddApply 添加公会申请
// 参数:
//   - apply: 申请记录
func (g *Guild) AddApply(apply *GuildApply) {
	g.Applies.Store(apply.ApplyId, apply)
}

// ProcessApply 处理公会申请
// 参数:
//   - applyId: 申请ID
//   - accept: 是否接受
//
// 返回:
//   - error: 处理错误
func (g *Guild) ProcessApply(applyId int64, accept bool) error {
	apply, exists := g.Applies.Load(applyId)
	if !exists {
		return errors.New("application not exists")
	}

	apply.Status = GuildApplyStatusRejected
	if accept {
		// 检查公会是否已满
		if g.MemberCount >= g.MaxMembers {
			return errors.New("guild is full")
		}

		g.AddMember(apply.PlayerId, apply.PlayerName, GuildPositionMember)
		apply.Status = GuildApplyStatusAccepted
	}

	g.Applies.Store(applyId, apply)

	return nil
}

// GetApplies 获取公会所有申请
// 返回:
//   - []*GuildApply: 申请列表
func (g *Guild) GetApplies() []*GuildApply {
	applies := make([]*GuildApply, 0)
	g.Applies.Range(func(key int64, apply *GuildApply) bool {
		applies = append(applies, apply)
		return true
	})
	return applies
}

// UpdateMemberOnlineStatus 更新公会成员在线状态
// 参数:
//   - playerId: 玩家ID
//   - online: 是否在线
//
// 返回:
//   - error: 更新错误
func (g *Guild) UpdateMemberOnlineStatus(playerId common.PlayerIdType, online bool) error {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}

	member.Online = online
	if !online {
		member.LastOnline = time.Now().UnixMilli()
	}

	g.Members.Store(playerId, member)

	return nil
}
