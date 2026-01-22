# zGameServer

一个基于Go语言开发的MMO游戏服务器框架，采用模块化设计，具有良好的可扩展性和性能。

## 项目特点

- **模块化设计**：清晰的代码结构，便于维护和扩展
- **多数据库支持**：支持MySQL和MongoDB数据库
- **异步数据库操作**：所有数据库操作均为异步，提高服务器性能
- **配置化管理**：通过配置文件和Excel表格管理服务器参数和游戏数据
- **完善的日志系统**：使用zap日志框架，支持不同级别日志输出
- **协议缓冲区**：使用Protocol Buffers进行高效的网络通信
- **安全机制**：账号密码验证、角色数据保护
- **崩溃恢复**：服务器崩溃时自动捕获堆栈信息并记录到日志
- **组件系统**：基于组件的游戏对象管理，提供高度的灵活性和可扩展性

## 技术栈

- **语言**：Go 1.25.5
- **网络**：TCP和HTTP协议，Protocol Buffers
- **数据库**：MySQL、MongoDB
- **日志**：zap日志框架
- **配置**：ini配置文件，Excel表格
- **依赖管理**：Go Modules

## 项目结构

```
zGameServer/
├── client/                 # 客户端测试代码
├── config/                 # 配置管理
│   ├── models/             # 配置模型
│   ├── tables/             # Excel配置表管理
│   └── ini_config.go       # ini配置文件管理
├── db/                     # 数据库相关代码
│   ├── connector/          # 数据库连接器
│   ├── dao/                # 数据访问对象
│   ├── models/             # 数据模型
│   └── db.go               # 数据库管理器
├── event/                  # 事件系统
├── game/                   # 游戏核心逻辑
│   ├── auction/            # 拍卖行系统
│   ├── common/             # 公共接口和工具
│   ├── guild/              # 公会系统
│   ├── maps/               # 地图系统
│   ├── monsters/           # 怪物系统
│   ├── npc/                # NPC系统
│   ├── object/             # 游戏对象系统
│   │   └── component/      # 组件系统
│   ├── pets/               # 宠物系统
│   ├── player/             # 玩家系统
│   └── systems/            # 核心游戏系统
│       ├── combat/         # 战斗系统
│       ├── movement/       # 移动系统
│       ├── property/       # 属性系统
│       └── skill/          # 技能系统
├── gameserver/             # 服务器核心
├── net/                    # 网络相关代码
│   ├── handler/            # 请求处理器
│   ├── protocol/           # 协议定义
│   ├── protolayer/         # 协议层实现
│   ├── router/             # 路由管理
│   └── service/            # 网络服务
├── resources/              # 资源文件
│   ├── excel_tables/       # Excel配置表
│   └── maps/               # 地图资源
├── config.ini              # 配置文件
├── go.mod                  # Go模块依赖
├── go.sum                  # 依赖校验
└── main.go                 # 主入口文件
```

## 核心功能模块

### 1. 网络服务

- TCP服务，处理客户端实时连接
- HTTP服务，处理非实时请求（如充值下发等）
- 数据包路由和处理
- 会话管理
- 支持多种协议格式（Protobuf、JSON、XML）

### 2. 游戏对象系统

- **GameObject**：基础游戏对象，包含ID、名称、位置等基本属性
- **LivingObject**：生命对象，继承自GameObject，增加生命值等属性
- **Component**：组件系统，允许动态为游戏对象添加功能
- **ComponentManager**：组件管理器，负责组件的添加、移除和管理

### 3. 玩家系统

- 账号创建和登录
- 角色创建、删除和选择
- **PlayerInventory**：玩家背包系统，管理物品
- **PlayerEquipment**：玩家装备系统，管理装备
- **PlayerSkill**：玩家技能系统，管理技能
- **PlayerTask**：玩家任务系统，管理任务
- **PlayerMailbox**：玩家邮箱系统，管理邮件

### 4. 怪物和NPC系统

- **Monster**：怪物对象，包含AI行为、掉落配置
- **NPC**：非玩家角色，包含对话、交互逻辑
- **AIBehavior**：AI行为系统，处理怪物和NPC的智能行为

### 5. 宠物系统

- **Pet**：宠物对象，包含成长、亲密度系统
- **PetGrowthSystem**：宠物成长系统，管理宠物等级、经验
- **IntimacySystem**：宠物亲密度系统，管理宠物与玩家的关系

### 6. 核心游戏系统

