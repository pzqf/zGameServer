package ai

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/game/common"
)

// VisionGrid 视野网格
type VisionGrid struct {
	GridSize   float32                  // 网格大小
	GridCells  map[int64]map[int64]bool // 网格单元格
	LastUpdate time.Time                // 最后更新时间
	UpdateRate time.Duration            // 更新频率
}

// PathNode 路径节点
type PathNode struct {
	Position common.Vector3 // 位置
	Parent   *PathNode      // 父节点
	G        float32        // 从起点到当前节点的代价
	H        float32        // 从当前节点到目标的估计代价
	F        float32        // 总代价
}

// GroupAI 群体AI
type GroupAI struct {
	GroupID    uint64          // 群体ID
	Members    map[uint64]bool // 成员列表
	LeaderID   uint64          // 领导者ID
	TargetID   uint64          // 群体目标
	Formation  string          // 阵型
	LastAction time.Time       // 最后行动时间
	Behavior   string          // 群体行为
}

// AIPerformanceManager AI性能管理器
type AIPerformanceManager struct {
	mu             sync.RWMutex
	visionGrids    map[uint64]*VisionGrid // 视野网格
	groupAIs       map[uint64]*GroupAI    // 群体AI
	pathNodePool   *zObject.GenericPool   // 路径节点对象池
	visionGridPool *zObject.GenericPool   // 视野网格对象池
	groupAIPool    *zObject.GenericPool   // 群体AI对象池
}

// NewAIPerformanceManager 创建AI性能管理器
func NewAIPerformanceManager() *AIPerformanceManager {
	return &AIPerformanceManager{
		visionGrids:    make(map[uint64]*VisionGrid),
		groupAIs:       make(map[uint64]*GroupAI),
		pathNodePool:   zObject.NewGenericPool(func() interface{} { return &PathNode{} }, 1000),
		visionGridPool: zObject.NewGenericPool(func() interface{} { return &VisionGrid{} }, 100),
		groupAIPool:    zObject.NewGenericPool(func() interface{} { return &GroupAI{} }, 100),
	}
}

// Init 初始化AI性能管理器
func (apm *AIPerformanceManager) Init() error {
	return nil
}

// GetVisionGrid 获取视野网格
func (apm *AIPerformanceManager) GetVisionGrid(instanceID uint64) *VisionGrid {
	apm.mu.RLock()
	defer apm.mu.RUnlock()

	grid, exists := apm.visionGrids[instanceID]
	if !exists {
		apm.mu.RUnlock()
		apm.mu.Lock()
		defer apm.mu.Unlock()

		grid, exists = apm.visionGrids[instanceID]
		if !exists {
			grid = apm.visionGridPool.Get().(*VisionGrid)
			grid.GridSize = 5.0
			grid.GridCells = make(map[int64]map[int64]bool)
			grid.UpdateRate = 100 * time.Millisecond
			grid.LastUpdate = time.Now()
			apm.visionGrids[instanceID] = grid
		}
	} else {
		if time.Since(grid.LastUpdate) > grid.UpdateRate {
			apm.updateVisionGrid(grid)
		}
	}

	return grid
}

// updateVisionGrid 更新视野网格
func (apm *AIPerformanceManager) updateVisionGrid(grid *VisionGrid) {
	grid.LastUpdate = time.Now()
}

// IsVisible 检查是否可见
func (apm *AIPerformanceManager) IsVisible(instanceID uint64, start, end common.Vector3) bool {
	grid := apm.GetVisionGrid(instanceID)

	startX := int64(start.X / grid.GridSize)
	startY := int64(start.Y / grid.GridSize)
	endX := int64(end.X / grid.GridSize)
	endY := int64(end.Y / grid.GridSize)

	dx := abs(endX - startX)
	dy := abs(endY - startY)
	sx := 1
	if startX > endX {
		sx = -1
	}
	sy := 1
	if startY > endY {
		sy = -1
	}
	err := dx - dy

	x, y := startX, startY
	for {
		if row, exists := grid.GridCells[x]; exists {
			if blocked, exists := row[y]; exists && blocked {
				return false
			}
		}

		if x == endX && y == endY {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += int64(sx)
		}
		if e2 < dx {
			err += dx
			y += int64(sy)
		}
	}

	return true
}

// abs 返回绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// FindPath 寻找路径
func (apm *AIPerformanceManager) FindPath(start, end common.Vector3, maxDistance float32) []common.Vector3 {
	openSet := make(map[common.Vector3]*PathNode)
	closedSet := make(map[common.Vector3]bool)

	startNode := apm.pathNodePool.Get().(*PathNode)
	startNode.Position = start
	startNode.G = 0
	startNode.H = start.DistanceTo(end)
	startNode.F = startNode.G + startNode.H
	openSet[start] = startNode

	var foundPath bool
	var current *PathNode

	for len(openSet) > 0 {
		current = nil
		for _, node := range openSet {
			if current == nil || node.F < current.F {
				current = node
			}
		}

		if current.Position.DistanceTo(end) < 1.0 {
			foundPath = true
			break
		}

		delete(openSet, current.Position)
		closedSet[current.Position] = true

		neighbors := apm.generateNeighbors(current.Position)
		for _, neighborPos := range neighbors {
			if _, exists := closedSet[neighborPos]; exists {
				continue
			}

			if neighborPos.DistanceTo(start) > maxDistance {
				continue
			}

			g := current.G + current.Position.DistanceTo(neighborPos)

			neighborNode, exists := openSet[neighborPos]
			if !exists {
				neighborNode = apm.pathNodePool.Get().(*PathNode)
				neighborNode.Position = neighborPos
				neighborNode.H = neighborPos.DistanceTo(end)
				openSet[neighborPos] = neighborNode
			} else if g >= neighborNode.G {
				continue
			}

			neighborNode.Parent = current
			neighborNode.G = g
			neighborNode.F = g + neighborNode.H
		}
	}

	path := make([]common.Vector3, 0)
	if foundPath && current != nil {
		for current != nil {
			path = append([]common.Vector3{current.Position}, path...)
			parent := current.Parent
			apm.pathNodePool.Put(current)
			current = parent
		}
	}

	for _, node := range openSet {
		apm.pathNodePool.Put(node)
	}

	return path
}

