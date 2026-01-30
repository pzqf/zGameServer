
# zGameServer

一个基于Go语言开发的MMO游戏服务器框架，采用模块化设计，具有良好的可扩展性和高性能。

## 项目架构详解

### 1. 项目整体架构

zGameServer采用三层架构设计，职责清晰：

```
┌─────────────────────────────────────────────────────────────────────┐
│                         zGameServer (业务层)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Player   │ │ Monster  │ │ Guild    │ │ Auction  │ │ Map      │ │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │ │ Service  │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                    Game Logic Systems                         │ │
│  │  AI System | Combat System | Skill System | Buff System       │ │
│  │  Movement | Property System | Object Manager                  │ │
│  └───────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                         zEngine (引擎层)                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Service  │ │ Actor    │ │ Event    │ │ Net      │ │ Script   │ │
│  │ Manager  │ │ System   │ │ Bus      │ │ Layer    │ │ Engine   │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Log      │ │ Inject   │ │ System   │ │ Object   │ │ Etcd     │ │
│  │ System   │ │ DI       │ │ Manager  │ │ Pool     │ │ Client   │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                         zUtil (工具层)                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ DataConv │ │ Cache    │ │ Map      │ │ Queue    │ │ Stack    │ │
│  │ Color    │ │ Crypto   │ │ Gps      │ │ File     │ │ String   │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Error    │ │ Time     │ │ Tree     │ │ List     │ │ Hash     │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### 2. 核心架构设计模式

#### 2.1 服务架构 (Service Architecture)

```go
type GameServer struct {
    *zService.ServiceManager   // 服务管理器（继承）
    wg            sync.WaitGroup
    packetRouter  *router.PacketRouter
    protocol      protolayer.Protocol
    objectManager *zObject.ObjectManager
}
```

**特点：**
- **服务拓扑排序**：自动计算服务依赖关系，确保有序启动/关闭
- **服务状态管理**：Created → Init → Running → Stopping → Stopped
- **依赖注入 (DI)**：基于名称的依赖注入容器
- **并行启动**：每个服务在独立goroutine中运行

#### 2.2 Actor模型 (Actor Model)

```go
type PlayerActor struct {
    *zActor.BaseActor
    Player *Player
}
```

**核心特点：**
- **消息驱动**：所有通信通过消息队列异步处理
- **并发隔离**：每个Actor拥有独立状态，避免竞态条件
- **全局系统**：统一管理所有Actor实例
- **类型安全**：强类型消息定义

#### 2.3 事件驱动架构 (Event-Driven Architecture)

```go
type EventBus struct {
    handlers map[EventType][]EventHandler
    mu       sync.RWMutex
    running  atomic.Bool
}
```

**核心功能：**
- **异步事件发布**：非阻塞式事件分发
- **事件订阅**：支持多订阅者监听同一事件
- **事件同步**：支持同步阻塞处理
- **事件监控**：事件处理统计和异常捕获

#### 2.4 ECS架构 (Entity-Component-System)

```go
type GameObject struct {
    *zObject.BaseObject
    name         string
    objectType   GameObjectType
    position     Vector3
    eventEmitter *zEvent.EventBus
    components   *component.ComponentManager
}
```

**ECS组成：**
- **Entity (实体)**：唯一标识符，无行为
- **Component (组件)**：纯数据容器
  - PropertyComponent：属性管理
  - CombatComponent：战斗逻辑
  - SkillComponent：技能系统
  - BuffComponent：Buff效果
  - MovementComponent：移动控制
- **System (系统)**：行为逻辑处理
  - AISystem：AI决策
  - CombatSystem：战斗计算
  - BuffSystem：Buff管理
  - PropertySystem：属性计算

#### 2.5 对象池设计 (Object Pool)

```go
type GenericPool struct {
    mu      sync.Mutex
    objects []interface{}
    newFunc func() interface{}
    maxSize int
}
```

**应用场景：**
- **技能对象池**：技能频繁创建/销毁
- **Buff对象池**：Buff效果管理
- **Actor对象池**：PlayerActor复用

#### 2.6 网络层架构 (Network Architecture)

```go
type Protocol interface {
    Encode(protoId int32, version int32, data interface{}) (*zNet.NetPacket, error)
    Decode(packet *zNet.NetPacket) (interface{}, error)
}
```

**支持多种协议：**
- ProtocolTypeProtobuf
- ProtocolTypeJSON  
- ProtocolTypeXML

**网络层特点：**
- **DDoS防护**：IP限流、连接控制
- **数据包路由**：基于ProtoId的消息分发
- **网络指标**：延迟监控、吞吐量统计
- **安全传输**：RSA加密 + AESEncryption

## 项目特点

### 1. 模块化设计
- 清晰的代码结构，便于维护和扩展
- 各模块低耦合，高内聚
- 易于进行单元测试和集成测试

### 2. 多数据库支持
- 支持MySQL和MongoDB数据库
- 异步数据库操作，提高服务器性能
- 数据模型与业务逻辑分离

### 3. 配置化管理
- 通过配置文件和Excel表格管理服务器参数和游戏数据
- 支持热更新配置，无需重启服务器
- 方便游戏策划和运营人员进行数值调整

### 4. 完善的日志系统
- 使用zap日志框架，支持不同级别日志输出
- 结构化日志，便于数据分析和监控
- 日志轮转和压缩，节省存储空间

### 5. 协议缓冲区
- 使用Protocol Buffers进行高效的网络通信
- 支持多种协议格式，满足不同场景需求
- 可插拔协议设计，易于扩展新协议

### 6. 安全机制
- 账号密码验证和角色数据保护
- DDoS防护机制，限制IP连接数和流量
- 协议验证和数据完整性检查

### 7. 崩溃恢复
- 服务器崩溃时自动捕获堆栈信息并记录到日志
- 完善的错误处理和恢复机制
- 提高系统稳定性和可维护性

### 8. 组件系统
- 基于组件的游戏对象管理
- 高度的灵活性和可扩展性
- 动态添加和移除组件，实现功能扩展

### 9. 依赖注入容器
- 实现了zDI包，支持单例和工厂两种依赖类型
- 提高代码的可维护性和可测试性
- 自动依赖解析，简化开发流程

### 10. 服务自动注册和发现
- 扩展了zService包，实现了服务的自动注册和发现机制
- 支持服务的依赖管理，确保服务按正确顺序启动
- 服务状态监控，便于系统维护

### 11. 监控系统
- Prometheus指标，通过 `/metrics` 端点暴露服务器运行指标
- 网络指标：连接数、延迟、吞吐量等
- 业务指标：玩家在线数、公会数量、拍卖行交易等
- 服务器状态：CPU、内存、GC等

## 技术栈

- **语言**：Go 1.25.5
- **网络**：TCP和HTTP协议，Protocol Buffers
- **数据库**：MySQL、MongoDB
- **日志**：zap日志框架
- **配置**：ini配置文件，Excel表格
- **依赖管理**：Go Modules
- **依赖注入**：zDI包，支持单例和工厂两种依赖类型
- **服务管理**：zService包，支持服务的自动注册和发现，依赖管理

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
│       ├── ai/             # AI系统
│       ├── buff/           # Buff系统
│       ├── combat/         # 战斗系统
│       ├── movement/       # 移动系统
│       ├── property/       # 属性系统
│       └── skill/          # 技能系统
├── gameserver/             # 服务器核心
├── metrics/                # 监控系统
├── net/                    # 网络相关代码
│   ├── handler/            # 请求处理器
│   ├── metrics/            # 网络指标
│   ├── protocol/           # 协议定义
│   ├── protolayer/         # 协议层实现
│   ├── router/             # 路由管理
│   └── service/            # 网络服务
├── resources/              # 资源文件
│   ├── excel_tables/       # Excel配置表
│   └── maps/               # 地图资源
├── util/                   # 工具类
├── config.ini              # 配置文件
├── go.mod                  # Go模块依赖
├── go.sum                  # 依赖校验
└── main.go                 # 主入口文件
```