- **CombatSystem**：战斗核心系统，处理战斗计算、仇恨管理
- **MovementSystem**：移动系统，处理游戏对象的移动
- **PropertySystem**：属性系统，管理游戏对象的属性
- **SkillSystem**：技能系统，处理技能释放、冷却管理

### 7. 全局系统

- **GuildSystem**：公会系统，处理公会创建、管理
- **AuctionSystem**：拍卖行系统，处理物品拍卖
- **MapSystem**：地图系统，处理地图加载、管理

### 8. 数据库模块

- 多数据库连接管理
- 异步数据库操作
- 数据模型定义
- 数据访问对象（DAO）

### 9. 配置系统

- **ini配置**：服务器基本配置，如端口、数据库连接等
- **Excel配置表**：游戏数据配置，如物品、技能、怪物等
- 支持热更新配置，无需重启服务器

### 10. 日志系统

- 服务器运行日志
- 登录登出日志
- 错误和崩溃日志

## 配置文件

### 1. ini配置文件

`config.ini`：服务器基本配置，包括：

- 服务器基本配置（监听地址、端口、最大连接数等）
- 日志配置（级别、路径、文件大小等）
- 数据库配置（MySQL和MongoDB连接信息）

### 2. Excel配置表

`resources/excel_tables/`：游戏数据配置，包括：

- `item.xlsx`：物品配置
- `skill.xlsx`：技能配置
- `monster.xlsx`：怪物配置
- `npc.xlsx`：NPC配置
- `pet.xlsx`：宠物配置
- `guild.xlsx`：公会配置
- `quest.xlsx`：任务配置
- `map.xlsx`：地图配置
- `shop.xlsx`：商店配置
- `player_level.xlsx`：玩家等级配置

## 快速开始

### 1. 环境要求

- Go 1.16+ 或更高版本
- MySQL 5.7+ 或更高版本
- MongoDB 4.0+ 或更高版本（可选）

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

1. 创建数据库：

   - MySQL：创建账号、游戏和日志数据库
   - MongoDB：创建游戏数据库（可选）
2. 在 `config.ini`中配置数据库连接信息

### 4. 配置Excel表格

- 确保 `resources/excel_tables/`目录下有所有必要的Excel配置表
- 根据游戏需求修改配置表内容

### 5. 编译和运行

```bash
# 编译
go build -o gameserver main.go

# 运行
./gameserver
```

### 6. 客户端测试

```bash
go run client/testclient.go
```

## 开发指南

### 1. 接入新网络消息

#### 步骤1：定义协议

1. 在 `net/protocol/game.proto`文件中添加新的消息类型：

```proto
// 示例：添加公会创建请求和响应消息
message CreateGuildRequest {
    string guild_name = 1;      // 公会名称
    string guild_emblem = 2;    // 公会徽章
}

message CreateGuildResponse {
    int32 result = 1;           // 结果（0成功，非0失败）
    int64 guild_id = 2;         // 公会ID
    string error_message = 3;   // 错误信息
}
```

#### 步骤2：生成Go代码

执行以下命令生成Go代码：

```bash
protoc --go_out=. net/protocol/game.proto
```

#### 步骤3：创建消息处理器

在 `net/handler/`目录下创建新的处理器文件，如 `guild_handler.go`：

```go
package handler

import (
    "github.com/pzqf/zGameServer/game/guild"
    "github.com/pzqf/zGameServer/net/protocol"
)

// HandleCreateGuild 处理创建公会请求
func (h *GameHandler) HandleCreateGuild(request *protocol.CreateGuildRequest, session interface{}) {
    // 获取会话信息
    playerSession := session.(*PlayerSession)
    playerId := playerSession.PlayerId
  
    // 调用公会服务创建公会
    guildId, err := guild.GetGuildService().CreateGuild(playerId, request.GuildName, request.GuildEmblem)
  
    // 构建响应
    response := &protocol.CreateGuildResponse{
        Result: 0,
        GuildId: guildId,
    }
  
    if err != nil {
        response.Result = 1
        response.ErrorMessage = err.Error()
    }
  
    // 发送响应
    playerSession.SendMessage(response)
}
```

#### 步骤4：注册消息路由

在 `net/router/router.go`文件中注册新的消息路由：

```go
// 注册消息路由
func (r *GameRouter) RegisterRoutes() {
    // 现有路由...
  
    // 注册公会相关路由
    r.RegisterRoute("CreateGuildRequest", h.HandleCreateGuild)
}
```

#### 步骤5：在客户端实现

在客户端代码中实现对应的消息发送和处理逻辑。

