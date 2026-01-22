package npc

import (
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// NPCType NPC类型定义
const (
	NPCTypeCommon     = 1 // 普通NPC
	NPCTypeMerchant   = 2 // 商人
	NPCTypeQuestGiver = 3 // 任务发布者
	NPCTypeTrainer    = 4 // 训练师
	NPCTypeHealer     = 5 // 治疗师
	NPCTypeGuard      = 6 // 守卫
)

// NPC 非玩家角色类
type NPC struct {
	object.LivingObject
	aiBehavior   *AIBehavior
	npcType      int
	interaction  *InteractionSystem
	dialogueTree *DialogueTree
}

// AIBehavior NPC AI行为
type AIBehavior struct {
	*component.BaseComponent
	// AI状态：静止、巡逻、跟随、战斗
	state string
	// 感知范围
	perceptionRange float32
	// 巡逻路径
	patrolPath []object.Vector3
	// 当前巡逻点
	currentPatrolPoint int
	// 是否可移动
	isMovable bool
}

// Init 初始化AI行为组件
func (ai *AIBehavior) Init() error {
	return nil
}

// Update 更新AI行为组件
func (ai *AIBehavior) Update(deltaTime float64) {
}

// Destroy 销毁AI行为组件
func (ai *AIBehavior) Destroy() {
}

// IsActive 检查AI行为组件是否激活
func (ai *AIBehavior) IsActive() bool {
	return ai.BaseComponent.IsActive()
}

// SetActive 设置AI行为组件是否激活
func (ai *AIBehavior) SetActive(active bool) {
	ai.BaseComponent.SetActive(active)
}

// InteractionSystem NPC交互系统
type InteractionSystem struct {
	*component.BaseComponent
	// 交互距离
	interactDistance float32
	// 可交互状态
	isInteractable bool
	// 交互类型：对话、交易、任务
	interactionType string
}

// Init 初始化交互系统组件
func (is *InteractionSystem) Init() error {
	return nil
}

// Update 更新交互系统组件
func (is *InteractionSystem) Update(deltaTime float64) {
}

// Destroy 销毁交互系统组件
func (is *InteractionSystem) Destroy() {
}

// IsActive 检查交互系统组件是否激活
func (is *InteractionSystem) IsActive() bool {
	return is.BaseComponent.IsActive()
}

// SetActive 设置交互系统组件是否激活
func (is *InteractionSystem) SetActive(active bool) {
	is.BaseComponent.SetActive(active)
}

// DialogueTree NPC对话树
type DialogueTree struct {
	*component.BaseComponent
	// 对话ID
	dialogueId int32
	// 对话内容
	dialogueContent string
	// 对话选项
	dialogueOptions []string
	// 对话分支
	dialogueBranches map[int]int32
}

// Init 初始化对话树组件
func (dt *DialogueTree) Init() error {
	return nil
}

// Update 更新对话树组件
func (dt *DialogueTree) Update(deltaTime float64) {
}

// Destroy 销毁对话树组件
func (dt *DialogueTree) Destroy() {
}

// IsActive 检查对话树组件是否激活
func (dt *DialogueTree) IsActive() bool {
	return dt.BaseComponent.IsActive()
}

// SetActive 设置对话树组件是否激活
func (dt *DialogueTree) SetActive(active bool) {
	dt.BaseComponent.SetActive(active)
}

// NewNPC 创建新的NPC对象
func NewNPC(id uint64, name string, npcType int) *NPC {
	// 创建基础生命对象
	livingObj := object.NewLivingObject(id, name)

	// 创建AI行为组件
	aiBehavior := &AIBehavior{
		BaseComponent:      component.NewBaseComponent("ai"),
		state:              "stationary", // 默认静止状态
		perceptionRange:    10.0,
		patrolPath:         make([]object.Vector3, 0),
		currentPatrolPoint: 0,
		isMovable:          false,
	}

	// 创建交互系统
	interaction := &InteractionSystem{
		BaseComponent:    component.NewBaseComponent("interaction"),
		interactDistance: 3.0,
		isInteractable:   true,
		interactionType:  "dialogue",
	}

	// 创建对话树
	dialogueTree := &DialogueTree{
		BaseComponent:    component.NewBaseComponent("dialogue"),
		dialogueId:       1,
		dialogueContent:  "欢迎来到我们的世界！",
		dialogueOptions:  []string{"你好", "再见"},
		dialogueBranches: make(map[int]int32),
	}

	// 创建NPC对象
	npc := &NPC{
		LivingObject: *livingObj,
		aiBehavior:   aiBehavior,
		npcType:      npcType,
		interaction:  interaction,
		dialogueTree: dialogueTree,
	}

	// 添加组件到游戏对象
	npc.AddComponentWithName("ai", aiBehavior)
	npc.AddComponentWithName("interaction", interaction)
	npc.AddComponentWithName("dialogue", dialogueTree)

	return npc
}

// GetNPCType 获取NPC类型
func (n *NPC) GetNPCType() int {
	return n.npcType
}

// SetAIState 设置AI状态
func (n *NPC) SetAIState(state string) {
	n.aiBehavior.state = state
}

// GetAIState 获取AI状态
func (n *NPC) GetAIState() string {
	return n.aiBehavior.state
}

// GetType 获取NPC类型
func (n *NPC) GetType() int {
	return object.GameObjectTypeNPC
}

// SetPatrolPath 设置巡逻路径
func (n *NPC) SetPatrolPath(path []object.Vector3) {
	n.aiBehavior.patrolPath = path
	n.aiBehavior.isMovable = true
}

// StartPatrol 开始巡逻
func (n *NPC) StartPatrol() {
	if n.aiBehavior.isMovable && len(n.aiBehavior.patrolPath) > 0 {
		n.aiBehavior.state = "patrol"
	}
}

// StopPatrol 停止巡逻
func (n *NPC) StopPatrol() {
	if n.aiBehavior.state == "patrol" {
		n.aiBehavior.state = "stationary"
	}
}

// Interact 与NPC交互
func (n *NPC) Interact(interactor common.IGameObject) {
	// TODO: 实现交互逻辑
}

// GetDialogueTree 获取对话树
func (n *NPC) GetDialogueTree() *DialogueTree {
	return n.dialogueTree
}

// SetDialogueTree 设置对话树
func (n *NPC) SetDialogueTree(dialogueTree *DialogueTree) {
	n.dialogueTree = dialogueTree
}

// Update 更新NPC状态
func (n *NPC) Update(deltaTime float64) {
	// 调用父类更新
	n.LivingObject.Update(deltaTime)

	// 更新AI行为
	n.updateAI(deltaTime)
}

// updateAI 更新AI行为
func (n *NPC) updateAI(deltaTime float64) {
	// TODO: 实现AI逻辑
}
