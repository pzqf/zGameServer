package monster

import (
	"math"

	"github.com/pzqf/zEngine/zScript"
	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/object"
)

const (
	AttackRange     float32 = 2.5
	ChaseRange      float32 = 15.0
	PerceptionRange float32 = 10.0
)

func init() {
	RegisterMonsterAIFunctions()
}

func RegisterMonsterAIFunctions() {
	zScript.RegisterScriptFunc(IsPlayerInRange)
	zScript.RegisterScriptFunc(IsAttackInRange)
	zScript.RegisterScriptFunc(BasicAttack)
	zScript.RegisterScriptFunc(IsOutOfTrackingRange)
	zScript.RegisterScriptFunc(ReturnToHome)
	zScript.RegisterScriptFunc(HealToFull)
	zScript.RegisterScriptFunc(PatrolToNextPoint)
	zScript.RegisterScriptFunc(MoveToPlayer)
}

func IsPlayerInRange(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	distance := monster.GetDistanceToNearestPlayer()
	return distance <= monster.GetAIBehavior().GetPerceptionRange()
}

func IsAttackInRange(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	distance := monster.GetDistanceToTarget()
	return distance <= monster.GetAIBehavior().GetAttackRange()
}

func BasicAttack(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	target := monster.GetTarget()
	if target == nil {
		return false
	}

	monster.Attack(target)
	return true
}

func IsOutOfTrackingRange(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	distance := monster.GetDistanceToTarget()
	return distance > monster.GetAIBehavior().GetChaseRange()
}

func ReturnToHome(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	monster.ReturnToHome()
	return true
}

func HealToFull(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	monster.HealToFull()
	return true
}

func PatrolToNextPoint(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	monster.PatrolToNextPoint()
	return true
}

func MoveToPlayer(holder *zScript.ScriptHolder, args ...interface{}) interface{} {
	monster := holder.GetContext().(*Monster)
	if monster == nil {
		return false
	}

	target := monster.GetTarget()
	if target == nil {
		return false
	}

	monster.MoveToTarget(target)
	return true
}

func (m *Monster) GetContext() *Monster {
	return m
}

func (m *Monster) GetDistanceToNearestPlayer() float32 {
	minDistance := float32(1000000.0)
	return minDistance
}

func (m *Monster) GetDistanceToTarget() float32 {
	target := m.GetTarget()
	if target == nil {
		return 1000000.0
	}

	return m.GetDistanceToObject(target)
}

func (m *Monster) GetDistanceToObject(obj common.IGameObject) float32 {
	if obj == nil {
		return 1000000.0
	}

	mPos := m.GetPosition()
	oPos := obj.GetPosition()

	dx := mPos.X - oPos.X
	dz := mPos.Z - oPos.Z

	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}

func (m *Monster) GetTarget() common.IGameObject {
	return nil
}

func (m *Monster) Attack(target common.IGameObject) {
}

func (m *Monster) ReturnToHome() {
	homePos := m.GetHomePosition()
	m.MoveTo(common.Vector3{
		X: homePos.X,
		Y: homePos.Y,
		Z: homePos.Z,
	})
}

func (m *Monster) HealToFull() {
}

func (m *Monster) PatrolToNextPoint() {
	ai := m.GetAIBehavior()
	path := ai.GetPatrolPath()

	if len(path) == 0 {
		return
	}

	currentPoint := ai.GetCurrentPatrolPoint()
	nextPoint := (currentPoint + 1) % len(path)

	ai.SetCurrentPatrolPoint(nextPoint)

	targetPos := path[nextPoint]
	m.MoveTo(common.Vector3{
		X: targetPos.X,
		Y: targetPos.Y,
		Z: targetPos.Z,
	})
}

func (m *Monster) MoveToTarget(target common.IGameObject) {
	if target == nil {
		return
	}

	targetPos := target.GetPosition()
	m.MoveTo(targetPos)
}

func (m *Monster) GetHomePosition() object.Vector3 {
	return object.Vector3{}
}