## 核心功能模块

### 1. 网络服务

- **TCP服务**：处理客户端实时连接
- **HTTP服务**：处理非实时请求（如充值下发等）
- **数据包路由和处理**：基于ProtoId的消息分发
- **会话管理**：玩家会话的创建、维护和销毁
- **支持多种协议格式**：Protobuf、JSON、XML
- **工作线程池**：并行处理数据包，提高服务器性能
- **DDoS保护**：限制每个IP的连接数、数据包数和流量

### 2. 游戏对象系统

- **GameObject**：基础游戏对象，包含ID、名称、位置等基本属性
- **LivingObject**：生命对象，继承自GameObject，增加生命值等属性
- **Component**：组件系统，允许动态为游戏对象添加功能
- **ComponentManager**：组件管理器，负责组件的添加、移除和管理

### 3. 玩家系统

- **账号创建和登录**：账号验证、角色选择
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

- **多数据库连接管理**：支持MySQL和MongoDB
- **异步数据库操作**：所有数据库操作均为异步，提高服务器性能
- **数据模型定义**：清晰的数据模型设计
- **数据访问对象（DAO）**：封装数据库访问逻辑
- **Repository层**：业务数据仓库，处理业务数据操作

### 9. 配置系统

- **ini配置**：服务器基本配置，如端口、数据库连接等
- **Excel配置表**：游戏数据配置，如物品、技能、怪物等
- **DDoS保护配置**：限制每个IP的连接数、数据包数、流量
- **支持热更新配置**：无需重启服务器