// generateNeighbors 生成邻居节点
func (apm *AIPerformanceManager) generateNeighbors(position common.Vector3) []common.Vector3 {
	neighbors := make([]common.Vector3, 0, 8)

	directions := []common.Vector3{
		{X: 1, Y: 0, Z: 0},
		{X: -1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: -1, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: -1, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: 0},
	}

	for _, dir := range directions {
		neighborPos := position.Add(dir.MultiplyScalar(1.0))
		neighbors = append(neighbors, neighborPos)
	}

	return neighbors
}

// CreateGroupAI 创建群体AI
func (apm *AIPerformanceManager) CreateGroupAI(members []uint64) uint64 {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	groupID := uint64(time.Now().UnixNano())
	groupAI := apm.groupAIPool.Get().(*GroupAI)
	groupAI.GroupID = groupID
	groupAI.Members = make(map[uint64]bool)
	groupAI.LastAction = time.Now()
	groupAI.Behavior = "normal"

	for _, memberID := range members {
		groupAI.Members[memberID] = true
	}

	if len(members) > 0 {
		groupAI.LeaderID = members[0]
	}

	apm.groupAIs[groupID] = groupAI
	return groupID
}

// AddToGroup 添加到群体
func (apm *AIPerformanceManager) AddToGroup(groupID, memberID uint64) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	groupAI, exists := apm.groupAIs[groupID]
	if !exists {
		return false
	}

	groupAI.Members[memberID] = true
	groupAI.LastAction = time.Now()
	return true
}

// RemoveFromGroup 从群体移除
func (apm *AIPerformanceManager) RemoveFromGroup(groupID, memberID uint64) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	groupAI, exists := apm.groupAIs[groupID]
	if !exists {
		return false
	}

	delete(groupAI.Members, memberID)
	groupAI.LastAction = time.Now()

	if groupAI.LeaderID == memberID && len(groupAI.Members) > 0 {
		for memberID := range groupAI.Members {
			groupAI.LeaderID = memberID
			break
		}
	}

	if len(groupAI.Members) == 0 {
		delete(apm.groupAIs, groupID)
		apm.groupAIPool.Put(groupAI)
	}

	return true
}

// SetGroupTarget 设置群体目标
func (apm *AIPerformanceManager) SetGroupTarget(groupID, targetID uint64) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	groupAI, exists := apm.groupAIs[groupID]
	if !exists {
		return false
	}

	groupAI.TargetID = targetID
	groupAI.LastAction = time.Now()

	return true
}

// UpdateGroupAI 更新群体AI
func (apm *AIPerformanceManager) UpdateGroupAI(groupID uint64) {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	groupAI, exists := apm.groupAIs[groupID]
	if !exists {
		return
	}

	currentTime := time.Now()
	if currentTime.Sub(groupAI.LastAction) < 100*time.Millisecond {
		return
	}

	switch groupAI.Behavior {
	case "attack":
		apm.updateAttackBehavior(groupAI)
	case "defend":
		apm.updateDefendBehavior(groupAI)
	case "patrol":
		apm.updatePatrolBehavior(groupAI)
	}

	groupAI.LastAction = currentTime
}

// updateAttackBehavior 更新攻击行为
func (apm *AIPerformanceManager) updateAttackBehavior(groupAI *GroupAI) {
	if groupAI.TargetID == 0 {
		return
	}

	target := apm.getGameObject(groupAI.TargetID)
	if target == nil {
		return
	}

	leader := apm.getGameObject(groupAI.LeaderID)
	if leader == nil {
		return
	}

	memberIndex := 0
	for memberID := range groupAI.Members {
		if memberID == groupAI.LeaderID {
			continue
		}

		memberIndex++
	}
}

// updateDefendBehavior 更新防御行为
func (apm *AIPerformanceManager) updateDefendBehavior(groupAI *GroupAI) {
}

// updatePatrolBehavior 更新巡逻行为
func (apm *AIPerformanceManager) updatePatrolBehavior(groupAI *GroupAI) {
}

// GetGroupAI 获取群体AI
func (apm *AIPerformanceManager) GetGroupAI(groupID uint64) (*GroupAI, bool) {
	apm.mu.RLock()
	defer apm.mu.RUnlock()

	groupAI, exists := apm.groupAIs[groupID]
	return groupAI, exists
}

// CleanupExpiredGroups 清理过期的群体
func (apm *AIPerformanceManager) CleanupExpiredGroups() {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	currentTime := time.Now()
	expiredGroups := make([]uint64, 0)

	for groupID, groupAI := range apm.groupAIs {
		if currentTime.Sub(groupAI.LastAction) > 5*time.Minute {
			expiredGroups = append(expiredGroups, groupID)
		}
	}

	for _, groupID := range expiredGroups {
		if groupAI, exists := apm.groupAIs[groupID]; exists {
			delete(apm.groupAIs, groupID)
			apm.groupAIPool.Put(groupAI)
		}
	}
}

// getGameObject 获取游戏对象
func (apm *AIPerformanceManager) getGameObject(objectID uint64) common.IGameObject {
	return nil
}
