package models

// Skill 技能配置结构
type Skill struct {
    SkillID     int32   `json:"skill_id"`
    Name        string  `json:"name"`
    Type        int32   `json:"type"`
    Level       int32   `json:"level"`
    ManaCost    int32   `json:"mana_cost"`
    Cooldown    float32 `json:"cooldown"`
    Damage      int32   `json:"damage"`
    Range       float32 `json:"range"`
    AreaRadius  float32 `json:"area_radius"`
    Description string  `json:"description"`
    Effects     string  `json:"effects"` // JSON格式的效果描述
}
