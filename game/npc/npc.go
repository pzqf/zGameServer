package npc

import (
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
	"github.com/pzqf/zGameServer/game/object/component"
)

// NPCType NPC类型定义
const (
	NPCTypeCommon     = 1 // 普通NPC（无特殊功能）
	NPCTypeMerchant   = 2 // 商人（出售物品）
	NPCTypeQuestGiver = 3 // 任务发布者
	NPCTypeTrainer    = 4 // 训练师（学习技能）
	NPCTypeHealer     = 5 // 治疗师（恢复生命魔法）
	NPCTypeGuard      = 6 // 守卫（攻击敌对目标）
)

// NPC 非玩家角色类
// 游戏中的NPC实体，继承自LivingObject
type NPC struct {
	*object.LivingObject
	aiBehavior   *AIBehavior        // AI行为组件
	npcType      int                // NPC类型（NPCType*）
	interaction  *InteractionSystem // 交互系统组件
	dialogueTree *DialogueTree      // 对话树组件
}

// AIBehavior NPC AI行为
// 管理NPC的AI状态和行为参数
type AIBehavior struct {
	*component.BaseComponent
	state              string           // AI状态（stationary/patrol/follow/combat）
	perceptionRange    float32          // 感知范围
	patrolPath         []object.Vector3 // 巡逻路径
	currentPatrolPoint int              // 当前巡逻点索引
	isMovable          bool             // 是否可移动
}

// Init 初始化AI行为组件
func (ai *AIBehavior) Init() error {
	return nil
}

// Update 更新AI行为组件
func (ai *AIBehavior) Update(deltaTime float64) {
}

// GetState 获取AI状态
func (ai *AIBehavior) GetState() string {
	return ai.state
}

// SetState 设置AI状态
func (ai *AIBehavior) SetState(state string) {
	ai.state = state
}

// GetPerceptionRange 获取感知范围
func (ai *AIBehavior) GetPerceptionRange() float32 {
	return ai.perceptionRange
}

// SetPerceptionRange 设置感知范围
func (ai *AIBehavior) SetPerceptionRange(range_ float32) {
	ai.perceptionRange = range_
}

// GetPatrolPath 获取巡逻路径
func (ai *AIBehavior) GetPatrolPath() []object.Vector3 {
	return ai.patrolPath
}

// SetPatrolPath 设置巡逻路径
func (ai *AIBehavior) SetPatrolPath(path []object.Vector3) {
	ai.patrolPath = path
}

// GetCurrentPatrolPoint 获取当前巡逻点
func (ai *AIBehavior) GetCurrentPatrolPoint() int {
	return ai.currentPatrolPoint
}

// SetCurrentPatrolPoint 设置当前巡逻点
func (ai *AIBehavior) SetCurrentPatrolPoint(point int) {
	ai.currentPatrolPoint = point
}

// IsMovable 检查是否可移动
func (ai *AIBehavior) IsMovable() bool {
	return ai.isMovable
}

// SetMovable 设置是否可移动
func (ai *AIBehavior) SetMovable(movable bool) {
	ai.isMovable = movable
}

// InteractionSystem 交互系统
// 管理NPC的交互功能
type InteractionSystem struct {
	*component.BaseComponent
	interactDistance float32 // 交互距离
	isInteractable   bool    // 是否可交互
	interactionType  string  // 交互类型（dialogue/trade/quest等）
}

// Init 初始化交互系统
func (isys *InteractionSystem) Init() error {
	return nil
}

// Update 更新交互系统
func (isys *InteractionSystem) Update(deltaTime float64) {
}

// GetInteractDistance 获取交互距离
func (isys *InteractionSystem) GetInteractDistance() float32 {
	return isys.interactDistance
}

// SetInteractDistance 设置交互距离
func (isys *InteractionSystem) SetInteractDistance(distance float32) {
	isys.interactDistance = distance
}

// IsInteractable 检查是否可交互
func (isys *InteractionSystem) IsInteractable() bool {
	return isys.isInteractable
}

// SetInteractable 设置是否可交互
func (isys *InteractionSystem) SetInteractable(interactable bool) {
	isys.isInteractable = interactable
}

