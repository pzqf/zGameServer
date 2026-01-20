package pet

import (
	"github.com/pzqf/zGameServer/game/object"
)

// PetGrowthSystem 宠物成长系统
type PetGrowthSystem struct {
	level      int32
	exp        int32
	maxExp     int32
	growthRate float32
}

// IntimacySystem 宠物亲密度系统
type IntimacySystem struct {
	intimacy    int32
	maxIntimacy int32
	mood        string
}

// 宠物类
type Pet struct {
	object.LivingObject
	petGrowth *PetGrowthSystem
	intimacy  *IntimacySystem
}

// NewPet 创建新的宠物对象
func NewPet(id uint64, name string) *Pet {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建宠物成长系统
	growthSystem := &PetGrowthSystem{
		level:      1,
		exp:        0,
		maxExp:     100,
		growthRate: 1.2,
	}

	// 创建宠物亲密度系统
	intimacySystem := &IntimacySystem{
		intimacy:    50,
		maxIntimacy: 100,
		mood:        "happy",
	}

	// 创建宠物对象
	pet := &Pet{
		LivingObject: *livingObj,
		petGrowth:    growthSystem,
		intimacy:     intimacySystem,
	}

	// 添加组件到游戏对象
	pet.AddComponent("growth", growthSystem)
	pet.AddComponent("intimacy", intimacySystem)

	return pet
}
