package guild

import (
	"errors"
	"time"

	"github.com/pzqf/zEngine/zLog"
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
	PlayerId     int64
	Name         string
	Position     int
	Contribution int64
	JoinTime     int64
	Online       bool
	LastOnline   int64
}

// GuildApply 公会申请
type GuildApply struct {
	ApplyId    int64
	PlayerId   int64
	PlayerName string
	GuildId    int64
	ApplyTime  int64
	Status     int
	Remark     string
}

// Guild 公会结构
type Guild struct {
	GuildId          int64
	Name             string
	Level            int
	Exp              int64
	LeaderId         int64
	MemberCount      int
	MaxMembers       int
	Members          *zMap.ShardedMap // key: int64(playerId), value: *GuildMember
	Applies          *zMap.ShardedMap // key: int64(applyId), value: *GuildApply
	CreateTime       int64
	Notice           string
	WarScore         int64         // 公会战积分
	Territories      []int         // 占领的领地
	PermissionConfig map[int]int64 // key: position, value: permission mask
}

// 自定义错误类型
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrPlayerNotInGuild = errors.New("player not in guild")
	ErrGuildNotExists   = errors.New("guild not exists")
)

// AddMember 添加成员到公会
func (g *Guild) AddMember(playerId int64, playerName string, position int) error {
	// 检查玩家是否已在公会中
	if _, exists := g.Members.Load(playerId); exists {
		return nil // 玩家已在公会中
	}

	// 检查公会是否已满
	if g.MemberCount >= g.MaxMembers {
		return nil // 公会已满
	}

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 创建公会成员
	guildMember := &GuildMember{
		PlayerId:     playerId,
		Name:         playerName,
		Position:     position,
		Contribution: 0,
		JoinTime:     currentTime,
		Online:       true,
		LastOnline:   currentTime,
	}

	// 添加到公会成员列表
	g.Members.Store(playerId, guildMember)
	g.MemberCount++

	return nil
}

// RemoveMember 从公会移除成员
func (g *Guild) RemoveMember(playerId int64) error {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 如果是会长，需要处理会长转让
	if member.Position == GuildPositionLeader {
		// 寻找新的会长（副会长 -> 精英 -> 普通成员）
		g.transferGuildLeader(playerId)
	}

	// 从公会成员中移除
	g.Members.Delete(playerId)
	g.MemberCount--

	return nil
}

