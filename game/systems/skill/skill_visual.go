package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
)

type SkillVisualType string

const (
	SkillVisualTypeProjectile SkillVisualType = "projectile"
	SkillVisualTypeAOE        SkillVisualType = "aoe"
	SkillVisualTypeBuff       SkillVisualType = "buff"
	SkillVisualTypeDebuff     SkillVisualType = "debuff"
	SkillVisualTypeImpact     SkillVisualType = "impact"
	SkillVisualTypeTrail      SkillVisualType = "trail"
	SkillVisualTypeExplosion  SkillVisualType = "explosion"
	SkillVisualTypeSummon     SkillVisualType = "summon"
	SkillVisualTypeCombo      SkillVisualType = "combo"
)

type SkillVisual struct {
	VisualID   int32
	Type       SkillVisualType
	Name       string
	Duration   float32
	Delay      float32
	Size       float32
	Color      string
	Sound      string
	Properties map[string]interface{}
}

type ActiveVisual struct {
	VisualID  common.VisualIdType
	Visual    *SkillVisual
	OwnerID   common.ObjectIdType
	Position  gamecommon.Vector3
	TargetID  common.ObjectIdType
	StartTime time.Time
	EndTime   time.Time
	Progress  float32
}

type SkillVisualSystem struct {
	mu             sync.RWMutex
	visuals        map[uint32]*SkillVisual
	activeVisuals  map[common.VisualIdType]*ActiveVisual
	visualPool     *zObject.GenericPool
	activePool     *zObject.GenericPool
	visualsBySkill map[int32][]*SkillVisual
}

var GlobalSkillVisualSystem *SkillVisualSystem

func init() {
	GlobalSkillVisualSystem = &SkillVisualSystem{
		visuals:        make(map[uint32]*SkillVisual),
		activeVisuals:  make(map[common.VisualIdType]*ActiveVisual),
		visualPool:     zObject.NewGenericPool(func() interface{} { return &SkillVisual{} }, 1000),
		activePool:     zObject.NewGenericPool(func() interface{} { return &ActiveVisual{} }, 1000),
		visualsBySkill: make(map[int32][]*SkillVisual),
	}
}

func (svs *SkillVisualSystem) Init() error {
	if err := svs.loadVisuals(); err != nil {
		return err
	}
	return nil
}

func (svs *SkillVisualSystem) loadVisuals() error {
	return nil
}

func (svs *SkillVisualSystem) AddVisual(visualID uint32, visual *SkillVisual) {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	svs.visuals[visualID] = visual
}

func (svs *SkillVisualSystem) AddVisualToSkill(skillID int32, visual *SkillVisual) {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	svs.visualsBySkill[skillID] = append(svs.visualsBySkill[skillID], visual)
}

func (svs *SkillVisualSystem) PlaySkillVisual(ownerID common.ObjectIdType, skillID int32, position gamecommon.Vector3, targetID common.ObjectIdType) []common.VisualIdType {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	visuals := svs.visualsBySkill[skillID]
	if len(visuals) == 0 {
		return nil
	}

	visualIDs := make([]common.VisualIdType, 0, len(visuals))
	for _, visual := range visuals {
		visualID := svs.playVisual(ownerID, visual, position, targetID)
		if visualID > 0 {
			visualIDs = append(visualIDs, visualID)
		}
	}

	return visualIDs
}

func (svs *SkillVisualSystem) playVisual(ownerID common.ObjectIdType, visual *SkillVisual, position gamecommon.Vector3, targetID common.ObjectIdType) common.VisualIdType {
	activeVisual := svs.activePool.Get().(*ActiveVisual)
	visualID, err := common.GenerateVisualID()
	if err != nil {
		return 0
	}
	activeVisual.VisualID = visualID
	activeVisual.Visual = visual
	activeVisual.OwnerID = ownerID
	activeVisual.Position = position
	activeVisual.TargetID = targetID
	activeVisual.StartTime = time.Now().Add(time.Duration(visual.Delay * float32(time.Second)))
	activeVisual.EndTime = activeVisual.StartTime.Add(time.Duration(visual.Duration * float32(time.Second)))
	activeVisual.Progress = 0

	svs.activeVisuals[visualID] = activeVisual
	return visualID
}

func (svs *SkillVisualSystem) Update() {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	currentTime := time.Now()
	expiredVisuals := make([]common.VisualIdType, 0)

	for visualID, visual := range svs.activeVisuals {
		if currentTime.Before(visual.StartTime) {
			continue
		}

		if currentTime.After(visual.EndTime) {
			expiredVisuals = append(expiredVisuals, visualID)
			continue
		}

		elapsed := currentTime.Sub(visual.StartTime).Seconds()
		total := visual.EndTime.Sub(visual.StartTime).Seconds()
		if total > 0 {
			visual.Progress = float32(elapsed / total)
		}

		svs.updateVisualByType(visual)
	}

	for _, visualID := range expiredVisuals {
		visual := svs.activeVisuals[visualID]
		if visual != nil {
			svs.activePool.Put(visual)
			delete(svs.activeVisuals, visualID)
		}
	}
}

func (svs *SkillVisualSystem) updateVisualByType(visual *ActiveVisual) {
	switch visual.Visual.Type {
	case SkillVisualTypeProjectile:
		svs.updateProjectileVisual(visual)
	case SkillVisualTypeAOE:
		svs.updateAOEVisual(visual)
	case SkillVisualTypeTrail:
		svs.updateTrailVisual(visual)
	}
}

func (svs *SkillVisualSystem) updateProjectileVisual(visual *ActiveVisual) {
	if visual.TargetID > 0 {
	}
}

func (svs *SkillVisualSystem) updateAOEVisual(visual *ActiveVisual) {
	visual.Visual.Size = 1.0 + visual.Progress*2.0
}

func (svs *SkillVisualSystem) updateTrailVisual(visual *ActiveVisual) {
}

func (svs *SkillVisualSystem) GetActiveVisuals(ownerID common.ObjectIdType) []*ActiveVisual {
	svs.mu.RLock()
	defer svs.mu.RUnlock()

	visuals := make([]*ActiveVisual, 0)
	for _, visual := range svs.activeVisuals {
		if visual.OwnerID == ownerID {
			visuals = append(visuals, visual)
		}
	}

	return visuals
}

func (svs *SkillVisualSystem) GetVisualsBySkill(skillID int32) []*SkillVisual {
	svs.mu.RLock()
	defer svs.mu.RUnlock()

	visuals, exists := svs.visualsBySkill[skillID]
	if !exists {
		return nil
	}

	visualsCopy := make([]*SkillVisual, len(visuals))
	for i, visual := range visuals {
		visualsCopy[i] = visual
	}

	return visualsCopy
}

func (svs *SkillVisualSystem) CleanupExpiredVisuals() {
	svs.Update()
}
