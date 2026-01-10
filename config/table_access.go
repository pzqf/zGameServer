package config

import (
	"github.com/pzqf/zGameServer/config/models"
)

// GetItemByID 根据ID获取物品配置
func GetItemByID(itemID int32) *models.Item {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	return GlobalTableLoader.items[itemID]
}

// GetAllItems 获取所有物品配置
func GetAllItems() map[int32]*models.Item {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	// 创建副本以避免并发修改问题
	result := make(map[int32]*models.Item)
	for k, v := range GlobalTableLoader.items {
		result[k] = v
	}

	return result
}

// GetMapByID 根据ID获取地图配置
func GetMapByID(mapID int32) *models.Map {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	return GlobalTableLoader.maps[mapID]
}

// GetAllMaps 获取所有地图配置
func GetAllMaps() map[int32]*models.Map {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	// 创建副本以避免并发修改问题
	result := make(map[int32]*models.Map)
	for k, v := range GlobalTableLoader.maps {
		result[k] = v
	}

	return result
}

// GetSkillByID 根据ID获取技能配置
func GetSkillByID(skillID int32) *models.Skill {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	return GlobalTableLoader.skills[skillID]
}

// GetAllSkills 获取所有技能配置
func GetAllSkills() map[int32]*models.Skill {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	// 创建副本以避免并发修改问题
	result := make(map[int32]*models.Skill)
	for k, v := range GlobalTableLoader.skills {
		result[k] = v
	}

	return result
}

// GetQuestByID 根据ID获取任务配置
func GetQuestByID(questID int32) *models.Quest {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	return GlobalTableLoader.quests[questID]
}

// GetAllQuests 获取所有任务配置
func GetAllQuests() map[int32]*models.Quest {
	if GlobalTableLoader == nil {
		return nil
	}

	GlobalTableLoader.mu.RLock()
	defer GlobalTableLoader.mu.RUnlock()

	// 创建副本以避免并发修改问题
	result := make(map[int32]*models.Quest)
	for k, v := range GlobalTableLoader.quests {
		result[k] = v
	}

	return result
}