// transferGuildLeader 转让会长职位
func (g *Guild) transferGuildLeader(oldLeaderId int64) {
	// 寻找新的会长（副会长 -> 精英 -> 普通成员）
	var newLeader *GuildMember
	var newLeaderId int64

	// 先找副会长
	g.Members.Range(func(key, value interface{}) bool {
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
		g.Members.Range(func(key, value interface{}) bool {
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
		g.Members.Range(func(key, value interface{}) bool {
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
		g.Members.Store(newLeaderId, newLeader)
		g.LeaderId = newLeaderId

		// 更新旧会长的职位为普通成员
		if memberInterface, exists := g.Members.Load(oldLeaderId); exists {
			oldLeader := memberInterface.(*GuildMember)
			oldLeader.Position = GuildPositionMember
			g.Members.Store(oldLeaderId, oldLeader)
		}

		zLog.Info("Guild leader transferred", zap.Int64("guildId", g.GuildId), zap.Int64("oldLeaderId", oldLeaderId), zap.Int64("newLeaderId", newLeaderId))
	}
}

// SetMemberPosition 设置公会成员职位
func (g *Guild) SetMemberPosition(playerId int64, newPosition int) error {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 如果设置为会长，需要处理旧会长
	if newPosition == GuildPositionLeader {
		// 如果目标玩家已经是会长，不需要处理
		if member.Position == GuildPositionLeader {
			return nil
		}

		// 获取旧会长
		oldLeaderInterface, exists := g.Members.Load(g.LeaderId)
		if exists {
			oldLeader := oldLeaderInterface.(*GuildMember)
			// 将旧会长职位改为普通成员
			oldLeader.Position = GuildPositionMember
			g.Members.Store(g.LeaderId, oldLeader)
		}

		// 更新公会的leaderId
		g.LeaderId = playerId
	}

	// 更新目标玩家职位
	member.Position = newPosition
	g.Members.Store(playerId, member)

	return nil
}

// UpdateNotice 更新公会公告
func (g *Guild) UpdateNotice(notice string) {
	g.Notice = notice
}

// GetMember 获取公会成员信息
func (g *Guild) GetMember(playerId int64) (*GuildMember, bool) {
	member, exists := g.Members.Load(playerId)
	if !exists {
		return nil, false
	}
	return member.(*GuildMember), true
}

// GetMembers 获取公会所有成员
func (g *Guild) GetMembers() []*GuildMember {
	members := make([]*GuildMember, 0, g.MemberCount)
	g.Members.Range(func(key, value interface{}) bool {
		members = append(members, value.(*GuildMember))
		return true
	})
	return members
}

// UpdateMemberContribution 更新公会成员贡献
func (g *Guild) UpdateMemberContribution(playerId int64, amount int64) error {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 更新贡献
	member.Contribution += amount
	g.Exp += amount // 公会经验也增加相应的贡献值

	// 存储更新后的成员信息
	g.Members.Store(playerId, member)

	return nil
}

// GetMemberContribution 获取公会成员贡献
func (g *Guild) GetMemberContribution(playerId int64) (int64, error) {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return 0, ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	return member.Contribution, nil
}

// Upgrade 升级公会
func (g *Guild) Upgrade() error {
	// 计算升级所需经验
	requiredExp := int64(g.Level * 10000)
	if g.Exp < requiredExp {
		return errors.New("guild exp not enough")
	}

	// 扣除经验
	g.Exp -= requiredExp
	// 升级
	g.Level++
	// 增加最大成员数
	g.MaxMembers += 10

	return nil
}

// CheckPermission 检查玩家是否有指定权限
func (g *Guild) CheckPermission(playerId int64, permission int64) error {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 检查权限（基于权限掩码）
	if g.PermissionConfig[member.Position]&permission == 0 {
		return ErrPermissionDenied
	}

	return nil
}

// AddApply 添加公会申请
func (g *Guild) AddApply(apply *GuildApply) {
	g.Applies.Store(apply.ApplyId, apply)
}

// ProcessApply 处理公会申请
func (g *Guild) ProcessApply(applyId int64, accept bool) error {
	// 获取申请
	applyInterface, exists := g.Applies.Load(applyId)
	if !exists {
		return errors.New("application not exists")
	}
	apply := applyInterface.(*GuildApply)

	// 更新申请状态
	apply.Status = GuildApplyStatusRejected
	if accept {
		// 检查公会是否已满
		if g.MemberCount >= g.MaxMembers {
			return errors.New("guild is full")
		}

		// 接受申请，加入公会
		g.AddMember(apply.PlayerId, apply.PlayerName, GuildPositionMember)
		apply.Status = GuildApplyStatusAccepted
	}

	// 保存更新后的申请状态
	g.Applies.Store(applyId, apply)

	return nil
}

// GetApplies 获取公会所有申请
func (g *Guild) GetApplies() []*GuildApply {
	applies := make([]*GuildApply, 0)
	g.Applies.Range(func(key, value interface{}) bool {
		applies = append(applies, value.(*GuildApply))
		return true
	})
	return applies
}

// UpdateMemberOnlineStatus 更新公会成员在线状态
func (g *Guild) UpdateMemberOnlineStatus(playerId int64, online bool) error {
	// 检查玩家是否在公会中
	memberInterface, exists := g.Members.Load(playerId)
	if !exists {
		return ErrPlayerNotInGuild
	}
	member := memberInterface.(*GuildMember)

	// 更新在线状态
	member.Online = online
	if !online {
		member.LastOnline = time.Now().UnixMilli()
	}

	// 存储更新后的成员信息
	g.Members.Store(playerId, member)

	return nil
}
