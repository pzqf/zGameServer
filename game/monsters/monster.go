package monster

import (
	"github.com/pzqf/zGameServer/game/object"
)

// Monster 怪物类
type Monster struct {
	object.LivingObject
	aiBehavior *AIBehavior
	dropConfig *DropConfig
}

// AIBehavior 怪物AI行为
type AIBehavior struct {
	// AI状态：巡逻、追击、战斗、逃跑
	state string
	// 巡逻路径
	patrolPath []object.Vector3
	// 当前巡逻点
	currentPatrolPoint int
	// 感知范围
	perceptionRange float32
	// 追击范围
	chaseRange float32
	// 逃跑范围
	runawayRange float32
}

// DropConfig 怪物掉落配置
type DropConfig struct {
	// 掉落物品列表
	dropItems map[int32]float32 // key: 物品ID, value: 掉落概率
	// 经验值
	exp int32
	// 金币
	gold int32
}

// NewMonster 创建新的怪物对象
func NewMonster(id uint64, name string) *Monster {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建AI行为组件
	aiBehavior := &AIBehavior{
		state:           "patrol",
		perceptionRange: 10.0,
		chaseRange:      20.0,
		runawayRange:    5.0,
	}

	// 创建掉落配置组件
	dropConfig := &DropConfig{
		dropItems: make(map[int32]float32),
		exp:       100,
		gold:      50,
	}

	// 创建怪物对象
	monster := &Monster{
		LivingObject: *livingObj,
		aiBehavior:   aiBehavior,
		dropConfig:   dropConfig,
	}

	// 添加组件到游戏对象
	monster.AddComponent("ai", aiBehavior)
	monster.AddComponent("drop", dropConfig)

	return monster
}
