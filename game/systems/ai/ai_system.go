package ai

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zEngine/zSystem"
	"github.com/pzqf/zGameServer/common"
	"github.com/pzqf/zGameServer/config/tables"
	gamecommon "github.com/pzqf/zGameServer/game/common"
)

type AIActionState int

const (
	AIActionStateIdle AIActionState = iota
	AIActionStatePatrolling
	AIActionStateChasing
	AIActionStateAttacking
	AIActionStateFleeing
	AIActionStateTalking
)

type AIType int

const (
	AITypeMonster AIType = iota
	AITypeNPC
)

type AI struct {
	ID             int32
	Type           AIType
	State          AIActionState
	TargetID       common.ObjectIdType
	PatrolPoints   []gamecommon.Vector3
	CurrentPoint   int
	DetectionRange float32
	AttackRange    float32
	ChaseRange     float32
	FleeHealth     float32
	LastAction     time.Time
}

type AIState struct {
	ownerID common.ObjectIdType
	ai      *AI
}

type AISystem struct {
	*zSystem.BaseSystem
	mu        sync.RWMutex
	aiStates  map[common.ObjectIdType]*AIState
	statePool *zObject.GenericPool
	aiPool    *zObject.GenericPool
}

func NewAISystem() *AISystem {
	return &AISystem{
		BaseSystem: zSystem.NewBaseSystem("AISystem"),
		aiStates:   make(map[common.ObjectIdType]*AIState),
		statePool:  zObject.NewGenericPool(func() interface{} { return &AIState{} }, 1000),
		aiPool:     zObject.NewGenericPool(func() interface{} { return &AI{} }, 5000),
	}
}

func (as *AISystem) InitAI(ownerID common.ObjectIdType, aiID int32, aiType AIType, detectionRange, attackRange, chaseRange, fleeHealth float32) {
	as.mu.Lock()
	defer as.mu.Unlock()

	aiState, exists := as.aiStates[ownerID]
	if !exists {
		aiState = as.statePool.Get().(*AIState)
		aiState.ownerID = ownerID
		as.aiStates[ownerID] = aiState
	}

	ai := as.aiPool.Get().(*AI)
	ai.ID = aiID
	ai.Type = aiType
	ai.State = AIActionStateIdle
	ai.TargetID = 0
	ai.PatrolPoints = make([]gamecommon.Vector3, 0)
	ai.CurrentPoint = 0
	ai.DetectionRange = detectionRange
	ai.AttackRange = attackRange
	ai.ChaseRange = chaseRange
	ai.FleeHealth = fleeHealth
	ai.LastAction = time.Now()

	aiState.ai = ai
}

func (as *AISystem) InitAIFromConfig(ownerID common.ObjectIdType, aiID int32) {
	aiConfig := tables.GetAIByID(aiID)
	if aiConfig == nil {
		return
	}

	as.InitAI(
		ownerID,
		aiConfig.AIID,
		AITypeMonster,
		aiConfig.DetectionRange,
		aiConfig.AttackRange,
		aiConfig.ChaseRange,
		aiConfig.FleeHealth,
	)

	patrolPoints := as.ParsePatrolPoints(aiConfig.PatrolPoints)

	if len(patrolPoints) > 0 {
		as.SetPatrolPoints(ownerID, patrolPoints)
	}
}

func (as *AISystem) ParsePatrolPoints(patrolPointsStr string) []gamecommon.Vector3 {
	if patrolPointsStr == "" {
		return nil
	}

	points := make([]gamecommon.Vector3, 0)
	pointStrs := strings.Split(patrolPointsStr, ";")

	for _, pointStr := range pointStrs {
		coords := strings.Split(pointStr, ",")
		if len(coords) != 3 {
			continue
		}

		x, err := strconv.ParseFloat(coords[0], 32)
		if err != nil {
			continue
		}

		y, err := strconv.ParseFloat(coords[1], 32)
		if err != nil {
			continue
		}

		z, err := strconv.ParseFloat(coords[2], 32)
		if err != nil {
			continue
		}

		points = append(points, gamecommon.Vector3{
			X: float32(x),
			Y: float32(y),
			Z: float32(z),
		})
	}

	return points
}

func (as *AISystem) SetPatrolPoints(ownerID common.ObjectIdType, points []gamecommon.Vector3) {
	as.mu.Lock()
	defer as.mu.Unlock()

	aiState, exists := as.aiStates[ownerID]
	if !exists {
		return
	}

	if aiState.ai == nil {
		return
	}

	aiState.ai.PatrolPoints = points
	aiState.ai.CurrentPoint = 0

	if len(points) > 0 {
		aiState.ai.State = AIActionStatePatrolling
	}
}

