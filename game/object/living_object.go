package object

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pzqf/zGameServer/game/common"
	"github.com/pzqf/zGameServer/game/maps"
)

const (
	DefaultHealth float32 = 100
	DefaultMana   float32 = 50
)

type Property struct {
	Name  string
	Value float32
}

type LivingObject struct {
	*GameObject
	mu         sync.RWMutex
	health     float32
	maxHealth  float32
	mana       float32
	maxMana    float32
	properties map[string]float32
	inCombat   bool
	targetID   uint64
	lastAttack time.Time
}

func NewLivingObject(id common.ObjectIdType, name string) *LivingObject {
	goObj := NewGameObject(id, name)
	goObj.SetType(common.GameObjectTypeLiving)

	livingObj := &LivingObject{
		GameObject: goObj,
		health:     DefaultHealth,
		maxHealth:  DefaultHealth,
		mana:       DefaultMana,
		maxMana:    DefaultMana,
		properties: make(map[string]float32),
		inCombat:   false,
		targetID:   0,
	}

	return livingObj
}

func (lo *LivingObject) GetHealth() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.health
}

func (lo *LivingObject) SetHealth(health float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	if health > lo.maxHealth {
		health = lo.maxHealth
	}
	lo.health = health
}

func (lo *LivingObject) GetMaxHealth() float32 {
	return lo.maxHealth
}

func (lo *LivingObject) SetMaxHealth(maxHealth float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.maxHealth = maxHealth
}

func (lo *LivingObject) GetMana() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.mana
}

func (lo *LivingObject) SetMana(mana float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	if mana > lo.maxMana {
		mana = lo.maxMana
	}
	lo.mana = mana
}

func (lo *LivingObject) GetMaxMana() float32 {
	return lo.maxMana
}

func (lo *LivingObject) SetMaxMana(maxMana float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.maxMana = maxMana
}

func (lo *LivingObject) GetProperty(name string) float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.properties[name]
}

func (lo *LivingObject) SetProperty(name string, value float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties[name] = value
}

func (lo *LivingObject) GetAllProperties() map[string]float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	props := make(map[string]float32)
	for k, v := range lo.properties {
		props[k] = v
	}
	return props
}

func (lo *LivingObject) SetProperties(props map[string]float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties = props
}

func (lo *LivingObject) RemoveProperty(name string) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	delete(lo.properties, name)
}

func (lo *LivingObject) ClearAllProperties() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.properties = make(map[string]float32)
}

func (lo *LivingObject) ResetAllProperties() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	for k := range lo.properties {
		delete(lo.properties, k)
	}
}

func (lo *LivingObject) TakeDamage(damage float32, attacker common.IGameObject) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	if attacker != nil {
		if lo.isDodge() {
			return
		}

		if livingAttacker, ok := attacker.(*LivingObject); ok {
			if livingAttacker.isCritical() {
				damage = damage * (1 + livingAttacker.GetProperty("critical_damage"))
			}
		}
	}

	attack := damage
	defense := lo.GetProperty("physical_defense")
	actualDamage := attack - defense*0.5

	if actualDamage < 1 {
		actualDamage = 1
	}

	lo.health -= actualDamage

	if lo.health <= 0 {
		lo.health = 0
	}
}

func (lo *LivingObject) Heal(amount float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	lo.health += amount
	if lo.health > lo.maxHealth {
		lo.health = lo.maxHealth
	}
}

func (lo *LivingObject) RegenMana(amount float32) {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health <= 0 {
		return
	}

	lo.mana += amount
	if lo.mana > lo.maxMana {
		lo.mana = lo.maxMana
	}
}

func (lo *LivingObject) Revive() {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.health > 0 {
		return
	}

	lo.health = lo.maxHealth * 0.3
	lo.mana = lo.maxMana * 0.3
	lo.inCombat = false
	lo.targetID = 0
}

func (lo *LivingObject) Teleport(targetPos common.Vector3) error {
	return lo.GameObject.Teleport(targetPos)
}

func (lo *LivingObject) SetTarget(targetID uint64) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.targetID = targetID
}

func (lo *LivingObject) GetTarget() uint64 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.targetID
}

func (lo *LivingObject) ClearTarget() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.targetID = 0
	lo.inCombat = false
}

func (lo *LivingObject) IsInCombat() bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.inCombat
}

func (lo *LivingObject) SetInCombat(inCombat bool) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = inCombat
}

func (lo *LivingObject) StartCombat() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = true
}

func (lo *LivingObject) EndCombat() {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.inCombat = false
	lo.targetID = 0
}

func (lo *LivingObject) HasTarget() bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	return lo.targetID != 0
}

func (lo *LivingObject) GetNearestTarget(objectType common.GameObjectType) common.IGameObject {
	mapObject := lo.GetMap()
	if mapObject == nil {
		return nil
	}

	position := lo.GetPosition()
	nearestObject := common.IGameObject(nil)
	minDistance := float32(0)

	if concreteMap, ok := mapObject.(*maps.Map); ok {
		for _, obj := range concreteMap.GetObjectsByType(objectType) {
			distance := position.DistanceTo(obj.GetPosition())
			if nearestObject == nil || distance < minDistance {
				nearestObject = obj
				minDistance = distance
			}
		}
	}

	return nearestObject
}

func (lo *LivingObject) IsInRangeWithTarget(target *LivingObject, radius float32) bool {
	lo.mu.RLock()
	defer lo.mu.RUnlock()

	distance := lo.GetPosition().DistanceTo(target.GetPosition())
	return distance <= radius*radius
}

func (lo *LivingObject) GetSpeed() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	speed := lo.properties["move_speed"]
	if speed <= 0 {
		speed = 3.0
	}
	return speed
}

func (lo *LivingObject) GetRange() float32 {
	lo.mu.RLock()
	defer lo.mu.RUnlock()
	rangeValue := lo.properties["attack_range"]
	if rangeValue <= 0 {
		rangeValue = 2.5
	}
	return rangeValue
}

func (lo *LivingObject) isDodge() bool {
	dodgeChance := lo.GetProperty("dodge")
	if dodgeChance < 0 {
		dodgeChance = 0
	}
	return dodgeChance > rand.Float32()
}

func (lo *LivingObject) isCritical() bool {
	criticalChance := lo.GetProperty("critical_rate")
	if criticalChance < 0 {
		criticalChance = 0
	}
	return criticalChance > rand.Float32()
}

func (lo *LivingObject) Update(deltaTime float64) {
	lo.GameObject.Update(deltaTime)

	now := time.Now()
	cooldown := 1.0 / float64(lo.GetSpeed())
	if now.Sub(lo.lastAttack).Seconds() < cooldown {
		return
	}

	lo.lastAttack = now
}