### 10. 日志系统

- **服务器运行日志**：记录服务器启动、关闭等关键事件
- **登录登出日志**：记录玩家的登录和登出行为
- **错误和崩溃日志**：记录服务器错误和崩溃信息
- **结构化日志**：便于数据分析和监控

### 11. 监控系统

- **Prometheus指标**：通过 `/metrics` 端点暴露服务器运行指标
- **网络指标**：连接数、延迟、吞吐量等
- **业务指标**：玩家在线数、公会数量、拍卖行交易等
- **服务器状态**：CPU、内存、GC等

### 12. 依赖注入容器

- **zDI包**：实现了功能完整的依赖注入容器
- **支持单例和工厂**：两种依赖类型，满足不同场景的需求
- **线程安全**：容器的实现考虑了并发安全性
- **依赖管理**：提供了依赖注册、解析、检查和管理的完整功能

### 13. 服务管理系统

- **zService包**：扩展了服务管理功能
- **服务自动注册和发现**：实现了服务的自动注册和发现机制
- **服务依赖管理**：支持服务的依赖管理，确保服务按照正确的顺序初始化和启动
- **服务生命周期管理**：管理服务的创建、初始化、运行和停止等生命周期

## 新手接入步骤

### 步骤1：环境准备

1. **安装Go环境**
   - 下载并安装Go 1.25.5或更高版本：https://golang.org/dl/
   - 配置GOPATH环境变量
   - 验证安装：`go version`

2. **安装数据库**
   - 安装MySQL 5.7+ 或更高版本
   - 安装MongoDB 4.0+ 或更高版本（可选）
   - 创建数据库和用户

3. **安装Protobuf编译器**
   - 下载Protobuf编译器：https://github.com/protocolbuffers/protobuf/releases
   - 配置环境变量，确保 `protoc` 命令可用

### 步骤2：克隆项目

```bash
git clone https://github.com/pzqf/zGameServer.git
cd zGameServer
```

### 步骤3：配置项目

1. **配置Go模块**

```bash
go mod tidy
```

2. **配置数据库**

- 打开 `config.ini` 文件
- 配置数据库连接信息：

```ini
[database.account]
host = localhost
port = 27017
user = 
password = 
dbname = account
charset = 
max_idle = 10
max_open = 100
driver = mongo
uri = mongodb://localhost:27017/account
max_pool_size = 100
min_pool_size = 10
connect_timeout = 30

[database.game]
host = localhost
port = 27017
user = 
password = 
dbname = game
charset = 
max_idle = 10
max_open = 100
driver = mongo
uri = mongodb://localhost:27017/game
max_pool_size = 100
min_pool_size = 10
connect_timeout = 30

[database.log]
host = localhost
port = 27017
user = 
password = 
dbname = log
charset = 
max_idle = 10
max_open = 100
driver = mongo
uri = mongodb://localhost:27017/log
max_pool_size = 100
min_pool_size = 10
connect_timeout = 30
```

3. **配置Excel表格**

- 确保 `resources/excel_tables/` 目录下有所有必要的Excel配置表
- 根据游戏需求修改配置表内容

### 步骤4：运行项目

1. **编译项目**

```bash
go build -o gameserver main.go
```

2. **运行服务器**

```bash
./gameserver
```

3. **运行测试客户端**

```bash
go run client/testclient.go
```

### 步骤5：测试服务器

1. **检查日志**
   - 查看 `logs/server.log` 文件
   - 确认服务器正常启动

2. **测试网络连接**
   - 使用telnet测试TCP连接：`telnet localhost 8888`
   - 使用浏览器测试HTTP连接：http://localhost:8080/metrics

3. **查看监控指标**
   - 访问Prometheus指标端点：http://localhost:8080/metrics
   - 查看服务器运行状态

### 步骤6：开始开发

1. **了解项目结构**

- 阅读本README文件
- 查看源代码注释和示例代码
- 浏览源代码，了解各模块功能