func (as *AISystem) Initialize() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.aiStates = make(map[common.ObjectIdType]*AIState)
	as.statePool = zObject.NewGenericPool(func() interface{} { return &AIState{} }, 1000)
	as.aiPool = zObject.NewGenericPool(func() interface{} { return &AI{} }, 5000)

	return nil
}

func (as *AISystem) Update(deltaTime float64) {
	as.UpdateAI()
}

func (as *AISystem) UpdateAI() {
	as.mu.RLock()
	ownerIDs := make([]common.ObjectIdType, 0, len(as.aiStates))
	for ownerID := range as.aiStates {
		ownerIDs = append(ownerIDs, ownerID)
	}
	as.mu.RUnlock()

	for _, ownerID := range ownerIDs {
		as.updateAIForOwner(ownerID)
	}
}

func (as *AISystem) updateAIForOwner(ownerID common.ObjectIdType) {
	as.mu.Lock()
	defer as.mu.Unlock()

	aiState, exists := as.aiStates[ownerID]
	if !exists || aiState.ai == nil {
		return
	}

	ai := aiState.ai
	currentTime := time.Now()

	switch ai.Type {
	case AITypeMonster:
		as.updateMonsterAI(ownerID, ai, currentTime)
	case AITypeNPC:
		as.updateNPCAI(ownerID, ai, currentTime)
	}
}

func (as *AISystem) updateMonsterAI(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	switch ai.State {
	case AIActionStateIdle:
		as.handleIdleState(ownerID, ai, currentTime)
	case AIActionStatePatrolling:
		as.handlePatrollingState(ownerID, ai, currentTime)
	case AIActionStateChasing:
		as.handleChasingState(ownerID, ai, currentTime)
	case AIActionStateAttacking:
		as.handleAttackingState(ownerID, ai, currentTime)
	case AIActionStateFleeing:
		as.handleFleeingState(ownerID, ai, currentTime)
	}
}

func (as *AISystem) updateNPCAI(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	if ai.State == AIActionStateIdle {
	}
}

func (as *AISystem) handleIdleState(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	targetID := as.SelectTarget(ownerID, ai)
	if targetID > 0 {
		ai.TargetID = targetID
		ai.State = AIActionStateChasing
		ai.LastAction = currentTime
		return
	}

	if len(ai.PatrolPoints) > 0 {
		ai.State = AIActionStatePatrolling
		ai.CurrentPoint = 0
		ai.LastAction = currentTime
	}
}

func (as *AISystem) handlePatrollingState(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	targetID := as.SelectTarget(ownerID, ai)
	if targetID > 0 {
		ai.TargetID = targetID
		ai.State = AIActionStateChasing
		ai.LastAction = currentTime
		return
	}

	if len(ai.PatrolPoints) == 0 {
		ai.State = AIActionStateIdle
		ai.LastAction = currentTime
		return
	}

	obj := as.getGameObject(ownerID)
	if obj == nil {
		return
	}

	startPos := obj.GetPosition()

	currentPoint := ai.PatrolPoints[ai.CurrentPoint]

	movementComponent := obj.GetComponent("movement")
	if movementComponent != nil {
		if move, ok := movementComponent.(interface {
			MoveTo(targetPos gamecommon.Vector3, speed float32)
		}); ok {
			move.MoveTo(currentPoint, 2.0)
		}
	}

	if startPos.DistanceTo(currentPoint) < 1.0 {
		ai.CurrentPoint = (ai.CurrentPoint + 1) % len(ai.PatrolPoints)
	}

	ai.LastAction = currentTime
}

func (as *AISystem) handleChasingState(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	if ai.TargetID == 0 {
		ai.State = AIActionStateIdle
		ai.LastAction = currentTime
		return
	}

	obj := as.getGameObject(ownerID)
	target := as.getGameObject(ai.TargetID)

	if obj == nil || target == nil {
		ai.State = AIActionStateIdle
		ai.TargetID = 0
		ai.LastAction = currentTime
		return
	}

	startPos := obj.GetPosition()
	targetPos := target.GetPosition()

	if startPos.DistanceTo(targetPos) > ai.ChaseRange {
		ai.State = AIActionStateIdle
		ai.TargetID = 0
		ai.LastAction = currentTime
		return
	}

	movementComponent := obj.GetComponent("movement")
	if movementComponent != nil {
		if move, ok := movementComponent.(interface {
			MoveTo(targetPos gamecommon.Vector3, speed float32)
		}); ok {
			move.MoveTo(targetPos, 3.0)
		}
	}

	if startPos.DistanceTo(targetPos) <= ai.AttackRange {
		ai.State = AIActionStateAttacking
		ai.LastAction = currentTime
	}

	ai.LastAction = currentTime
}

