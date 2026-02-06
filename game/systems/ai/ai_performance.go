package ai

import (
	"sync"
	"time"

	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/common"
	gamecommon "github.com/pzqf/zGameServer/game/common"
)

type VisionGrid struct {
	GridSize   float32
	GridCells  map[int64]map[int64]bool
	LastUpdate time.Time
	UpdateRate time.Duration
}

type PathNode struct {
	Position gamecommon.Vector3
	Parent   *PathNode
	G        float32
	H        float32
	F        float32
}

type TeamAI struct {
	TeamID     common.TeamIdType
	Members    map[common.ObjectIdType]bool
	LeaderID   common.PlayerIdType
	TargetID   common.ObjectIdType
	Formation  string
	LastAction time.Time
	Behavior   string
}

type AIPerformanceManager struct {
	mu             sync.RWMutex
	visionGrids    map[common.MapIdType]*VisionGrid
	teamAIs        map[common.TeamIdType]*TeamAI
	pathNodePool   *zObject.GenericPool
	visionGridPool *zObject.GenericPool
	teamAIPool     *zObject.GenericPool
}

func NewAIPerformanceManager() *AIPerformanceManager {
	return &AIPerformanceManager{
		visionGrids:    make(map[common.MapIdType]*VisionGrid),
		teamAIs:        make(map[common.TeamIdType]*TeamAI),
		pathNodePool:   zObject.NewGenericPool(func() interface{} { return &PathNode{} }, 1000),
		visionGridPool: zObject.NewGenericPool(func() interface{} { return &VisionGrid{} }, 100),
		teamAIPool:     zObject.NewGenericPool(func() interface{} { return &TeamAI{} }, 100),
	}
}

func (apm *AIPerformanceManager) Init() error {
	return nil
}

func (apm *AIPerformanceManager) GetVisionGrid(instanceID common.MapIdType) *VisionGrid {
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

func (apm *AIPerformanceManager) updateVisionGrid(grid *VisionGrid) {
	grid.LastUpdate = time.Now()
}

func (apm *AIPerformanceManager) IsVisible(instanceID common.MapIdType, start, end gamecommon.Vector3) bool {
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

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (apm *AIPerformanceManager) FindPath(start, end gamecommon.Vector3, maxDistance float32) []gamecommon.Vector3 {
	openSet := make(map[gamecommon.Vector3]*PathNode)
	closedSet := make(map[gamecommon.Vector3]bool)

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

	path := make([]gamecommon.Vector3, 0)
	if foundPath && current != nil {
		for current != nil {
			path = append([]gamecommon.Vector3{current.Position}, path...)
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

func (apm *AIPerformanceManager) generateNeighbors(position gamecommon.Vector3) []gamecommon.Vector3 {
	neighbors := make([]gamecommon.Vector3, 0, 8)

	directions := []gamecommon.Vector3{
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

func (apm *AIPerformanceManager) CreateTeamAI(members []common.ObjectIdType) common.TeamIdType {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	teamID, err := common.GenerateTeamID()
	if err != nil {
		return 0
	}
	teamAI := apm.teamAIPool.Get().(*TeamAI)
	teamAI.TeamID = teamID
	teamAI.Members = make(map[common.ObjectIdType]bool)
	teamAI.LastAction = time.Now()
	teamAI.Behavior = "normal"

	for _, memberID := range members {
		teamAI.Members[memberID] = true
	}

	if len(members) > 0 {
		teamAI.LeaderID = common.PlayerIdType(members[0])
	}

	apm.teamAIs[teamID] = teamAI
	return teamID
}

func (apm *AIPerformanceManager) AddToTeam(teamID common.TeamIdType, memberID common.ObjectIdType) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	teamAI, exists := apm.teamAIs[teamID]
	if !exists {
		return false
	}

	teamAI.Members[memberID] = true
	teamAI.LastAction = time.Now()
	return true
}

func (apm *AIPerformanceManager) RemoveFromTeam(teamID common.TeamIdType, memberID common.ObjectIdType) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	teamAI, exists := apm.teamAIs[teamID]
	if !exists {
		return false
	}

	delete(teamAI.Members, memberID)
	teamAI.LastAction = time.Now()

	if common.ObjectIdType(teamAI.LeaderID) == memberID && len(teamAI.Members) > 0 {
		for memberID := range teamAI.Members {
			teamAI.LeaderID = common.PlayerIdType(memberID)
			break
		}
	}

	if len(teamAI.Members) == 0 {
		delete(apm.teamAIs, teamID)
		apm.teamAIPool.Put(teamAI)
	}

	return true
}

func (apm *AIPerformanceManager) SetTeamTarget(teamID common.TeamIdType, targetID common.ObjectIdType) bool {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	teamAI, exists := apm.teamAIs[teamID]
	if !exists {
		return false
	}

	teamAI.TargetID = targetID
	teamAI.LastAction = time.Now()

	return true
}

func (apm *AIPerformanceManager) UpdateTeamAI(teamID common.TeamIdType) {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	teamAI, exists := apm.teamAIs[teamID]
	if !exists {
		return
	}

	currentTime := time.Now()
	if currentTime.Sub(teamAI.LastAction) < 100*time.Millisecond {
		return
	}

	switch teamAI.Behavior {
	case "attack":
		apm.updateAttackBehavior(teamAI)
	case "defend":
		apm.updateDefendBehavior(teamAI)
	case "patrol":
		apm.updatePatrolBehavior(teamAI)
	}

	teamAI.LastAction = currentTime
}

func (apm *AIPerformanceManager) updateAttackBehavior(teamAI *TeamAI) {
	if teamAI.TargetID == 0 {
		return
	}

	target := apm.getGameObject(teamAI.TargetID)
	if target == nil {
		return
	}

	leader := apm.getGameObject(common.ObjectIdType(teamAI.LeaderID))
	if leader == nil {
		return
	}

	memberIndex := 0
	for memberID := range teamAI.Members {
		if memberID == common.ObjectIdType(teamAI.LeaderID) {
			continue
		}

		memberIndex++
	}
}

func (apm *AIPerformanceManager) updateDefendBehavior(teamAI *TeamAI) {
}

func (apm *AIPerformanceManager) updatePatrolBehavior(teamAI *TeamAI) {
}

func (apm *AIPerformanceManager) GetTeamAI(teamID common.TeamIdType) (*TeamAI, bool) {
	apm.mu.RLock()
	defer apm.mu.RUnlock()

	teamAI, exists := apm.teamAIs[teamID]
	return teamAI, exists
}

func (apm *AIPerformanceManager) CleanupExpiredTeams() {
	apm.mu.Lock()
	defer apm.mu.Unlock()

	currentTime := time.Now()
	expiredTeams := make([]common.TeamIdType, 0)

	for teamID, teamAI := range apm.teamAIs {
		if currentTime.Sub(teamAI.LastAction) > 5*time.Minute {
			expiredTeams = append(expiredTeams, teamID)
		}
	}

	for _, teamID := range expiredTeams {
		if teamAI, exists := apm.teamAIs[teamID]; exists {
			delete(apm.teamAIs, teamID)
			apm.teamAIPool.Put(teamAI)
		}
	}
}

func (apm *AIPerformanceManager) getGameObject(objectID common.ObjectIdType) gamecommon.IGameObject {
	return nil
}
