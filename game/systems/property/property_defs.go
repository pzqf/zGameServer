package property

// PropertyType 属性类型枚举
type PropertyType string

// 属性类型常量定义
const (
	// 生命值相关
	PropertyHP    PropertyType = "health"     // 当前血量值
	PropertyMaxHP PropertyType = "max_health" // 最大血量值

	// 魔法值相关
	PropertyMP    PropertyType = "mana"     // 当前魔法值
	PropertyMaxMP PropertyType = "max_mana" // 最大魔法值

	// 战斗属性
	PropertyCriticalRate   PropertyType = "critical_rate"   // 暴击率
	PropertyCriticalDamage PropertyType = "critical_damage" // 暴击伤害
	PropertyHaste          PropertyType = "speed"           // 急速（对应现有speed属性）
	PropertyHit            PropertyType = "hit"             // 命中
	PropertyDodge          PropertyType = "dodge"           // 闪避

	// 攻击属性
	PropertyPhysicalAttack PropertyType = "attack"       // 物理攻击（对应现有attack属性）
	PropertyMagicAttack    PropertyType = "magic_attack" // 魔法攻击
	PropertyAttackPower    PropertyType = "attack_power" // 攻击力
	PropertySkillPower     PropertyType = "skill_power"  // 技能强度

	// 防御属性
	PropertyPhysicalDefense PropertyType = "defense"       // 物理防御（对应现有defense属性）
	PropertyMagicDefense    PropertyType = "magic_defense" // 魔法防御

	// 属性相关
	PropertyIntellect      PropertyType = "intellect"       // 智力
	PropertyManaEfficiency PropertyType = "mana_efficiency" // 魔法效率

	// 其他属性
	PropertyCooldownReduction PropertyType = "cooldown_reduction" // 冷却缩减
	PropertyRangeBonus        PropertyType = "range_bonus"        // 范围加成
	PropertyExp               PropertyType = "exp"                // 经验值
	PropertyMoveSpeed         PropertyType = "move_speed"         // 移动速度
	PropertyAttackRange       PropertyType = "attack_range"       // 攻击范围
)

// GetPropertyType 获取属性类型字符串
func GetPropertyType(pt PropertyType) string {
	return string(pt)
}

// GetAllPropertyTypes 获取所有属性类型
func GetAllPropertyTypes() []PropertyType {
	return []PropertyType{
		PropertyHP,
		PropertyMaxHP,
		PropertyMP,
		PropertyMaxMP,
		PropertyCriticalRate,
		PropertyCriticalDamage,
		PropertyHaste,
		PropertyHit,
		PropertyDodge,
		PropertyPhysicalAttack,
		PropertyMagicAttack,
		PropertyAttackPower,
		PropertySkillPower,
		PropertyPhysicalDefense,
		PropertyMagicDefense,
		PropertyIntellect,
		PropertyManaEfficiency,
		PropertyCooldownReduction,
		PropertyRangeBonus,
		PropertyExp,
		PropertyMoveSpeed,
		PropertyAttackRange,
	}
}
