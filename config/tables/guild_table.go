package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// GuildTableLoader 公会表加载器
type GuildTableLoader struct {
	mu     sync.RWMutex
	guilds map[int32]*models.Guild
}

// NewGuildTableLoader 创建公会表加载器
func NewGuildTableLoader() *GuildTableLoader {
	return &GuildTableLoader{
		guilds: make(map[int32]*models.Guild),
	}
}

// Load 加载公会表数据
func (gtl *GuildTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "guild.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 7,
		TableName:  "guilds",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempGuilds := make(map[int32]*models.Guild)

	err := ReadExcelFile(config, dir, func(row []string) error {
		guild := &models.Guild{
			GuildLevel:    StrToInt32(row[0]),
			RequiredExp:   StrToInt64(row[1]),
			MaxMembers:    StrToInt32(row[2]),
			BuildingSlots: StrToInt32(row[3]),
			TaxRate:       StrToFloat32(row[4]),
			SkillPoints:   StrToInt32(row[5]),
		}

		tempGuilds[guild.GuildLevel] = guild
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		gtl.mu.Lock()
		gtl.guilds = tempGuilds
		gtl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (gtl *GuildTableLoader) GetTableName() string {
	return "guilds"
}

// GetGuild 根据ID获取公会配置
func (gtl *GuildTableLoader) GetGuild(guildLevel int32) (*models.Guild, bool) {
	gtl.mu.RLock()
	guild, ok := gtl.guilds[guildLevel]
	gtl.mu.RUnlock()
	return guild, ok
}

// GetAllGuilds 获取所有公会配置
func (gtl *GuildTableLoader) GetAllGuilds() map[int32]*models.Guild {
	gtl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	guildsCopy := make(map[int32]*models.Guild, len(gtl.guilds))
	for id, guild := range gtl.guilds {
		guildsCopy[id] = guild
	}
	gtl.mu.RUnlock()
	return guildsCopy
}
