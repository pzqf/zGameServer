package models

// Map 地图配置结构
type Map struct {
    MapID     int32   `json:"map_id"`
    Name      string  `json:"name"`
    Width     int32   `json:"width"`
    Height    int32   `json:"height"`
    MaxPlayer int32   `json:"max_player"`
    MonsterConfig string `json:"monster_config"` // JSON格式的怪物配置
    TerrainData string `json:"terrain_data"` // JSON格式的地形数据
    RespawnPointX float32 `json:"respawn_point_x"`
    RespawnPointY float32 `json:"respawn_point_y"`
}
