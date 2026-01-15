package player

import (
	"go.uber.org/zap"

	"github.com/pzqf/zUtil/zMap"
)

// 任务状态定义
const (
	TaskStatusNotAccepted = 1
	TaskStatusInProgress  = 2
	TaskStatusCompleted   = 3
	TaskStatusRewarded    = 4
)

// 任务类型定义
const (
	TaskTypeMain   = 1 // 主线任务
	TaskTypeSide   = 2 // 支线任务
	TaskTypeDaily  = 3 // 日常任务
	TaskTypeWeekly = 4 // 周常任务
	TaskTypeEvent  = 5 // 活动任务
)

// 任务条件类型定义
const (
	TaskCondTypeKillMonster  = 1 // 击杀怪物
	TaskCondTypeCollectItem  = 2 // 收集物品
	TaskCondTypeTalkToNPC    = 3 // 与NPC对话
	TaskCondTypeReachLevel   = 4 // 达到等级
	TaskCondTypeCompleteTask = 5 // 完成前置任务
)

// 任务奖励类型定义
const (
	TaskRewardTypeCoin       = 1 // 金币
	TaskRewardTypeExp        = 2 // 经验
	TaskRewardTypeItem       = 3 // 物品
	TaskRewardTypeSkillPoint = 4 // 技能点
	TaskRewardTypeReputation = 5 // 声望
)

// TaskCondition 任务条件
type TaskCondition struct {
	condType  int // 条件类型
	condValue int // 条件值
	progress  int // 已完成进度
	target    int // 目标值
}

// TaskReward 任务奖励
type TaskReward struct {
	rewardType int   // 奖励类型
	value      int64 // 奖励值
	itemId     int64 // 物品ID（如果奖励是物品）
	count      int   // 物品数量
}

// Task 任务结构
type Task struct {
	taskId       int64
	taskType     int
	title        string
	description  string
	conditions   []*TaskCondition
	rewards      []*TaskReward
	status       int
	acceptTime   int64
	completeTime int64
}

// TaskManager 任务管理系统
type TaskManager struct {
	playerId int64
	logger   *zap.Logger
	tasks    *zMap.Map // key: int64(taskId), value: *Task
	maxCount int
}

func NewTaskManager(playerId int64, logger *zap.Logger) *TaskManager {
	return &TaskManager{
		playerId: playerId,
		logger:   logger,
		tasks:    zMap.NewMap(),
		maxCount: 20, // 最大同时进行的任务数量
	}
}

func (tm *TaskManager) Init() {
	// 初始化任务管理系统
	tm.logger.Debug("Initializing task manager", zap.Int64("playerId", tm.playerId))
}

// AcceptTask 接受任务
func (tm *TaskManager) AcceptTask(task *Task) error {
	// 检查任务是否已接受
	if _, exists := tm.tasks.Get(task.taskId); exists {
		return nil // 任务已接受
	}

	// 检查是否达到最大任务数量
	if tm.tasks.Len() >= int64(tm.maxCount) {
		return nil // 已达到最大任务数量
	}

	// 接受任务
	task.status = TaskStatusInProgress
	tm.tasks.Store(task.taskId, task)
	tm.logger.Info("Task accepted", zap.Int64("taskId", task.taskId), zap.Int64("playerId", tm.playerId))
	return nil
}

// UpdateTaskProgress 更新任务进度
func (tm *TaskManager) UpdateTaskProgress(taskId int64, condType int, progress int) error {
	// 获取任务
	taskInterface, exists := tm.tasks.Get(taskId)
	if !exists {
		return nil // 任务不存在
	}

	task := taskInterface.(*Task)
	if task.status != TaskStatusInProgress {
		return nil // 任务不在进行中
	}

	// 更新任务条件进度
	for _, cond := range task.conditions {
		if cond.condType == condType {
			// 确保进度不超过目标值
			if progress > cond.target {
				progress = cond.target
			}
			cond.progress = progress
			break
		}
	}

	// 检查所有任务条件是否都已完成
	allCompleted := true
	for _, cond := range task.conditions {
		if cond.progress < cond.target {
			allCompleted = false
			break
		}
	}

	// 如果所有条件都完成，将任务状态改为已完成
	if allCompleted {
		task.status = TaskStatusCompleted
		tm.logger.Info("Task completed", zap.Int64("taskId", taskId), zap.Int64("playerId", tm.playerId))
	}

	// 更新任务
	tm.tasks.Store(taskId, task)
	return nil
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskId int64) ([]*TaskReward, error) {
	// 获取任务
	taskInterface, exists := tm.tasks.Get(taskId)
	if !exists {
		return nil, nil // 任务不存在
	}

	task := taskInterface.(*Task)
	if task.status != TaskStatusCompleted {
		return nil, nil // 任务未完成
	}

	// 标记任务为已领取奖励
	task.status = TaskStatusRewarded
	tm.tasks.Store(taskId, task)

	// 返回任务奖励
	tm.logger.Info("Task rewarded", zap.Int64("taskId", taskId), zap.Int64("playerId", tm.playerId))
	return task.rewards, nil
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(taskId int64) (*Task, bool) {
	task, exists := tm.tasks.Get(taskId)
	if !exists {
		return nil, false
	}
	return task.(*Task), true
}

// GetAllTasks 获取所有任务
func (tm *TaskManager) GetAllTasks() []*Task {
	var tasks []*Task
	tm.tasks.Range(func(key, value interface{}) bool {
		if value != nil {
			tasks = append(tasks, value.(*Task))
		}
		return true
	})
	return tasks
}

// GetInProgressTasks 获取进行中的任务
func (tm *TaskManager) GetInProgressTasks() []*Task {
	var tasks []*Task
	tm.tasks.Range(func(key, value interface{}) bool {
		if value != nil {
			task := value.(*Task)
			if task.status == TaskStatusInProgress {
				tasks = append(tasks, task)
			}
		}
		return true
	})
	return tasks
}

// GetCompletedTasks 获取已完成但未领取奖励的任务
func (tm *TaskManager) GetCompletedTasks() []*Task {
	var tasks []*Task
	tm.tasks.Range(func(key, value interface{}) bool {
		if value != nil {
			task := value.(*Task)
			if task.status == TaskStatusCompleted {
				tasks = append(tasks, task)
			}
		}
		return true
	})
	return tasks
}

// AbandonTask 放弃任务
func (tm *TaskManager) AbandonTask(taskId int64) bool {
	// 获取任务
	taskInterface, exists := tm.tasks.Get(taskId)
	if !exists {
		return false // 任务不存在
	}

	task := taskInterface.(*Task)
	// 只能放弃进行中的任务
	if task.status != TaskStatusInProgress {
		return false
	}

	// 删除任务
	tm.tasks.Delete(taskId)
	tm.logger.Info("Task abandoned", zap.Int64("taskId", taskId), zap.Int64("playerId", tm.playerId))
	return true
}
