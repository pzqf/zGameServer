package tables

import (
	"github.com/pzqf/zGameServer/config/models"
)

// Load 加载技能表数据
func (stl *SkillTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "skill.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 11,
		TableName:  "skills",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempSkills := make(map[int32]*models.Skill)

	err := ReadExcelFile(config, dir, func(row []string) error {
		skill := &models.Skill{
			SkillID:     StrToInt32(row[0]),
			Name:        row[1],
			Type:        StrToInt32(row[2]),
			Level:       StrToInt32(row[3]),
			ManaCost:    StrToInt32(row[4]),
			Cooldown:    StrToFloat32(row[5]),
			Damage:      StrToInt32(row[6]),
			Range:       StrToFloat32(row[7]),
			AreaRadius:  StrToFloat32(row[8]),
			Description: row[9],
			Effects:     row[10],
		}

		tempSkills[skill.SkillID] = skill
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		stl.mu.Lock()
		stl.skills = tempSkills
		stl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (stl *SkillTableLoader) GetTableName() string {
	return "skills"
}

// GetSkill 根据ID获取技能
func (stl *SkillTableLoader) GetSkill(skillID int32) (*models.Skill, bool) {
	stl.mu.RLock()
	skill, ok := stl.skills[skillID]
	stl.mu.RUnlock()
	return skill, ok
}

// GetAllSkills 获取所有技能
func (stl *SkillTableLoader) GetAllSkills() map[int32]*models.Skill {
	stl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	skillsCopy := make(map[int32]*models.Skill, len(stl.skills))
	for id, skill := range stl.skills {
		skillsCopy[id] = skill
	}
	stl.mu.RUnlock()
	return skillsCopy
}
