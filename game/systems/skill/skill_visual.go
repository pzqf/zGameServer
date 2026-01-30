package skill

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
)

// SkillVisualType 技能特效类型
type SkillVisualType string

const (
	SkillVisualTypeProjectile SkillVisualType = "projectile" // 投射物
	SkillVisualTypeAOE        SkillVisualType = "aoe"        // 范围效果
	SkillVisualTypeBuff       SkillVisualType = "buff"       // 增益特效
	SkillVisualTypeDebuff     SkillVisualType = "debuff"     // 减益特效
	SkillVisualTypeImpact     SkillVisualType = "impact"     // 击中特效
	SkillVisualTypeTrail      SkillVisualType = "trail"      // 轨迹特效
	SkillVisualTypeExplosion  SkillVisualType = "explosion"  // 爆炸特效
	SkillVisualTypeSummon     SkillVisualType = "summon"     // 召唤特效
	SkillVisualTypeCombo      SkillVisualType = "combo"      // 连击特效
)

// SkillVisual 技能特效
type SkillVisual struct {
	VisualID   int32                  // 特效ID
	Type       SkillVisualType        // 特效类型
	Name       string                 // 特效名称
	Duration   float32                // 持续时间（秒）
	Delay      float32                // 延迟（秒）
	Size       float32                // 大小
	Color      string                 // 颜色
	Sound      string                 // 音效
	Properties map[string]interface{} // 附加属性
}

// ActiveVisual 活跃特效
type ActiveVisual struct {
	VisualID  uint64         // 活跃特效ID
	Visual    *SkillVisual   // 特效
	OwnerID   uint64         // 所有者ID
	Position  common.Vector3 // 位置
	TargetID  uint64         // 目标ID
	StartTime time.Time      // 开始时间
	EndTime   time.Time      // 结束时间
	Progress  float32        // 进度
}

// SkillVisualSystem 技能特效系统
type SkillVisualSystem struct {
	mu             sync.RWMutex
	visuals        map[uint32]*SkillVisual  // 特效配置
	activeVisuals  map[uint64]*ActiveVisual // 活跃特效
	visualPool     *zObject.GenericPool     // 特效对象池
	activePool     *zObject.GenericPool     // 活跃特效对象池
	visualsBySkill map[int32][]*SkillVisual // 技能特效映射
}

// GlobalSkillVisualSystem 全局技能特效系统
var GlobalSkillVisualSystem *SkillVisualSystem

// init 初始化全局技能特效系统
func init() {
	GlobalSkillVisualSystem = &SkillVisualSystem{
		visuals:        make(map[uint32]*SkillVisual),
		activeVisuals:  make(map[uint64]*ActiveVisual),
		visualPool:     zObject.NewGenericPool(func() interface{} { return &SkillVisual{} }, 1000),
		activePool:     zObject.NewGenericPool(func() interface{} { return &ActiveVisual{} }, 1000),
		visualsBySkill: make(map[int32][]*SkillVisual),
	}
}

// Init 初始化技能特效系统
func (svs *SkillVisualSystem) Init() error {
	// 加载特效配置
	if err := svs.loadVisuals(); err != nil {
		return err
	}
	return nil
}

// loadVisuals 加载特效配置
func (svs *SkillVisualSystem) loadVisuals() error {
	// 从配置文件加载特效
	// 这里可以根据实际配置格式加载
	return nil
}

// AddVisual 添加特效配置
func (svs *SkillVisualSystem) AddVisual(visualID uint32, visual *SkillVisual) {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	svs.visuals[visualID] = visual
}

// AddVisualToSkill 为技能添加特效
func (svs *SkillVisualSystem) AddVisualToSkill(skillID int32, visual *SkillVisual) {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	svs.visualsBySkill[skillID] = append(svs.visualsBySkill[skillID], visual)
}

