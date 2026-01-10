package player

import (
	"go.uber.org/zap"

	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zUtil/zTime"
)

// 玩家状态定义
const (
	PlayerStatusOnline  = 1
	PlayerStatusOffline = 2
	PlayerStatusBusy    = 3
)

// Player 玩家对象

type Player struct {
	playerId int64
	name     string
	Session  *zNet.TcpServerSession
	logger   *zap.Logger
	status   int
	// 玩家系统组件
	basicInfo *BasicInfo
	inventory *Inventory
	equipment *Equipment
	mailbox   *Mailbox
	tasks     *TaskManager
	skills    *SkillManager
}

// BasicInfo 玩家基础信息

type BasicInfo struct {
	Level      int
	Exp        int64
	Gold       int64
	VipLevel   int
	ServerId   int
	CreateTime int64
}

func NewPlayer(playerId int64, name string, session *zNet.TcpServerSession, logger *zap.Logger) *Player {
	player := &Player{
		playerId: playerId,
		name:     name,
		Session:  session,
		logger:   logger,
		status:   PlayerStatusOnline,
		basicInfo: &BasicInfo{
			Level:      1,
			Exp:        0,
			Gold:       1000,
			VipLevel:   0,
			ServerId:   1,
			CreateTime: zTime.Now().Time().UnixMilli(), // 设置为当前时间
		},
		inventory: NewInventory(playerId, logger),
		equipment: NewEquipment(playerId, logger),
		mailbox:   NewMailbox(playerId, logger),
		tasks:     NewTaskManager(playerId, logger),
		skills:    NewSkillManager(playerId, logger),
	}

	// 初始化玩家系统组件
	player.inventory.Init()
	player.equipment.Init()
	player.mailbox.Init()
	player.tasks.Init()
	player.skills.Init()

	return player
}

// GetPlayerId 获取玩家ID
func (p *Player) GetPlayerId() int64 {
	return p.playerId
}

// GetName 获取玩家名称
func (p *Player) GetName() string {
	return p.name
}

// GetSession 获取玩家会话
func (p *Player) GetSession() *zNet.TcpServerSession {
	return p.Session
}

// GetStatus 获取玩家状态
func (p *Player) GetStatus() int {
	return p.status
}

// SetStatus 设置玩家状态
func (p *Player) SetStatus(status int) {
	p.status = status
}

// GetBasicInfo 获取玩家基础信息
func (p *Player) GetBasicInfo() *BasicInfo {
	return p.basicInfo
}

// GetInventory 获取玩家背包
func (p *Player) GetInventory() *Inventory {
	return p.inventory
}

// GetEquipment 获取玩家装备
func (p *Player) GetEquipment() *Equipment {
	return p.equipment
}

// GetMailbox 获取玩家邮箱
func (p *Player) GetMailbox() *Mailbox {
	return p.mailbox
}

// GetTasks 获取玩家任务
func (p *Player) GetTasks() *TaskManager {
	return p.tasks
}

// GetSkills 获取玩家技能
func (p *Player) GetSkills() *SkillManager {
	return p.skills
}

// OnConnect 玩家连接成功
func (p *Player) OnConnect() {
	p.status = PlayerStatusOnline
	p.logger.Info("Player connected", zap.Int64("playerId", p.playerId), zap.String("name", p.name))
}

// OnDisconnect 玩家断开连接
func (p *Player) OnDisconnect() {
	p.status = PlayerStatusOffline
	p.logger.Info("Player disconnected", zap.Int64("playerId", p.playerId), zap.String("name", p.name))
}

// SendPacket 给玩家发送数据包
func (p *Player) SendPacket(protoId int32, data []byte) error {
	if p.Session == nil {
		return nil
	}
	return p.Session.Send(protoId, data)
}
