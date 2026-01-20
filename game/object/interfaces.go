package object

import "github.com/pzqf/zEngine/zEvent"

// IGameObject 所有游戏对象的最基本行为接口
type IGameObject interface {
	// 获取唯一标识
	GetID() uint64
	// 获取名称
	GetName() string
	// 获取位置信息
	GetPosition() Vector3
	// 设置位置
	SetPosition(pos Vector3)
	// 更新逻辑
	Update(deltaTime float64)
	// 销毁对象
	Destroy()
	// 检查是否存活
	IsActive() bool
	// 设置是否激活
	SetActive(active bool)
	// 获取事件总线
	GetEventEmitter() *zEvent.EventBus
}

// IDrawable 可绘制接口
type IDrawable interface {
	// 绘制对象
	Draw()
	// 获取渲染层级
	GetRenderLayer() int
	// 设置渲染层级
	SetRenderLayer(layer int)
}

// IMovable 可移动接口
type IMovable interface {
	// 移动到目标位置
	MoveTo(target Vector3, speed float32)
	// 停止移动
	StopMoving()
	// 获取移动速度
	GetSpeed() float32
	// 设置移动速度
	SetSpeed(speed float32)
	// 检查是否正在移动
	IsMoving() bool
	// 获取移动方向
	GetDirection() Vector3
}

// IInteractable 可交互接口
type IInteractable interface {
	// 与其他对象交互
	Interact(interactor IGameObject)
	// 检查是否可交互
	IsInteractable() bool
	// 设置交互距离
	SetInteractDistance(distance float32)
	// 获取交互距离
	GetInteractDistance() float32
}

// ILivingObject 生命对象接口
type ILivingObject interface {
	IGameObject
	// 获取生命值
	GetHealth() float32
	// 设置生命值
	SetHealth(health float32)
	// 获取最大生命值
	GetMaxHealth() float32
	// 设置最大生命值
	SetMaxHealth(maxHealth float32)
	// 受到伤害
	TakeDamage(damage float32, attacker IGameObject)
	// 回复生命值
	Heal(amount float32)
	// 死亡处理
	Die(killer IGameObject)
	// 检查是否存活
	IsAlive() bool
}

// IPropertyHolder 属性持有者接口
type IPropertyHolder interface {
	// 获取属性值
	GetProperty(key string) float32
	// 设置属性值
	SetProperty(key string, value float32)
	// 增加属性值
	AddProperty(key string, value float32)
	// 减少属性值
	SubProperty(key string, value float32)
	// 获取所有属性
	GetAllProperties() map[string]float32
}

// ICombatable 可战斗接口
type ICombatable interface {
	// 获取攻击力
	GetAttack() float32
	// 获取防御力
	GetDefense() float32
	// 开始战斗
	StartCombat(target ILivingObject)
	// 结束战斗
	EndCombat()
	// 检查是否在战斗中
	IsInCombat() bool
	// 获取当前目标
	GetCurrentTarget() ILivingObject
	// 设置当前目标
	SetCurrentTarget(target ILivingObject)
	// 获取仇恨列表
	GetHateList() map[ILivingObject]float32
	// 增加仇恨
	AddHate(target ILivingObject, amount float32)
	// 减少仇恨
	SubHate(target ILivingObject, amount float32)
	// 清除仇恨
	ClearHate()
}

// ISkillUser 技能使用者接口
type ISkillUser interface {
	// 学习技能
	LearnSkill(skillID int32)
	// 使用技能
	UseSkill(skillID int32, target IGameObject)
	// 升级技能
	UpgradeSkill(skillID int32)
	// 检查是否可以使用技能
	CanUseSkill(skillID int32) bool
}

// IHasInventory 拥有背包接口
type IHasInventory interface {
	// 获取背包容量
	GetInventoryCapacity() int
	// 增加物品
	AddItem(itemID int32, count int32) bool
	// 移除物品
	RemoveItem(itemID int32, count int32) bool
	// 检查物品数量
	HasItem(itemID int32, count int32) bool
	// 获取物品数量
	GetItemCount(itemID int32) int32
	// 获取所有物品
	GetAllItems() map[int32]int32
}