func (as *AISystem) handleAttackingState(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	if ai.TargetID == 0 {
		ai.State = AIActionStateIdle
		ai.LastAction = currentTime
		return
	}

	obj := as.getGameObject(ownerID)
	target := as.getGameObject(ai.TargetID)

	if obj == nil || target == nil {
		ai.State = AIActionStateIdle
		ai.TargetID = 0
		ai.LastAction = currentTime
		return
	}

	startPos := obj.GetPosition()
	targetPos := target.GetPosition()

	if startPos.DistanceTo(targetPos) > ai.AttackRange {
		ai.State = AIActionStateChasing
		ai.LastAction = currentTime
		return
	}

	combatComponent := obj.GetComponent("combat")
	if combatComponent != nil {
		if combat, ok := combatComponent.(interface {
			StartCombat(targetID common.ObjectIdType)
		}); ok {
			combat.StartCombat(ai.TargetID)
		}
	}

	if ai.FleeHealth > 0 {
		propertyComponent := obj.GetComponent("property")
		var currentHealth, maxHealth float32
		if propertyComponent != nil {
			if prop, ok := propertyComponent.(interface{ GetProperty(name string) float32 }); ok {
				currentHealth = prop.GetProperty("hp")
				maxHealth = prop.GetProperty("max_hp")
			}
		}

		healthPercent := currentHealth / maxHealth
		if healthPercent < ai.FleeHealth {
			ai.State = AIActionStateFleeing
			ai.LastAction = currentTime
		}
	}

	ai.LastAction = currentTime
}

func (as *AISystem) handleFleeingState(ownerID common.ObjectIdType, ai *AI, currentTime time.Time) {
	obj := as.getGameObject(ownerID)
	if obj == nil {
		ai.State = AIActionStateIdle
		ai.TargetID = 0
		ai.LastAction = currentTime
		return
	}

	startPos := obj.GetPosition()
	var targetPos gamecommon.Vector3

	if ai.TargetID > 0 {
		target := as.getGameObject(ai.TargetID)
		if target != nil {
			targetPos = target.GetPosition()
		} else {
			targetPos = startPos.Add(gamecommon.Vector3{X: 1, Y: 1, Z: 0})
		}
	} else {
		targetPos = startPos.Add(gamecommon.Vector3{X: 1, Y: 1, Z: 0})
	}

	direction := startPos.Subtract(targetPos).Normalize()
	fleePos := startPos.Add(direction.MultiplyScalar(10.0))

	movementComponent := obj.GetComponent("movement")
	if movementComponent != nil {
		if move, ok := movementComponent.(interface {
			MoveTo(targetPos gamecommon.Vector3, speed float32)
		}); ok {
			move.MoveTo(fleePos, 4.0)
		}
	}

	if ai.TargetID > 0 {
		target := as.getGameObject(ai.TargetID)
		if target != nil {
			currentDistance := startPos.DistanceTo(target.GetPosition())
			if currentDistance > ai.ChaseRange {
				ai.State = AIActionStateIdle
				ai.TargetID = 0
				ai.LastAction = currentTime
				return
			}
		}
	}

	ai.LastAction = currentTime
}

func (as *AISystem) SelectTarget(ownerID common.ObjectIdType, ai *AI) common.ObjectIdType {
	obj := as.getGameObject(ownerID)
	if obj == nil {
		return 0
	}

	startPos := obj.GetPosition()

	mapObj := obj.GetMap()
	if mapObj == nil {
		return 0
	}

	objects := mapObj.GetObjectsInRange(startPos, ai.DetectionRange)

	for _, target := range objects {
		if target.GetType() == gamecommon.GameObjectTypePlayer {
			if startPos.DistanceTo(target.GetPosition()) <= ai.DetectionRange*ai.DetectionRange {
				return target.GetID()
			}
		}
	}

	return 0
}

func (as *AISystem) Shutdown() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	for _, state := range as.aiStates {
		if state.ai != nil {
			as.aiPool.Put(state.ai)
		}
		if state != nil {
			as.statePool.Put(state)
		}
	}

	as.aiStates = nil
	return nil
}

func (as *AISystem) getGameObject(objectID common.ObjectIdType) gamecommon.IGameObject {
	return nil
}