2. **学习核心模块**

- **网络模块**：`net/service/tcp_service.go`
- **玩家系统**：`game/player/player.go`
- **战斗系统**：`game/systems/combat/combat_system.go`
- **配置系统**：`config/ini_config.go`

3. **参与开发**

- 选择一个模块进行学习
- 阅读相关代码和文档
- 尝试修改代码并测试
- 提交代码到版本控制系统

## 开发指南

### 1. 依赖注入容器使用

#### 1.1 注册依赖

```go
import (
    "github.com/pzqf/zEngine/zInject"
)

// 创建依赖注入容器
container := zInject.NewContainer()

// 注册单例依赖
container.RegisterSingleton("config", configInstance)

// 注册工厂依赖
container.Register("playerService", func() interface{} {
    return NewPlayerService()
})
```

#### 1.2 解析依赖

```go
// 解析单例依赖
config, err := container.Resolve("config")
if err != nil {
    // 处理错误
}

// 解析工厂依赖
playerService, err := container.Resolve("playerService")
if err != nil {
    // 处理错误
}
```

#### 1.3 在GameServer中使用

```go
// 注册核心依赖
gameServer.RegisterSingleton("protocol", protolayer.NewProtobufProtocol())

// 解析依赖
protocol, err := gameServer.ResolveDependency("protocol")
if err != nil {
    // 处理错误
}
```

### 2. 服务管理使用

#### 2.1 注册服务

```go
// 注册服务到GameServer
gameServer.RegisterService(func() zService.Service {
    return NewPlayerService()
})

// 注册带依赖的服务
gameServer.RegisterService(func() zService.Service {
    return NewGuildService()
}, "playerService") // 依赖playerService
```

#### 2.2 自动注册服务

```go
// 自动注册所有已注册的服务
err := gameServer.AutoRegisterServices()
if err != nil {
    // 处理错误
}
```

### 3. 接入新网络消息

#### 步骤1：定义协议

1. 在 `net/protocol/game.proto` 文件中添加新的消息类型：

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

在 `net/handler/` 目录下创建新的处理器文件，如 `guild_handler.go`：

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

在 `net/router/router.go` 文件中注册新的消息路由：

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

### 4. 增加读写数据库代码

#### 步骤1：定义数据模型

在 `db/models/` 目录下创建新的数据模型文件，如 `guild_model.go`：

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

在 `db/dao/` 目录下创建新的DAO文件，如 `guild_dao.go`：

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

## 配置文件

### 1. ini配置文件

`config.ini`：服务器基本配置，包括：

- 服务器基本配置（监听地址、端口、最大连接数等）
- HTTP配置（基于zNet.HttpConfig）：包括监听地址、最大客户端数、最大数据包大小等
- 日志配置（基于zLog.Config）：包括级别、路径、文件大小、最大天数等
- **DDoS保护配置**：包括每个IP的最大连接数、最大数据包数、最大流量、封禁时间等
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

- Go 1.25.5+ 或更高版本
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
2. 在 `config.ini` 中配置数据库连接信息

### 4. 配置Excel表格

- 确保 `resources/excel_tables/` 目录下有所有必要的Excel配置表
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

## 架构优势

### 1. 高并发支持

- **Actor模型**：单线程处理，避免竞态条件
- **协程调度**：Goroutine轻量级并发
- **无锁设计**：原子操作、无锁数据结构

### 2. 易维护性

- **ECS架构**：数据与逻辑分离
- **模块化设计**：低耦合、高内聚
- **依赖注入**：组件依赖解耦

### 3. 易扩展

- **服务插件化**：动态加载/卸载服务
- **协议扩展**：支持多种数据格式
- **配置驱动**：Excel配置表灵活修改

### 4. 监控完善

- **多维度指标**：网络、业务、系统
- **可视化监控**：Prometheus集成
- **异常报警**：错误自动捕获

## 总结

这个游戏服务器系统是一个**设计精良、架构合理**的企业级解决方案，主要特点：

1. **分层设计**：业务层 → 引擎层 → 工具层，职责清晰
2. **模式运用**：Actor、ECS、事件驱动、对象池等成熟模式
3. **高性能**：Goroutine、对象池、缓存优化等
4. **高并发**：消息驱动、无锁设计、原子操作

## 许可证

MIT

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如果您有任何问题或建议，欢迎通过GitHub Issues与我们联系。

---

**zGameServer** - 简单高效的游戏服务器框架，为您的游戏提供强大的基础架构支持！