### 2. 增加读写数据库代码

#### 步骤1：定义数据模型

在 `db/models/`目录下创建新的数据模型文件，如 `guild_model.go`：

```go
package models

import (
    "time"
)

// Guild 公会数据模型
type Guild struct {
    GuildId        int64     `json:"guild_id" gorm:"primaryKey"`
    GuildName      string    `json:"guild_name" gorm:"size:50;not null"`
    GuildEmblem    string    `json:"guild_emblem" gorm:"size:255"`
    LeaderId       int64     `json:"leader_id" gorm:"not null"`
    Level          int       `json:"level" gorm:"default:1"`
    Exp            int64     `json:"exp" gorm:"default:0"`
    MemberCount    int       `json:"member_count" gorm:"default:1"`
    MaxMembers     int       `json:"max_members" gorm:"default:20"`
    Notice         string    `json:"notice" gorm:"size:500"`
    CreateTime     time.Time `json:"create_time" gorm:"autoCreateTime"`
    LastUpdateTime time.Time `json:"last_update_time" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Guild) TableName() string {
    return "guilds"
}

// GuildMember 公会成员数据模型
type GuildMember struct {
    Id            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    GuildId       int64     `json:"guild_id" gorm:"index;not null"`
    PlayerId      int64     `json:"player_id" gorm:"index;not null"`
    Name          string    `json:"name" gorm:"size:50;not null"`
    Position      int       `json:"position" gorm:"not null"` // 职位：0普通成员，1官员，2副会长，3会长
    Contribution  int64     `json:"contribution" gorm:"default:0"`
    JoinTime      int64     `json:"join_time" gorm:"not null"`
    LastOnline    int64     `json:"last_online" gorm:"not null"`
    CreateTime    time.Time `json:"create_time" gorm:"autoCreateTime"`
    LastUpdateTime time.Time `json:"last_update_time" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (GuildMember) TableName() string {
    return "guild_members"
}
```

#### 步骤2：创建数据访问对象（DAO）

在 `db/dao/`目录下创建新的DAO文件，如 `guild_dao.go`：

```go
package dao

import (
    "github.com/pzqf/zGameServer/db/models"
)

// GuildDAO 公会数据访问对象
type GuildDAO struct {
    BaseDAO
}

// NewGuildDAO 创建公会DAO实例
func NewGuildDAO() *GuildDAO {
    return &GuildDAO{}
}

// Create 创建公会记录
func (dao *GuildDAO) Create(guild *models.Guild) error {
    return dao.DB.Create(guild).Error
}

// GetByID 根据ID获取公会
func (dao *GuildDAO) GetByID(guildId int64) (*models.Guild, error) {
    var guild models.Guild
    err := dao.DB.First(&guild, guildId).Error
    if err != nil {
        return nil, err
    }
    return &guild, nil
}

// Update 更新公会信息
func (dao *GuildDAO) Update(guild *models.Guild) error {
    return dao.DB.Save(guild).Error
}

// Delete 删除公会
func (dao *GuildDAO) Delete(guildId int64) error {
    return dao.DB.Delete(&models.Guild{}, guildId).Error
}

// GuildMemberDAO 公会成员数据访问对象
type GuildMemberDAO struct {
    BaseDAO
}

// NewGuildMemberDAO 创建公会成员DAO实例
func NewGuildMemberDAO() *GuildMemberDAO {
    return &GuildMemberDAO{}
}

// Create 创建公会成员记录
func (dao *GuildMemberDAO) Create(member *models.GuildMember) error {
    return dao.DB.Create(member).Error
}

// GetByGuildID 获取公会所有成员
func (dao *GuildMemberDAO) GetByGuildID(guildId int64) ([]*models.GuildMember, error) {
    var members []*models.GuildMember
    err := dao.DB.Where("guild_id = ?", guildId).Find(&members).Error
    if err != nil {
        return nil, err
    }
    return members, nil
}

// GetByPlayerID 根据玩家ID获取公会成员信息
func (dao *GuildMemberDAO) GetByPlayerID(playerId int64) (*models.GuildMember, error) {
    var member models.GuildMember
    err := dao.DB.Where("player_id = ?", playerId).First(&member).Error
    if err != nil {
        return nil, err
    }
    return &member, nil
}

// Update 更新公会成员信息
func (dao *GuildMemberDAO) Update(member *models.GuildMember) error {
    return dao.DB.Save(member).Error
}

// Delete 删除公会成员
func (dao *GuildMemberDAO) Delete(guildId int64, playerId int64) error {
    return dao.DB.Where("guild_id = ? AND player_id = ?", guildId, playerId).Delete(&models.GuildMember{}).Error
}
```

#### 步骤3：在服务中使用DAO

在服务代码中使用DAO进行数据库操作：

```go
package guild

import (
    "github.com/pzqf/zGameServer/db/dao"
    "github.com/pzqf/zGameServer/db/models"
)

// CreateGuild 创建公会
func (gs *GuildService) CreateGuild(leaderId int64, guildName string, guildEmblem string) (int64, error) {
    // 生成公会ID
    guildId := generateGuildId()
  
    // 创建公会数据模型
    guildModel := &models.Guild{
        GuildId:     guildId,
        GuildName:   guildName,
        GuildEmblem: guildEmblem,
        LeaderId:    leaderId,
        Level:       1,
        Exp:         0,
        MemberCount: 1,
        MaxMembers:  20,
        Notice:      "",
    }
  
    // 保存到数据库
    guildDAO := dao.NewGuildDAO()
    if err := guildDAO.Create(guildModel); err != nil {
        return 0, err
    }
  
    // 创建公会对象并缓存
    guild := NewGuild(guildModel)
    gs.guilds.Store(guildId, guild)
  
    return guildId, nil
}
```

### 3. 增加新模块

#### 3.1 全局模块

全局模块是指在服务器运行期间存在且唯一的模块，如公会系统、拍卖行系统等。

##### 步骤1：创建模块目录结构

在 `game/`目录下创建新的模块目录，如 `guild/`：

```
game/guild/
├── guild.go               # 公会数据结构和核心逻辑
├── guild_service.go       # 公会服务
└── guild_const.go         # 公会相关常量
```

##### 步骤2：实现模块核心逻辑

在 `guild.go`中实现公会数据结构和核心逻辑：

```go
package guild

import (
    "sync"
    "time"

    "github.com/pzqf/zEngine/zLog"
    "github.com/pzqf/zUtil/zMap"
    "go.uber.org/zap"
)

// Guild 公会结构体
type Guild struct {
    GuildId       int64               // 公会ID
    GuildName     string              // 公会名称
    GuildEmblem   string              // 公会徽章
    LeaderId      int64               // 会长ID
    Level         int                 // 公会等级
    Exp           int64               // 公会经验
    MemberCount   int                 // 成员数量
    MaxMembers    int                 // 最大成员数量
    Notice        string              // 公会公告
    Members       sync.Map            // 公会成员（PlayerId -> *GuildMember）
    Applies       sync.Map            // 公会申请（ApplyId -> *GuildApply）
    PermissionConfig map[int]int64    // 职位权限配置
}

// GuildMember 公会成员结构体
type GuildMember struct {
    PlayerId     int64  // 玩家ID
    Name         string // 玩家名称
    Position     int    // 职位
    Contribution int64  // 贡献值
    JoinTime     int64  // 加入时间
    Online       bool   // 是否在线
    LastOnline   int64  // 最后在线时间
}

// NewGuild 创建新公会
func NewGuild(guildModel *models.Guild) *Guild {
    guild := &Guild{
        GuildId:       guildModel.GuildId,
        GuildName:     guildModel.GuildName,
        GuildEmblem:   guildModel.GuildEmblem,
        LeaderId:      guildModel.LeaderId,
        Level:         guildModel.Level,
        Exp:           guildModel.Exp,
        MemberCount:   guildModel.MemberCount,
        MaxMembers:    guildModel.MaxMembers,
        Notice:        guildModel.Notice,
        Members:       sync.Map{},
        Applies:       sync.Map{},
        PermissionConfig: map[int]int64{
            GuildPositionLeader:    GuildPermissionAll,
            GuildPositionViceLeader: GuildPermissionKick | GuildPermissionSetPosition | GuildPermissionUpdateNotice,
            GuildPositionOfficer:    GuildPermissionKick,
            GuildPositionMember:     0,
        },
    }
    return guild
}

// 公会方法实现...
```

##### 步骤3：实现模块服务

在 `guild_service.go`中实现模块服务：

```go
package guild

import (
    "sync"

    "github.com/pzqf/zEngine/zLog"
    "github.com/pzqf/zUtil/zMap"
    "go.uber.org/zap"
)

// GuildService 公会服务
type GuildService struct {
    guilds      sync.Map // 公会ID -> *Guild
    playerGuild sync.Map // 玩家ID -> 公会ID
}

var (
    guildService *GuildService
    once         sync.Once
)

// GetGuildService 获取公会服务单例
func GetGuildService() *GuildService {
    once.Do(func() {
        guildService = &GuildService{
            guilds:      sync.Map{},
            playerGuild: sync.Map{},
        }
    })
    return guildService
}

// CreateGuild 创建公会
func (gs *GuildService) CreateGuild(leaderId int64, guildName string, guildEmblem string) (int64, error) {
    // 实现逻辑...
}

// GetGuild 根据ID获取公会
func (gs *GuildService) GetGuild(guildId int64) (*Guild, bool) {
    guild, exists := gs.guilds.Load(guildId)
    if !exists {
        return nil, false
    }
    return guild.(*Guild), true
}

// GetPlayerGuild 获取玩家所在公会
func (gs *GuildService) GetPlayerGuild(playerId int64) (*Guild, bool) {
    guildId, exists := gs.playerGuild.Load(playerId)
    if !exists {
        return nil, false
    }
    return gs.GetGuild(guildId.(int64))
}

// 其他服务方法...
```

##### 步骤4：注册模块到游戏服务器

在游戏服务器初始化时注册模块：

```go
// 初始化游戏服务
func InitGameServices() {
    // 初始化现有服务...
  
    // 初始化公会服务
    guild.GetGuildService()
}
```

#### 3.2 玩家独有模块

玩家独有模块是指每个玩家拥有独立实例的模块，如背包系统、装备系统等。

##### 步骤1：创建模块目录结构

在 `game/player/`目录下创建新的模块目录，如 `inventory/`：

```
game/player/inventory/
├── inventory.go           # 背包核心逻辑
└── inventory_component.go # 背包组件
```

##### 步骤2：实现模块核心逻辑

在 `inventory.go`中实现背包核心逻辑：

```go
package inventory

import (
    "sync"

    "github.com/pzqf/zGameServer/game/object"
)

// Inventory 背包结构体
type Inventory struct {
    PlayerId    int64              // 玩家ID
    Capacity    int                // 背包容量
    Items       sync.Map           // 物品（ItemId -> *Item）
    player      *object.Player     // 关联的玩家对象
}

// Item 物品结构体
type Item struct {
    ItemId      int64  // 物品ID
    TemplateId  int32  // 物品模板ID
    Count       int32  // 数量
    Position    int32  // 位置
    Durability  int32  // 耐久度
}

// NewInventory 创建新背包
func NewInventory(playerId int64, capacity int, player *object.Player) *Inventory {
    return &Inventory{
        PlayerId: playerId,
        Capacity: capacity,
        Items:    sync.Map{},
        player:   player,
    }
}

// AddItem 添加物品
func (inv *Inventory) AddItem(templateId int32, count int32) error {
    // 实现逻辑...
}

// RemoveItem 移除物品
func (inv *Inventory) RemoveItem(itemId int64, count int32) error {
    // 实现逻辑...
}

// GetItem 获取物品
func (inv *Inventory) GetItem(itemId int64) (*Item, bool) {
    item, exists := inv.Items.Load(itemId)
    if !exists {
        return nil, false
    }
    return item.(*Item), true
}

// GetItems 获取所有物品
func (inv *Inventory) GetItems() []*Item {
    var items []*Item
    inv.Items.Range(func(key, value interface{}) bool {
        items = append(items, value.(*Item))
        return true
    })
    return items
}

// 其他背包方法...
```

##### 步骤3：实现模块组件

在 `inventory_component.go`中实现背包组件，使其能够附加到玩家对象上：

```go
package inventory

import (
    "github.com/pzqf/zGameServer/game/object"
    "github.com/pzqf/zGameServer/game/object/component"
)

// InventoryComponent 背包组件
type InventoryComponent struct {
    component.BaseComponent
    inventory *Inventory
}

// NewInventoryComponent 创建背包组件
func NewInventoryComponent(player *object.Player) *InventoryComponent {
    component := &InventoryComponent{
        BaseComponent: component.BaseComponent{
            Owner: player,
        },
    }
  
    // 创建背包
    component.inventory = NewInventory(player.GetID(), 50, player)
  
    return component
}

// GetInventory 获取背包
func (c *InventoryComponent) GetInventory() *Inventory {
    return c.inventory
}

// OnAttach 组件附加时调用
func (c *InventoryComponent) OnAttach() {
    // 附加逻辑...
}

// OnDetach 组件分离时调用
func (c *InventoryComponent) OnDetach() {
    // 分离逻辑...
}

// GetName 获取组件名称
func (c *InventoryComponent) GetName() string {
    return "Inventory"
}
```

##### 步骤4：在玩家创建时附加组件

在玩家创建时附加背包组件：

```go
package player

import (
    "github.com/pzqf/zGameServer/game/object"
    "github.com/pzqf/zGameServer/game/player/inventory"
)

// CreatePlayer 创建玩家对象
func CreatePlayer(playerId int64, name string) *object.Player {
    player := object.NewPlayer(playerId, name)
  
    // 附加背包组件
    inventoryComponent := inventory.NewInventoryComponent(player)
    player.AddComponent(inventoryComponent)
  
    // 附加其他组件...
  
    return player
}
```

##### 步骤5：使用玩家独有模块

通过玩家对象获取模块实例：

```go
// 获取玩家背包
func GetPlayerInventory(player *object.Player) *inventory.Inventory {
    inventoryComponent := player.GetComponent("Inventory")
    if inventoryComponent == nil {
        return nil
    }
    return inventoryComponent.(*inventory.InventoryComponent).GetInventory()
}

// 使用背包
func UseItem(player *object.Player, itemId int64) error {
    inv := GetPlayerInventory(player)
    if inv == nil {
        return errors.New("inventory not found")
    }
  
    item, exists := inv.GetItem(itemId)
    if !exists {
        return errors.New("item not found")
    }
  
    // 使用物品逻辑...
  
    return nil
}
```

### 4. 扩展现有系统

#### 4.1 组件扩展

利用组件系统，为游戏对象添加新组件：

1. 在 `game/object/component/`中创建新组件
2. 实现组件接口
3. 将组件附加到游戏对象

#### 4.2 系统功能扩展

为现有系统添加新功能：

1. 遵循现有代码风格和架构模式
2. 实现新功能逻辑
3. 提供必要的接口

#### 4.3 数据库扩展

在 `db/models/`中添加新的数据库模型，在 `db/dao/`中添加对应的数据访问方法。

## 测试

运行测试：

```bash
go test ./...
```

## 日志

日志文件默认输出到控制台，可根据配置文件修改为输出到文件。

## 安全

### 1. 网络安全

- 验证所有客户端输入，防止注入攻击
- 使用加密传输敏感数据
- 限制客户端请求频率，防止DoS攻击

### 2. 数据库安全

- 使用参数化查询，防止SQL注入
- 数据库密码等敏感信息通过配置文件管理，不硬编码
- 定期备份数据库

### 3. 服务器安全

- 限制服务器端口访问
- 定期更新服务器软件
- 监控服务器状态，及时发现异常

## 性能优化

### 1. 网络优化

- 使用连接池管理网络连接
- 优化协议结构，减少数据传输量
- 使用压缩算法减少数据包大小

### 2. 内存优化

- 合理使用对象池，减少GC压力
- 避免频繁的内存分配和释放
- 使用适当的数据结构，平衡内存使用和性能

### 3. 数据库优化

- 使用索引优化数据库查询
- 合理设计数据库表结构
- 使用缓存减少数据库访问

### 4. 并发优化

- 合理使用goroutine，避免过度并发
- 使用适当的同步机制，避免竞态条件
- 优化锁的粒度，减少锁竞争

## 注意事项

### 1. 代码规范

- **命名规范**：

  - 包名：小写，使用简短的名词
  - 函数名：驼峰命名，首字母大写表示可导出
  - 变量名：驼峰命名，首字母小写表示私有
- **代码风格**：

  - 遵循Go语言标准代码风格
  - 使用 `go fmt`格式化代码
  - 代码注释清晰，解释关键逻辑

### 2. 常见问题

- **网络连接问题**：

  - 检查网络配置和防火墙设置
  - 确保客户端和服务器使用相同的协议版本
- **数据库连接问题**：

  - 检查数据库配置是否正确
  - 确保数据库服务已启动
  - 检查数据库用户权限
- **性能问题**：

  - 检查服务器资源使用情况
  - 优化代码逻辑，减少不必要的计算
  - 使用性能分析工具定位瓶颈
- **配置问题**：

  - 确保配置文件格式正确
  - 检查配置项是否完整
  - 注意配置文件的大小写敏感问题

## 未来规划

- 支持更多游戏类型和玩法
- 提供更多工具和脚本，简化开发流程
- 增强监控和运维工具，提高服务器稳定性
- 支持更多平台和设备，扩大应用范围
- 分布式服务器支持
- 负载均衡
- 热更新功能
- 监控系统

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License
