package movement

import (
	"sync"

	"github.com/pzqf/zGameServer/game/object"
)

// MovementState 移动状态
type MovementState struct {
	ownerID     uint64
	startPos    object.Vector3
	targetPos   object.Vector3
	direction   object.Vector3
	distance    float32
	speed       float32
	isMoving    bool
	progress    float32
	totalTime   float32
	timeElapsed float32
}

// MovementSystem 移动系统
type MovementSystem struct {
	mu             sync.RWMutex
	movementStates map[uint64]*MovementState
}

// GlobalMovementSystem 全局移动系统实例
var GlobalMovementSystem *MovementSystem

// init 初始化全局移动系统
func init() {
	GlobalMovementSystem = &MovementSystem{
		movementStates: make(map[uint64]*MovementState),
	}
}

// MoveTo 移动到目标位置
func (ms *MovementSystem) MoveTo(ownerID uint64, target object.Vector3, speed float32, currentPos object.Vector3) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 如果速度为0，不移动
	if speed <= 0 {
		return
	}

	// 计算移动距离
	distance := currentPos.Distance(target)

	// 如果距离为0，不移动
	if distance <= 0.01 {
		if state, exists := ms.movementStates[ownerID]; exists {
			state.isMoving = false
		}
		return
	}

	// 设置移动参数
	state := &MovementState{
		ownerID:     ownerID,
		startPos:    currentPos,
		targetPos:   target,
		distance:    distance,
		speed:       speed,
		direction:   target.Sub(currentPos).Normalize(),
		isMoving:    true,
		progress:    0,
		timeElapsed: 0,
		totalTime:   distance / speed,
	}

	ms.movementStates[ownerID] = state
}

// StopMoving 停止移动
func (ms *MovementSystem) StopMoving(ownerID uint64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if state, exists := ms.movementStates[ownerID]; exists {
		state.isMoving = false
	}
}

// IsMoving 检查是否正在移动
func (ms *MovementSystem) IsMoving(ownerID uint64) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if state, exists := ms.movementStates[ownerID]; exists {
		return state.isMoving
	}
	return false
}

// GetCurrentPosition 获取当前位置
func (ms *MovementSystem) GetCurrentPosition(ownerID uint64, basePos object.Vector3) object.Vector3 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if state, exists := ms.movementStates[ownerID]; exists && state.isMoving {
		// 计算当前位置
		currentProgress := state.timeElapsed / state.totalTime
		if currentProgress > 1.0 {
			currentProgress = 1.0
		}
		return state.startPos.Add(state.direction.Mul(state.distance * currentProgress))
	}
	return basePos
}

// Update 更新移动状态
func (ms *MovementSystem) Update(deltaTime float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, state := range ms.movementStates {
		if !state.isMoving {
			continue
		}

		// 更新移动时间
		state.timeElapsed += float32(deltaTime)

		// 计算移动进度
		state.progress = state.timeElapsed / state.totalTime

		// 检查是否到达目标
		if state.progress >= 1.0 {
			state.isMoving = false
			state.progress = 1.0
		}
	}
}