// GetInteractionType 获取交互类型
func (isys *InteractionSystem) GetInteractionType() string {
	return isys.interactionType
}

// SetInteractionType 设置交互类型
func (isys *InteractionSystem) SetInteractionType(interactionType string) {
	isys.interactionType = interactionType
}

// DialogueTree 对话树
// 管理NPC的对话内容和选项
type DialogueTree struct {
	*component.BaseComponent
	dialogueId       int32         // 对话ID
	dialogueContent  string        // 对话内容
	dialogueOptions  []string      // 对话选项
	dialogueBranches map[int]int32 // 对话分支（选项索引 -> 下一个对话ID）
}

// Init 初始化对话树
func (dt *DialogueTree) Init() error {
	return nil
}

// Update 更新对话树
func (dt *DialogueTree) Update(deltaTime float64) {
}

// GetDialogueId 获取对话ID
func (dt *DialogueTree) GetDialogueId() int32 {
	return dt.dialogueId
}

// SetDialogueId 设置对话ID
func (dt *DialogueTree) SetDialogueId(dialogueId int32) {
	dt.dialogueId = dialogueId
}

// GetDialogueContent 获取对话内容
func (dt *DialogueTree) GetDialogueContent() string {
	return dt.dialogueContent
}

// SetDialogueContent 设置对话内容
func (dt *DialogueTree) SetDialogueContent(content string) {
	dt.dialogueContent = content
}

// GetDialogueOptions 获取对话选项
func (dt *DialogueTree) GetDialogueOptions() []string {
	return dt.dialogueOptions
}

// SetDialogueOptions 设置对话选项
func (dt *DialogueTree) SetDialogueOptions(options []string) {
	dt.dialogueOptions = options
}

// GetDialogueBranches 获取对话分支
func (dt *DialogueTree) GetDialogueBranches() map[int]int32 {
	return dt.dialogueBranches
}

// SetDialogueBranches 设置对话分支
func (dt *DialogueTree) SetDialogueBranches(branches map[int]int32) {
	dt.dialogueBranches = branches
}

// SetActive 设置对话树组件是否激活
func (dt *DialogueTree) SetActive(active bool) {
	dt.BaseComponent.SetActive(active)
}

// NewNPC 创建新的NPC对象
// 参数:
//   - id: NPC对象ID
//   - name: NPC名称
//   - npcType: NPC类型
//
// 返回: 新创建的NPC对象
func NewNPC(id common.ObjectIdType, name string, npcType int) *NPC {
	livingObj := object.NewLivingObject(id, name)

	aiBehavior := &AIBehavior{
		BaseComponent:      component.NewBaseComponent("ai"),
		state:              "stationary",
		perceptionRange:    10.0,
		patrolPath:         make([]object.Vector3, 0),
		currentPatrolPoint: 0,
		isMovable:          false,
	}

	interaction := &InteractionSystem{
		BaseComponent:    component.NewBaseComponent("interaction"),
		interactDistance: 3.0,
		isInteractable:   true,
		interactionType:  "dialogue",
	}

	dialogueTree := &DialogueTree{
		BaseComponent:    component.NewBaseComponent("dialogue"),
		dialogueId:       1,
		dialogueContent:  "欢迎来到我们的世界！",
		dialogueOptions:  []string{"你好", "再见"},
		dialogueBranches: make(map[int]int32),
	}

	npc := &NPC{
		LivingObject: livingObj,
		aiBehavior:   aiBehavior,
		npcType:      npcType,
		interaction:  interaction,
		dialogueTree: dialogueTree,
	}

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

// GetInteraction 获取交互系统组件
func (n *NPC) GetInteraction() *InteractionSystem {
	return n.interaction
}

// GetDialogueTree 获取对话树组件
func (n *NPC) GetDialogueTree() *DialogueTree {
	return n.dialogueTree
}

// Interact 与NPC交互
// 参数:
//   - player: 交互的玩家对象
func (n *NPC) Interact(player gamecommon.IGameObject) {
	if n.interaction != nil && n.interaction.isInteractable {
		// 执行交互逻辑
	}
}
