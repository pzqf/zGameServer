package models

// Map 地图模型
type Map struct {
	MapID       int32   `json:"map_id"`
	Name        string  `json:"name"`
	MapType     int32   `json:"map_type"`
	Width       int32   `json:"width"`
	Height      int32   `json:"height"`
	RegionSize  int32   `json:"region_size"`
	TileWidth   int32   `json:"tile_width"`
	TileHeight  int32   `json:"tile_height"`
	IsInstance  bool    `json:"is_instance"`
	MaxPlayers  int32   `json:"max_players"`
	Description string  `json:"description"`
	Background  string  `json:"background"`
	Music       string  `json:"music"`
	WeatherType string  `json:"weather_type"`
	MinLevel    int32   `json:"min_level"`
	MaxLevel    int32   `json:"max_level"`
	RespawnRate float64 `json:"respawn_rate"`
}

// MapSpawnPoint 地图生成点模型
type MapSpawnPoint struct {
	ID        int32   `json:"id"`
	MapID     int32   `json:"map_id"`
	Type      string  `json:"type"`
	ObjectID  int32   `json:"object_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Name      string  `json:"name"`
	Frequency int32   `json:"frequency"`
	GroupID   int32   `json:"group_id"`
}

// MapTeleportPoint 地图传送点模型
type MapTeleportPoint struct {
	ID            int32   `json:"id"`
	MapID         int32   `json:"map_id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             float64 `json:"z"`
	TargetMapID   int32   `json:"target_map_id"`
	TargetX       float64 `json:"target_x"`
	TargetY       float64 `json:"target_y"`
	TargetZ       float64 `json:"target_z"`
	Name          string  `json:"name"`
	RequiredLevel int32   `json:"required_level"`
	RequiredItem  int32   `json:"required_item"`
	IsActive      bool    `json:"is_active"`
}

// MapBuilding 地图建筑物模型
type MapBuilding struct {
	ID      int32   `json:"id"`
	MapID   int32   `json:"map_id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Z       float64 `json:"z"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Type    string  `json:"type"`
	Name    string  `json:"name"`
	Level   int32   `json:"level"`
	HP      int32   `json:"hp"`
	Faction int32   `json:"faction"`
}

// MapEvent 地图事件模型
type MapEvent struct {
	EventID     int32   `json:"event_id"`
	MapID       int32   `json:"map_id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	Radius      float64 `json:"radius"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Duration    int32   `json:"duration"`
	RewardID    int32   `json:"reward_id"`
	IsActive    bool    `json:"is_active"`
}

// MapResource 地图资源模型
type MapResource struct {
	ResourceID  int32   `json:"resource_id"`
	MapID       int32   `json:"map_id"`
	Type        string  `json:"type"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	RespawnTime int32   `json:"respawn_time"`
	ItemID      int32   `json:"item_id"`
	Quantity    int32   `json:"quantity"`
	Level       int32   `json:"level"`
	IsGathering bool    `json:"is_gathering"`
}
