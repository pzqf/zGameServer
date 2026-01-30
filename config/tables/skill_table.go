package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// SkillTableLoader 技能表格加载器
type SkillTableLoader struct {
	mu     sync.RWMutex
	skills map[int32]*models.Skill
}

// NewSkillTableLoader 创建技能表格加载器
func NewSkillTableLoader() *SkillTableLoader {
	return &SkillTableLoader{
		skills: make(map[int32]*models.Skill),
	}
}

// Load 加载技能表数据
func (stl *SkillTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "skill.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 25,
		TableName:  "skills",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempSkills := make(map[int32]*models.Skill)

	err := ReadExcelFile(config, dir, func(row []string) error {
		// 确保行数据足够长
		for len(row) < 25 {
			row = append(row, "")
		}

		skill := &models.Skill{
			SkillID:              StrToInt32(row[0]),
			Name:                 row[1],
			Type:                 StrToInt32(row[2]),
			Level:                StrToInt32(row[3]),
			ManaCost:             StrToInt32(row[4]),
			Cooldown:             StrToFloat32(row[5]),
			Damage:               StrToInt32(row[6]),
			Range:                StrToFloat32(row[7]),
			AreaRadius:           StrToFloat32(row[8]),
			Description:          row[9],
			Effects:              row[10],
			DamageType:           row[11],
			EffectType:           row[12],
			CooldownGrowth:       StrToFloat32(row[13]),
			DamageGrowth:         StrToFloat32(row[14]),
			RangeGrowth:          StrToFloat32(row[15]),
			RequiredLevel:        StrToInt32(row[16]),
			AnimationID:          StrToInt32(row[17]),
			SoundID:              StrToInt32(row[18]),
			IconID:               StrToInt32(row[19]),
			PreSkillID:           StrToInt32(row[20]),
			BuffID:               StrToInt32(row[21]),
			SkillCastTime:        StrToFloat32(row[22]),
			SkillProjectileSpeed: StrToFloat32(row[23]),
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