// PlaySkillVisual 播放技能特效
func (svs *SkillVisualSystem) PlaySkillVisual(ownerID uint64, skillID int32, position common.Vector3, targetID uint64) []uint64 {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	// 获取技能特效
	visuals := svs.visualsBySkill[skillID]
	if len(visuals) == 0 {
		return nil
	}

	// 播放特效
	visualIDs := make([]uint64, 0, len(visuals))
	for _, visual := range visuals {
		visualID := svs.playVisual(ownerID, visual, position, targetID)
		if visualID > 0 {
			visualIDs = append(visualIDs, visualID)
		}
	}

	return visualIDs
}

// playVisual 播放特效
func (svs *SkillVisualSystem) playVisual(ownerID uint64, visual *SkillVisual, position common.Vector3, targetID uint64) uint64 {
	activeVisual := svs.activePool.Get().(*ActiveVisual)
	visualID := uint64(time.Now().UnixNano())
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

// Update 更新技能特效系统
func (svs *SkillVisualSystem) Update() {
	svs.mu.Lock()
	defer svs.mu.Unlock()

	currentTime := time.Now()
	expiredVisuals := make([]uint64, 0)

	// 更新活跃特效
	for visualID, visual := range svs.activeVisuals {
		if currentTime.Before(visual.StartTime) {
			// 特效尚未开始
			continue
		}

		if currentTime.After(visual.EndTime) {
			// 特效已结束
			expiredVisuals = append(expiredVisuals, visualID)
			continue
		}

		// 更新进度
		elapsed := currentTime.Sub(visual.StartTime).Seconds()
		total := visual.EndTime.Sub(visual.StartTime).Seconds()
		if total > 0 {
			visual.Progress = float32(elapsed / total)
		}

		// 根据特效类型更新
		svs.updateVisualByType(visual)
	}

	// 清理过期特效
	for _, visualID := range expiredVisuals {
		visual := svs.activeVisuals[visualID]
		if visual != nil {
			svs.activePool.Put(visual)
			delete(svs.activeVisuals, visualID)
		}
	}
}

// updateVisualByType 根据特效类型更新
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

// updateProjectileVisual 更新投射物特效
func (svs *SkillVisualSystem) updateProjectileVisual(visual *ActiveVisual) {
	// 更新投射物位置
	if visual.TargetID > 0 {
		// 这里需要从对象管理器获取目标
		// 暂时简化处理
		// target := object.GetObjectByID(visual.TargetID)
		// if target != nil {
		// 	// 计算当前位置
		// 	targetPos := target.GetPosition()
		// 	startPos := visual.Position
		// 	currentPos := startPos.Lerp(targetPos, visual.Progress)
		// 	visual.Position = currentPos
		// }
	}
}

// updateAOEVisual 更新范围特效
func (svs *SkillVisualSystem) updateAOEVisual(visual *ActiveVisual) {
	// 更新范围特效大小
	visual.Visual.Size = 1.0 + visual.Progress*2.0
}

// updateTrailVisual 更新轨迹特效
func (svs *SkillVisualSystem) updateTrailVisual(visual *ActiveVisual) {
	// 更新轨迹特效
	// 这里可以添加轨迹点等
}

// GetActiveVisuals 获取活跃特效
func (svs *SkillVisualSystem) GetActiveVisuals(ownerID uint64) []*ActiveVisual {
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

// GetVisualsBySkill 获取技能特效
func (svs *SkillVisualSystem) GetVisualsBySkill(skillID int32) []*SkillVisual {
	svs.mu.RLock()
	defer svs.mu.RUnlock()

	visuals, exists := svs.visualsBySkill[skillID]
	if !exists {
		return nil
	}

	// 创建副本
	visualsCopy := make([]*SkillVisual, len(visuals))
	for i, visual := range visuals {
		visualsCopy[i] = visual
	}

	return visualsCopy
}

// CleanupExpiredVisuals 清理过期特效
func (svs *SkillVisualSystem) CleanupExpiredVisuals() {
	svs.Update()
}
