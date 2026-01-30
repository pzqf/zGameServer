package movement

import (
	"sync"

	"github.com/pzqf/zGameServer/game/common"
)

// MovementState 移动状态
type MovementState struct {
	mu          sync.RWMutex
	startPos    common.Vector3
	targetPos   common.Vector3
	direction   common.Vector3
	distance    float32
	speed       float32
	isMoving    bool
	progress    float32
	totalTime   float32
	timeElapsed float32
}

func NewMovementState() *MovementState {
	return &MovementState{}
}

func (state *MovementState) StartMove(startPos, targetPos common.Vector3, speed float32) {
	if speed <= 0 {
		return
	}

	distance := startPos.DistanceTo(targetPos)
	if distance <= 0.01 {
		state.mu.Lock()
		defer state.mu.Unlock()
		state.isMoving = false
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	state.startPos = startPos
	state.targetPos = targetPos
	state.distance = distance
	state.speed = speed
	state.direction = targetPos.Subtract(startPos).Normalize()
	state.isMoving = true
	state.progress = 0
	state.timeElapsed = 0
	state.totalTime = distance / speed
}

func (state *MovementState) StopMove() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.isMoving = false
}

func (state *MovementState) IsMoving() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.isMoving
}

func (state *MovementState) GetCurrentPosition(basePos common.Vector3) common.Vector3 {
	state.mu.RLock()
	defer state.mu.RUnlock()

	if state.isMoving {
		currentProgress := state.timeElapsed / state.totalTime
		if currentProgress > 1.0 {
			currentProgress = 1.0
		}
		return state.startPos.Add(state.direction.MultiplyScalar(state.distance * currentProgress))
	}
	return basePos
}

func (state *MovementState) Update(deltaTime float64) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.isMoving {
		return
	}

	state.timeElapsed += float32(deltaTime)
	state.progress = state.timeElapsed / state.totalTime

	if state.progress >= 1.0 {
		state.isMoving = false
		state.progress = 1.0
	}
}

// MovementComponent 移动组件
type MovementComponent struct {
	mu            sync.RWMutex
	movementState *MovementState
	owner         common.IGameObject
}

func NewMovementComponent(owner common.IGameObject) *MovementComponent {
	return &MovementComponent{
		movementState: NewMovementState(),
		owner:         owner,
	}
}

func (mc *MovementComponent) MoveTo(targetPos common.Vector3, speed float32) {
	if speed <= 0 || mc.owner == nil {
		return
	}

	currentPos := mc.owner.GetPosition()
	mc.movementState.StartMove(currentPos, targetPos, speed)
}

func (mc *MovementComponent) StopMoving() {
	mc.movementState.StopMove()
}

func (mc *MovementComponent) IsMoving() bool {
	return mc.movementState.IsMoving()
}

func (mc *MovementComponent) GetCurrentPosition(basePos common.Vector3) common.Vector3 {
	return mc.movementState.GetCurrentPosition(basePos)
}

func (mc *MovementComponent) Update(deltaTime float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.movementState.Update(deltaTime)
}

func (mc *MovementComponent) GetSpeed() float32 {
	if mc.owner == nil {
		return 0
	}
	if propertyComponent := mc.getOwnerPropertyComponent(); propertyComponent != nil {
		if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
			return prop.GetProperty("move_speed")
		}
	}
	return 0
}

func (mc *MovementComponent) getOwnerPropertyComponent() interface{} {
	if mc.owner == nil {
		return nil
	}
	return mc.owner.GetComponent("property")
}
