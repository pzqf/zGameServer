# zGameServer

## 🏗️ 游戏服务器架构设计文档

zGameServer是一个基于Go语言开发的MMO游戏服务器框架，采用模块化设计，具有良好的可扩展性和高性能。

---

## 📋 目录

- [项目概述](#-项目概述)
- [整体架构设计](#-整体架构设计)
- [核心架构设计模式](#-核心架构设计模式)
- [模块架构详解](#-模块架构详解)
- [实现方式详解](#-实现方式详解)
- [新手使用指引](#-新手使用指引)
- [最佳实践](#-最佳实践)

---

## 🎯 项目概述

### 项目定位

zGameServer是一个**高性能、可扩展、模块化**的MMO游戏服务器框架，专为大型多人在线游戏设计。

### 技术栈

- **语言**：Go 1.25+
- **网络**：TCP、UDP、WebSocket、HTTP
- **协议**：Protocol Buffers、JSON、XML
- **数据库**：MySQL、MongoDB
- **日志**：zap日志框架
- **配置**：ini配置文件、Excel表格
- **依赖注入**：zInject包
- **服务管理**：zService包
- **监控**：Prometheus指标

### 项目特点

1. **三层架构设计** - 业务层、引擎层、工具层，职责清晰
2. **高性能** - 基于Go的并发特性，支持高并发在线
3. **可扩展** - 模块化设计，易于扩展新功能
4. **易维护** - 代码结构清晰，便于维护和调试
5. **完整功能** - 包含玩家系统、战斗系统、公会系统、拍卖行系统等

---

## 🗺️ 整体架构设计

### 三层架构设计

```
┌─────────────────────────────────────────────────────────────────────┐
│                         zGameServer (业务层)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Player   │ │ Monster  │ │ Guild    │ │ Auction  │ │ Map      │ │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │ │ Service  │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                    Game Logic Systems                         │ │
│  │  AISystem | CombatSystem | SkillSystem | BuffSystem           │ │
│  │  MovementSystem | PropertySystem | ObjectManager             │ │
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

### 架构设计理念

#### 1. 模块化设计理念

**核心理念：单一职责、低耦合、高内聚**

```go
// 每个模块负责一个特定的功能
type Player struct {
    *object.LivingObject
    playerId int64
    session  *zNet.TcpServerSession
}

type CombatSystem struct {
    attacks map[int64]*AttackRecord
    mu      sync.RWMutex
}

type GuildSystem struct {
    guilds map[int64]*Guild
    mu     sync.RWMutex
}
```

#### 2. 事件驱动设计理念

**核心理念：通过事件实现模块间解耦通信**

```go
// 事件定义
const (
    EventTypeUserLogin EventType = iota
    EventTypeUserLogout
    EventTypeCharacterCreated
    EventTypeCharacterSelected
    EventTypeCharacterEnteredMap
    EventTypeCharacterLeftMap
    EventTypePlayerAttack
    EventTypePlayerMove
    EventTypePlayerSkill
)

// 事件使用
func (ps *PlayerService) HandleCharacterEnteredMap(player *Player) {
    // 发布角色进入地图事件
    event := NewEvent(EventTypeCharacterEnteredMap, &CharacterEnteredMapEvent{
        PlayerID: player.GetPlayerId(),
        MapID:    player.GetMapId(),
    })
    eventBus.Publish(event)
}
```

#### 3. Actor并发设计理念

**核心理念：基于消息传递的并发模型**

```go
// PlayerActor实现
type PlayerActor struct {
    *zActor.BaseActor
    Player *Player
}

// 创建PlayerActor
func NewPlayerActor(player *Player) *PlayerActor {
    pa := &PlayerActor{
        BaseActor: *zActor.NewBaseActor("player", player.GetPlayerId()),
        Player:    player,
    }

    // 启动消息处理循环
    go pa.Run()

    return pa
}

// 消息处理循环
func (pa *PlayerActor) Run() {
    for {
        select {
        case msg := <-pa.Mailbox:
            // 处理网络消息
            pa.handleMessage(msg)

        case <-time.After(time.Second / 10): // 100ms tick
            // 玩家主循环
            pa.update()
        }
    }
}
```

#### 4. ECS架构设计理念

**核心理念：实体-组件-系统分离**

```go
// Entity - 实体（唯一标识符）
type GameObject struct {
    *zObject.BaseObject
    name         string
    objectType   GameObjectType
    position     Vector3
    eventEmitter *zEvent.EventBus
    components   *component.ComponentManager
}

// Component - 组件（纯数据容器）
type BaseInfo struct {
    *component.BaseComponent
    name       string
    session    *zNet.TcpServerSession
    status     atomic.Int32
    exp        atomic.Int64
    gold       atomic.Int64
    level      atomic.Int32
    vipLevel   atomic.Int32
    serverId   int
    createTime int64
}

// System - 系统（行为逻辑处理）
type CombatSystem struct {
    attacks map[int64]*AttackRecord
    mu      sync.RWMutex
}

func (cs *CombatSystem) HandleAttack(attacker, target *Player) {
    // 计算伤害
    damage := cs.calculateDamage(attacker, target)

    // 处理伤害
    target.GetBaseInfo().SubHP(damage)

    // 发布战斗事件
    event := NewEvent(EventTypePlayerAttack, &PlayerAttackEvent{
        AttackerID: attacker.GetPlayerId(),
        TargetID:   target.GetPlayerId(),
        Damage:     damage,
    })
    eventBus.Publish(event)
}
```

---

## 🎨 核心架构设计模式

### 1. 服务架构模式 (Service Architecture)

```go
// 服务器核心
type GameServer struct {
    *zService.ServiceManager   // 服务管理器（继承）
    wg            sync.WaitGroup
    packetRouter  *router.PacketRouter
    protocol      protolayer.Protocol
    objectManager *zObject.ObjectManager
}

// 服务基类
type BaseService struct {
    zObject.BaseObject
    state ServiceState
    mu    sync.RWMutex
}

// 服务状态
type ServiceState int

const (
    ServiceStateCreated ServiceState = iota
    ServiceStateInit
    ServiceStateRunning
    ServiceStateStopping
    ServiceStateStopped
)
```

**核心特点**：

- **服务拓扑排序**：自动计算服务依赖关系，确保有序启动/关闭
- **服务状态管理**：Created → Init → Running → Stopping → Stopped
- **依赖注入 (DI)**：基于名称的依赖注入容器
- **并行启动**：每个服务在独立goroutine中运行

### 2. 事件驱动架构模式 (Event-Driven Architecture)

```go
// 事件总线
type EventBus struct {
    handlers map[EventType][]EventHandler
    mu       sync.RWMutex
    running  atomic.Bool
}

// 事件处理
func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    handlers, exists := eb.handlers[event.Type()]
    if !exists {
        return
    }

    // 异步处理事件
    for _, handler := range handlers {
        go handler(event)
    }
}
```

**核心功能**：

- **异步事件发布**：非阻塞式事件分发
- **事件订阅**：支持多订阅者监听同一事件
- **事件同步**：支持同步阻塞处理
- **事件监控**：事件处理统计和异常捕获

### 3. Actor模型模式 (Actor Model)

```go
// PlayerActor实现
type PlayerActor struct {
    *zActor.BaseActor
    Player *Player
}

// 创建PlayerActor
func NewPlayerActor(player *Player) *PlayerActor {
    pa := &PlayerActor{
        BaseActor: *zActor.NewBaseActor("player", player.GetPlayerId()),
        Player:    player,
    }

    // 启动消息处理循环
    go pa.Run()

    return pa
}

// 消息处理循环
func (pa *PlayerActor) Run() {
    for {
        select {
        case msg := <-pa.Mailbox:
            // 处理网络消息
            pa.handleMessage(msg)

        case <-time.After(time.Second / 10): // 100ms tick
            // 玩家主循环
            pa.update()
        }
    }
}

// 处理网络消息
func (pa *PlayerActor) handleMessage(msg *zActor.Message) {
    switch msg.Type {
    case protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN:
        pa.handleLogin(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_ATTACK:
        pa.handleAttack(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_MOVE:
        pa.handleMove(msg)

    default:
        // 处理其他消息
    }
}
```

**核心特点**：

- **消息驱动**：所有通信通过消息队列异步处理
- **并发隔离**：每个Actor拥有独立状态，避免竞态条件
- **全局系统**：统一管理所有Actor实例
- **类型安全**：强类型消息定义

### 4. ECS架构模式 (Entity-Component-System)

```go
// Entity - 实体（唯一标识符）
type GameObject struct {
    *zObject.BaseObject
    name         string
    objectType   GameObjectType
    position     Vector3
    eventEmitter *zEvent.EventBus
    components   *component.ComponentManager
}

// 组件访问
func (g *GameObject) GetComponent(name string) interface{} {
    return g.components.GetComponent(name)
}

func (g *GameObject) AddComponent(component common.IComponent) {
    g.components.AddComponent(component)
}

// System - 系统（行为逻辑处理）
type CombatSystem struct {
    attacks map[int64]*AttackRecord
    mu      sync.RWMutex
}

func (cs *CombatSystem) HandleAttack(attacker, target *Player) {
    // 获取攻击者基础信息
    attackerBaseInfo := attacker.GetComponent("baseinfo").(player.IBaseInfo)

    // 获取目标基础信息
    targetBaseInfo := target.GetComponent("baseinfo").(player.IBaseInfo)

    // 计算伤害
    damage := cs.calculateDamage(attacker, target)

    // 处理伤害
    targetBaseInfo.SubHP(damage)

    // 发布战斗事件
    event := NewEvent(EventTypePlayerAttack, &PlayerAttackEvent{
        AttackerID: attacker.GetPlayerId(),
        TargetID:   target.GetPlayerId(),
        Damage:     damage,
    })
    eventBus.Publish(event)
}
```

**ECS组成**：

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

### 5. 对象池设计模式 (Object Pool)

```go
// 对象池实现
type GenericPool struct {
    mu      sync.Mutex
    objects []interface{}
    newFunc func() interface{}
    maxSize int
}

// 创建对象池
func NewGenericPool(newFunc func() interface{}, maxSize int) *GenericPool {
    return &GenericPool{
        newFunc: newFunc,
        maxSize: maxSize,
        objects: make([]interface{}, 0, maxSize),
    }
}

// 获取对象
func (p *GenericPool) Get() interface{} {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.objects) > 0 {
        obj := p.objects[len(p.objects)-1]
        p.objects = p.objects[:len(p.objects)-1]
        return obj
    }

    return p.newFunc()
}

// 归还对象
func (p *GenericPool) Put(obj interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.objects) < p.maxSize {
        p.objects = append(p.objects, obj)
    }
    // 否则丢弃对象
}
```

**应用场景**：

- **技能对象池**：技能频繁创建/销毁
- **Buff对象池**：Buff效果管理
- **Actor对象池**：PlayerActor复用

---

## 🏗️ 模块架构详解

### 1. 玩家系统模块

#### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                       Player                            │
├─────────────────────────────────────────────────────────┤
│  playerId  session  object.LivingObject                 │
├─────────────────────────────────────────────────────────┤
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ BaseInfo │ │ Inventory │ │ Equipment │ │ Mailbox │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐              │
│  │ TaskMgr  │ │ SkillMgr │ │ Position │              │
│  └──────────┘ └──────────┘ └──────────┘              │
├─────────────────────────────────────────────────────────┤
│  AddComponent()  GetComponent()  Update()              │
└─────────────────────────────────────────────────────────┘
         │
         └──> ┌─────────────────────────────────────────┐
              │               PlayerActor               │
              ├─────────────────────────────────────────┤
              │  Mailbox  Run()  handleMessage()         │
              │  update()  sendToClient()               │
              └─────────────────────────────────────────┘
```

#### 设计理念

**核心理念：组件化、消息驱动、主循环**

1. **组件化设计**

   - 使用组件系统组织玩家数据
   - 支持动态添加/移除组件
   - 高内聚、低耦合的设计
2. **消息驱动**

   - 网络消息通过消息队列异步处理
   - 主循环处理玩家逻辑
   - 事件机制处理模块间通信
3. **主循环设计**

   - 固定时间间隔更新（100ms）
   - 处理网络消息
   - 处理玩家逻辑更新

#### 实现方式

```go
// Player结构体
type Player struct {
    *object.LivingObject
    playerId int64
    session  *zNet.TcpServerSession
}

// 获取组件
func (p *Player) GetComponent(name string) interface{} {
    return p.LivingObject.GetComponent(name)
}

// 获取等级
func (p *Player) GetLevel() int {
    baseInfo := p.GetComponent("baseinfo")
    if baseInfo == nil {
        return 1
    }
    return baseInfo.(player.IBaseInfo).GetLevel()
}

// 获取经验
func (p *Player) GetExp() int64 {
    baseInfo := p.GetComponent("baseinfo")
    if baseInfo == nil {
        return 0
    }
    return baseInfo.(player.IBaseInfo).GetExp()
}

// PlayerActor实现
type PlayerActor struct {
    *zActor.BaseActor
    Player *Player
}

// 消息处理循环
func (pa *PlayerActor) Run() {
    for {
        select {
        case msg := <-pa.Mailbox:
            // 处理网络消息
            pa.handleMessage(msg)

        case <-time.After(time.Second / 10): // 100ms tick
            // 玩家主循环
            pa.update()
        }
    }
}

// 处理网络消息
func (pa *PlayerActor) handleMessage(msg *zActor.Message) {
    switch msg.Type {
    case protocol.PlayerMsgId_MSG_PLAYER_ATTACK:
        pa.handleAttack(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_MOVE:
        pa.handleMove(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_SKILL:
        pa.handleSkill(msg)

    default:
        // 处理其他消息
    }
}

// 玩家主循环
func (pa *PlayerActor) update() {
    // 更新玩家组件
    pa.Player.Update(0.1) // 100ms

    // 处理玩家状态
    pa.handlePlayerState()

    // 同步玩家数据到客户端
    pa.syncPlayerData()
}
```

### 2. 战斗系统模块

#### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                     CombatSystem                       │
├─────────────────────────────────────────────────────────┤
│  attacks: map[int64]*AttackRecord                      │
│  mu: sync.RWMutex                                       │
├─────────────────────────────────────────────────────────┤
│  HandleAttack()  calculateDamage()  handleDamage()       │
│  handleHit()     handleCritical()  handleMiss()         │
│  handleBlock()   handleParry()     handleDodge()        │
└─────────────────────────────────────────────────────────┘
         │
         └──> ┌─────────────────────────────────────────┐
              │              AttackRecord               │
              ├─────────────────────────────────────────┤
              │  attackerID  targetID  damage            │
              │  hitType  critical  timestamp           │
              └─────────────────────────────────────────┘
         │
         └──> ┌─────────────────────────────────────────┐
              │                BuffSystem               │
              ├─────────────────────────────────────────┤
              │  buffs: map[int64][]*Buff                │
              │  ApplyBuff()  RemoveBuff()  UpdateBuffs()│
              └─────────────────────────────────────────┘
```

#### 设计理念

**核心理念：事件驱动、状态机、结果回调**

1. **事件驱动**

   - 战斗事件通过事件总线分发
   - 支持战斗前、战斗中、战斗后的事件处理
   - 模块化的战斗逻辑
2. **状态机设计**

   - 攻击状态、防御状态、暴击状态、闪避状态
   - 状态转换清晰
   - 便于扩展新状态
3. **结果回调**

   - 战斗结果通过回调通知
   - 支持异步战斗结果处理
   - 便于扩展战斗结果处理

#### 实现方式

```go
// CombatSystem结构体
type CombatSystem struct {
    attacks map[int64]*AttackRecord
    mu      sync.RWMutex
}

// 处理攻击事件
func (cs *CombatSystem) HandleAttack(attacker, target *Player) {
    // 获取攻击者基础信息
    attackerBaseInfo := attacker.GetComponent("baseinfo").(player.IBaseInfo)

    // 获取目标基础信息
    targetBaseInfo := target.GetComponent("baseinfo").(player.IBaseInfo)

    // 计算伤害
    damage := cs.calculateDamage(attacker, target)

    // 处理伤害
    targetBaseInfo.SubHP(damage)

    // 记录攻击记录
    attackRecord := &AttackRecord{
        attackerID: attacker.GetPlayerId(),
        targetID:   target.GetPlayerId(),
        damage:    damage,
        timestamp:  time.Now().UnixMilli(),
    }
    cs.attacks[attackRecord.attackerID] = attackRecord

    // 发布战斗事件
    event := NewEvent(EventTypePlayerAttack, &PlayerAttackEvent{
        AttackerID: attacker.GetPlayerId(),
        TargetID:   target.GetPlayerId(),
        Damage:     damage,
    })
    eventBus.Publish(event)

    // 发送攻击结果到客户端
    attackerActor := player.GetPlayerActor(attacker.GetPlayerId())
    attackerActor.sendToClient(&protocol.PlayerAttackResponse{
        Success: true,
        TargetID: target.GetPlayerId(),
        Damage:   damage,
    })
}

// 计算伤害
func (cs *CombatSystem) calculateDamage(attacker, target *Player) int32 {
    // 获取攻击者攻击力
    attack := attacker.GetComponent("property").(player.IProperty).GetAttack()

    // 获取目标防御力
    defense := target.GetComponent("property").(player.IProperty).GetDefense()

    // 计算基础伤害
    damage := attack - defense
    if damage < 1 {
        damage = 1
    }

    // 计算暴击
    if cs.isCritical(attacker) {
        damage *= 2
    }

    return damage
}

// 暴击判定
func (cs *CombatSystem) isCritical(attacker *Player) bool {
    // 获取暴击率
    critRate := attacker.GetComponent("property").(player.IProperty).GetCritRate()

    // 随机判定
    if rand.Float32() < critRate {
        return true
    }

    return false
}
```

### 3. 公会系统模块

#### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                      GuildSystem                        │
├─────────────────────────────────────────────────────────┤
│  guilds: map[int64]*Guild                              │
│  mu: sync.RWMutex                                       │
├─────────────────────────────────────────────────────────┤
│  CreateGuild()  DestroyGuild()  GetGuild()             │
│  JoinGuild()    LeaveGuild()    KickFromGuild()        │
│  UpgradeGuild()  DonateGuild()  ApplyGuild()           │
│  HandleGuildWar()  HandleGuildEvent()                   │
└─────────────────────────────────────────────────────────┘
         │
         └──> ┌─────────────────────────────────────────┐
              │                  Guild                  │
              ├─────────────────────────────────────────┤
              │  guildID  name  level  exp  money       │
              │  leaderID  members: map[int64]*GuildMember│
              │  applications: map[int64]*GuildApplication│
              ├─────────────────────────────────────────┤
              │  GetLevel()  GetExp()  AddExp()         │
              │  GetMoney()  AddMoney()  KickMember()    │
              │  AcceptApplication()  RejectApplication()│
              └─────────────────────────────────────────┘
         │
         └──> ┌─────────────────────────────────────────┐
              │              GuildMember                 │
              ├─────────────────────────────────────────┤
              │  playerID  role  joinTime  contribution  │
              │  isOnline                                 │
              └─────────────────────────────────────────┘
```

#### 设计理念

**核心理念：模块化、事件驱动、权限控制**

1. **模块化设计**

   - 公会系统独立于玩家系统
   - 支持公会创建、升级、解散等操作
   - 高内聚、低耦合的设计
2. **事件驱动**

   - 公会事件通过事件总线分发
   - 支持公会创建、升级、解散等事件处理
   - 事件机制处理模块间通信
3. **权限控制**

   - 公会成员权限控制
   - 公会管理权限控制
   - 灵活的权限配置

#### 实现方式

```go
// GuildSystem结构体
type GuildSystem struct {
    guilds map[int64]*Guild
    mu     sync.RWMutex
}

// 创建公会
func (gs *GuildSystem) CreateGuild(leader *Player, name string) (*Guild, error) {
    // 检查公会名称
    if gs.isGuildNameExist(name) {
        return nil, errors.New("公会名称已存在")
    }

    // 创建公会
    guild := &Guild{
        guildID:   generateGuildID(),
        name:      name,
        level:     1,
        exp:       0,
        money:     0,
        leaderID:  leader.GetPlayerId(),
        members:   make(map[int64]*GuildMember),
        applications: make(map[int64]*GuildApplication),
    }

    // 添加会长
    guild.AddMember(leader, GuildRoleLeader)

    // 保存公会
    gs.mu.Lock()
    defer gs.mu.Unlock()
    gs.guilds[guild.guildID] = guild

    // 发布公会创建事件
    event := NewEvent(EventTypeGuildCreated, &GuildCreatedEvent{
        GuildID:   guild.guildID,
        GuildName: guild.name,
        LeaderID:  leader.GetPlayerId(),
    })
    eventBus.Publish(event)

    return guild, nil
}

// 加入公会
func (gs *GuildSystem) JoinGuild(guildID int64, player *Player) error {
    // 获取公会
    guild, exists := gs.GetGuild(guildID)
    if !exists {
        return errors.New("公会不存在")
    }

    // 检查玩家是否已加入公会
    if gs.GetPlayerGuildID(player.GetPlayerId()) != 0 {
        return errors.New("玩家已加入公会")
    }

    // 添加公会成员
    guild.AddMember(player, GuildRoleMember)

    // 发布公会加入事件
    event := NewEvent(EventTypeGuildJoined, &GuildJoinedEvent{
        GuildID:    guildID,
        PlayerID:   player.GetPlayerId(),
    })
    eventBus.Publish(event)

    return nil
}

// 离开公会
func (gs *GuildSystem) LeaveGuild(guildID int64, player *Player) error {
    // 获取公会
    guild, exists := gs.GetGuild(guildID)
    if !exists {
        return errors.New("公会不存在")
    }

    // 检查玩家是否在公会中
    if guild.GetPlayerGuildRole(player.GetPlayerId()) == GuildRoleNone {
        return errors.New("玩家不在公会中")
    }

    // 检查是否是会长
    if guild.leaderID == player.GetPlayerId() {
        return errors.New("会长不能离开公会，只能解散公会")
    }

    // 移除公会成员
    guild.RemoveMember(player.GetPlayerId())

    // 发布公会离开事件
    event := NewEvent(EventTypeGuildLeft, &GuildLeftEvent{
        GuildID:    guildID,
        PlayerID:   player.GetPlayerId(),
    })
    eventBus.Publish(event)

    return nil
}
```

---

## 🔧 实现方式详解

### 1. 网络模块实现

#### 网络服务实现

```go
// 网络服务结构体
type TcpService struct {
    server     *zNet.TcpServer
    logger      zLog.Logger
    handler     NetHandler
    stopChan    chan struct{}
    wg          sync.WaitGroup
}

// 创建网络服务
func NewTcpService(config *zNet.TcpConfig, logger zLog.Logger, handler NetHandler) *TcpService {
    return &TcpService{
        logger:      logger,
        handler:     handler,
        stopChan:    make(chan struct{}),
    }
}

// 启动网络服务
func (s *TcpService) Start() error {
    // 创建TCP服务器
    server := zNet.NewTcpServer(s.config,
        zNet.WithLogger(s.logger),
        zNet.WithWorkerPoolSize(100),
    )

    // 注册消息处理器
    server.RegisterDispatcher(func(session interface{}, packet *zNet.NetPacket) error {
        return s.handler(session, packet)
    }, 100)

    // 启动服务器
    if err := server.Start(); err != nil {
        return err
    }

    s.server = server
    return nil
}

// 停止网络服务
func (s *TcpService) Stop() error {
    if s.server != nil {
        if err := s.server.Stop(); err != nil {
            return err
        }
    }

    return nil
}
```

#### 消息处理器实现

```go
// 网络消息处理器
type NetHandler func(session interface{}, packet *zNet.NetPacket) error

// 玩家网络消息处理器
func PlayerNetHandler(session interface{}, packet *zNet.NetPacket) error {
    // 解析消息
    msgId := packet.MessageID
    data := packet.Data

    // 处理消息
    switch msgId {
    case protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN:
        return handleLogin(session, data)

    case protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGOUT:
        return handleLogout(session, data)

    case protocol.PlayerMsgId_MSG_PLAYER_ATTACK:
        return handleAttack(session, data)

    case protocol.PlayerMsgId_MSG_PLAYER_MOVE:
        return handleMove(session, data)

    case protocol.PlayerMsgId_MSG_PLAYER_SKILL:
        return handleSkill(session, data)

    default:
        return errors.New("未知消息ID: " + strconv.Itoa(int(msgId)))
    }
}

// 处理玩家登录
func handleLogin(session interface{}, data []byte) error {
    // 解析登录请求
    var req protocol.PlayerAccountLoginRequest
    if err := proto.Unmarshal(data, &req); err != nil {
        return err
    }

    // 创建玩家对象
    player, err := player.NewPlayer(req.AccountID, req.PlayerID)
    if err != nil {
        return err
    }

    // 关联会话
    player.SetSession(session.(*zNet.TcpServerSession))

    // 初始化玩家组件
    player.InitComponents()

    // 创建玩家Actor
    actor := player.NewPlayerActor(player)
    player.SetActor(actor)

    // 存储玩家
    playerStorage.StorePlayer(player)

    // 发送登录成功
    resp := &protocol.PlayerAccountLoginResponse{
        Success: true,
        AccountInfo: &protocol.PlayerAccountInfo{
            AccountID: req.AccountID,
            PlayerID:  req.PlayerID,
        },
    }
    return session.(*zNet.TcpServerSession).Send(resp)
}

// 处理玩家攻击
func handleAttack(session interface{}, data []byte) error {
    // 解析攻击请求
    var req protocol.PlayerAttackRequest
    if err := proto.Unmarshal(data, &req); err != nil {
        return err
    }

    // 获取玩家Actor
    playerActor, exists := player.GetPlayerActor(req.PlayerID)
    if !exists {
        return errors.New("玩家不存在")
    }

    // 发送攻击消息到玩家Actor
    msg := actor.NewMessage(protocol.PlayerMsgId_MSG_PLAYER_ATTACK, data)
    return playerActor.Send(msg)
}
```

### 2. 玩家Actor实现

```go
// PlayerActor结构体
type PlayerActor struct {
    *zActor.BaseActor
    Player *Player
}

// 创建PlayerActor
func NewPlayerActor(player *Player) *PlayerActor {
    pa := &PlayerActor{
        BaseActor: *zActor.NewBaseActor("player", player.GetPlayerId()),
        Player:    player,
    }

    // 启动消息处理循环
    go pa.Run()

    return pa
}

// 消息处理循环
func (pa *PlayerActor) Run() {
    for {
        select {
        case msg := <-pa.Mailbox:
            // 处理网络消息
            pa.handleMessage(msg)

        case <-time.After(time.Second / 10): // 100ms tick
            // 玩家主循环
            pa.update()
        }
    }
}

// 处理网络消息
func (pa *PlayerActor) handleMessage(msg *zActor.Message) {
    switch msg.Type {
    case protocol.PlayerMsgId_MSG_PLAYER_ATTACK:
        pa.handleAttack(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_MOVE:
        pa.handleMove(msg)

    case protocol.PlayerMsgId_MSG_PLAYER_SKILL:
        pa.handleSkill(msg)

    default:
        // 处理其他消息
    }
}

// 处理玩家攻击
func (pa *PlayerActor) handleAttack(msg *zActor.Message) {
    // 解析攻击请求
    var req protocol.PlayerAttackRequest
    if err := proto.Unmarshal(msg.Data, &req); err != nil {
        pa.logger.Error("解析攻击请求失败", zap.Error(err))
        return
    }

    // 获取目标玩家
    targetPlayer, exists := player.GetPlayer(req.TargetID)
    if !exists {
        pa.sendToClient(&protocol.PlayerAttackResponse{
            Success: false,
            ErrorMsg: "目标不存在",
        })
        return
    }

    // 处理攻击
    combatSystem.HandleAttack(pa.Player, targetPlayer)
}

// 处理玩家移动
func (pa *PlayerActor) handleMove(msg *zActor.Message) {
    // 解析移动请求
    var req protocol.PlayerMoveRequest
    if err := proto.Unmarshal(msg.Data, &req); err != nil {
        pa.logger.Error("解析移动请求失败", zap.Error(err))
        return
    }

    // 更新玩家位置
    pa.Player.SetPosition(req.PositionX, req.PositionY, req.PositionZ)

    // 发送移动结果
    pa.sendToClient(&protocol.PlayerMoveResponse{
        Success: true,
    })
}

// 玩家主循环
func (pa *PlayerActor) update() {
    // 更新玩家组件
    pa.Player.Update(0.1) // 100ms

    // 处理玩家状态
    pa.handlePlayerState()

    // 同步玩家数据到客户端
    pa.syncPlayerData()
}

// 同步玩家数据
func (pa *PlayerActor) syncPlayerData() {
    // 构建玩家数据
    playerData := &protocol.PlayerData{
        PlayerID:  pa.Player.GetPlayerId(),
        Level:     int32(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetLevel()),
        Exp:       int64(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetExp()),
        Gold:      int64(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetGold()),
        PositionX: pa.Player.GetComponent("position").(player.IPosition).GetX(),
        PositionY: pa.Player.GetComponent("position").(player.IPosition).GetY(),
        PositionZ: pa.Player.GetComponent("position").(player.IPosition).GetZ(),
    }

    // 发送到客户端
    pa.sendToClient(&protocol.PlayerSyncResponse{
        Success:  true,
        PlayerData: playerData,
    })
}

// 发送消息到客户端
func (pa *PlayerActor) sendToClient(resp interface{}) error {
    if pa.Player.GetSession() == nil {
        return errors.New("会话不存在")
    }

    return pa.Player.GetSession().Send(resp)
}
```

### 3. 数据库模块实现

#### 数据库管理器实现

```go
// 数据库管理器
type DBManager struct {
    accountDB *db.MongoConnector
    gameDB    *db.MongoConnector
    logDB     *db.MongoConnector
    mu        sync.RWMutex
}

// 创建数据库管理器
func NewDBManager(accountConfig *db.MongoConfig, gameConfig *db.MongoConfig, logConfig *db.MongoConfig) *DBManager {
    return &DBManager{
        accountDB: db.NewMongoConnector(accountConfig),
        gameDB:    db.NewMongoConnector(gameConfig),
        logDB:     db.NewMongoConnector(logConfig),
    }
}

// 初始化数据库
func (m *DBManager) Init() error {
    // 初始化账号数据库
    if err := m.accountDB.Init(); err != nil {
        return err
    }

    // 初始化游戏数据库
    if err := m.gameDB.Init(); err != nil {
        return err
    }

    // 初始化日志数据库
    if err := m.logDB.Init(); err != nil {
        return err
    }

    return nil
}

// 关闭数据库
func (m *DBManager) Close() error {
    if err := m.accountDB.Close(); err != nil {
        return err
    }

    if err := m.gameDB.Close(); err != nil {
        return err
    }

    if err := m.logDB.Close(); err != nil {
        return err
    }

    return nil
}

// 获取账号数据库
func (m *DBManager) GetAccountDB() *db.MongoConnector {
    return m.accountDB
}

// 获取游戏数据库
func (m *DBManager) GetGameDB() *db.MongoConnector {
    return m.gameDB
}

// 获取日志数据库
func (m *DBManager) GetLogDB() *db.MongoConnector {
    return m.logDB
}
```

#### 账号仓库实现

```go
// 账号仓库
type AccountRepository struct {
    db *db.MongoConnector
}

// 创建账号仓库
func NewAccountRepository(db *db.MongoConnector) *AccountRepository {
    return &AccountRepository{
        db: db,
    }
}

// 创建账号
func (r *AccountRepository) CreateAccount(account *models.Account) error {
    collection := r.db.GetDB().Collection("accounts")

    // 插入数据
    _, err := collection.InsertOne(context.Background(), account)
    if err != nil {
        return err
    }

    return nil
}

// 查询账号
func (r *AccountRepository) GetAccount(accountID string) (*models.Account, error) {
    collection := r.db.GetDB().Collection("accounts")

    // 查询数据
    var account models.Account
    err := collection.FindOne(context.Background(), bson.M{"account_id": accountID}).Decode(&account)
    if err != nil {
        return nil, err
    }

    return &account, nil
}

// 更新账号
func (r *AccountRepository) UpdateAccount(account *models.Account) error {
    collection := r.db.GetDB().Collection("accounts")

    // 更新数据
    _, err := collection.UpdateOne(
        context.Background(),
        bson.M{"account_id": account.AccountID},
        bson.M{"$set": bson.M{
            "password": account.Password,
            "status":   account.Status,
        }},
    )
    if err != nil {
        return err
    }

    return nil
}

// 删除账号
func (r *AccountRepository) DeleteAccount(accountID string) error {
    collection := r.db.GetDB().Collection("accounts")

    // 删除数据
    _, err := collection.DeleteOne(context.Background(), bson.M{"account_id": accountID})
    if err != nil {
        return err
    }

    return nil
}

// 异步查询账号
func (r *AccountRepository) GetAccountAsync(accountID string, callback func(*models.Account, error)) {
    go func() {
        account, err := r.GetAccount(accountID)
        callback(account, err)
    }()
}
```

---

## 📚 新手使用指引

### 快速开始

#### 步骤1：环境准备

```bash
# 安装Go环境
# 下载并安装：https://golang.org/dl/
go version  # 验证安装

# 配置GOPATH
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin

# 安装依赖管理工具
go install github.com/golang/dlv/cmd/dlv@latest
go install github.com/cosmtrek/air@latest
```

#### 步骤2：克隆项目

```bash
git clone https://github.com/pzqf/zGameServer.git
cd zGameServer

# 安装依赖
go mod tidy
```

#### 步骤3：配置项目

```bash
# 配置数据库
# 打开 config.ini 文件
# 配置数据库连接信息

# 配置Excel表格
# 确保 resources/excel_tables/ 目录下有所有必要的Excel配置表

# 运行数据库迁移
# 根据需要创建数据库表结构
```

#### 步骤4：运行项目

```bash
# 编译项目
go build -o zGameServer.exe .

# 运行服务器
./zGameServer.exe

# 运行测试客户端
go run client/testclient.go
```

### 模块使用示例

#### 1. 玩家系统使用

```go
// 创建玩家
player, err := player.NewPlayer(accountID, playerID)
if err != nil {
    return err
}

// 初始化玩家组件
player.InitComponents()

// 创建玩家Actor
actor := player.NewPlayerActor(player)
player.SetActor(actor)

// 存储玩家
playerStorage.StorePlayer(player)

// 获取玩家组件
baseInfo := player.GetComponent("baseinfo").(player.IBaseInfo)
inventory := player.GetComponent("inventory").(player.IInventory)
equipment := player.GetComponent("equipment").(player.IEquipment)

// 使用组件
level := baseInfo.GetLevel()
exp := baseInfo.GetExp()
gold := baseInfo.GetGold()

// 发送消息到玩家Actor
msg := actor.NewMessage(protocol.PlayerMsgId_MSG_PLAYER_ATTACK, data)
actor.Send(msg)
```

#### 2. 战斗系统使用

```go
// 创建战斗系统
combatSystem := NewCombatSystem()

// 处理攻击
combatSystem.HandleAttack(attackerPlayer, targetPlayer)

// 计算伤害
damage := combatSystem.CalculateDamage(attackerPlayer, targetPlayer)

// 暴击判定
if combatSystem.IsCritical(attackerPlayer) {
    damage *= 2
}

// 处理伤害
targetPlayer.GetComponent("baseinfo").(player.IBaseInfo).SubHP(damage)

// 发布战斗事件
event := NewEvent(EventTypePlayerAttack, &PlayerAttackEvent{
    AttackerID: attackerPlayer.GetPlayerId(),
    TargetID:   targetPlayer.GetPlayerId(),
    Damage:     damage,
})
eventBus.Publish(event)
```

#### 3. 公会系统使用

```go
// 创建公会系统
guildSystem := NewGuildSystem()

// 创建公会
guild, err := guildSystem.CreateGuild(leaderPlayer, guildName)
if err != nil {
    return err
}

// 加入公会
err := guildSystem.JoinGuild(guildID, player)
if err != nil {
    return err
}

// 离开公会
err := guildSystem.LeaveGuild(guildID, player)
if err != nil {
    return err
}

// 升级公会
err := guildSystem.UpgradeGuild(guildID)
if err != nil {
    return err
}

// 获取公会信息
guildInfo, err := guildSystem.GetGuildInfo(guildID)
if err != nil {
    return err
}
```

### 开发指南

#### 1. 代码规范

```go
// 结构体命名：大驼峰式
type Player struct {
    // 字段命名：小驼峰式（包内可见）
    playerId int64
    // 私有字段：小驼峰式
    session *zNet.TcpServerSession
}

// 函数命名：大驼峰式
func NewPlayer(accountID, playerID string) *Player {
    return &Player{
        accountID: accountID,
        playerID:  playerID,
    }
}

// 方法命名：大驼峰式
func (p *Player) GetPlayerId() string {
    return p.playerID
}

// 接口命名：大驼峰式，以I开头
type IBaseInfo interface {
    GetLevel() int
    GetExp() int64
    GetGold() int64
}

// 常量命名：大驼峰式
const (
    PlayerStatusOffline = 0
    PlayerStatusOnline  = 1
)

// 变量命名：小驼峰式
var playerCount int

// 全局变量：首字母大写，公开
var PlayerStorage *PlayerStorage

// 私有变量：首字母小写，不公开
var playerMap map[int64]*Player
```

#### 2. 错误处理

```go
// 返回错误，而不是panic
func DoSomething() error {
    // 检查错误
    if err != nil {
        return err
    }

    // 处理错误
    return nil
}

// 调用错误处理
if err := DoSomething(); err != nil {
    // 记录错误
    logger.Error("Failed to do something", zap.Error(err))
    // 返回错误
    return err
}

// 使用错误链
if err := DoSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

#### 3. 并发控制

```go
// 互斥锁
var mu sync.Mutex

func (s *Service) Process() {
    mu.Lock()
    defer mu.Unlock()

    // 临界区代码
}

// 读写锁
var rwmu sync.RWMutex

func (s *Service) Get() {
    rwmu.RLock()
    defer rwmu.RUnlock()

    // 读操作
}

func (s *Service) Set() {
    rwmu.Lock()
    defer rwmu.Unlock()

    // 写操作
}

// 原子操作
var counter atomic.Int32

func (s *Service) Increment() {
    counter.Add(1)
}

// 通道通信
type Worker struct {
    jobChan chan Job
    stopChan chan struct{}
    wg sync.WaitGroup
}

func (w *Worker) Loop() {
    for {
        select {
        case job := <-w.jobChan:
            // 处理任务
        case <-w.stopChan:
            return
        }
    }
}
```

---

## 🎯 最佳实践

### 1. 性能优化

#### 网络性能优化

1. **连接复用** - 使用连接池，减少连接创建销毁开销
2. **批量处理** - 批量发送/接收数据，减少系统调用
3. **零拷贝** - 避免不必要的数据复制
4. **异步I/O** - 使用非阻塞I/O，提高并发处理能力
5. **内存池** - 减少内存分配和GC压力

```go
// 连接池实现
type ConnectionPool struct {
    pool  chan *Connection
    newFunc func() *Connection
    mu    sync.Mutex
}

func (cp *ConnectionPool) Get() *Connection {
    select {
    case conn := <-cp.pool:
        return conn
    default:
        return cp.newFunc()
    }
}

func (cp *ConnectionPool) Put(conn *Connection) {
    select {
    case cp.pool <- conn:
    default:
        // 池已满，关闭连接
        conn.Close()
    }
}
```

#### 数据库性能优化

1. **连接池** - 使用连接池管理数据库连接
2. **批量操作** - 批量插入、更新数据
3. **查询优化** - 使用索引，避免全表扫描
4. **缓存策略** - 缓存热点数据
5. **异步查询** - 使用异步查询，避免阻塞

```go
// 数据库连接池
type DBConnectionPool struct {
    pool  chan *sql.DB
    newFunc func() *sql.DB
    mu    sync.Mutex
}

func (cp *DBConnectionPool) Get() *sql.DB {
    select {
    case db := <-cp.pool:
        return db
    default:
        return cp.newFunc()
    }
}

func (cp *DBConnectionPool) Put(db *sql.DB) {
    select {
    case cp.pool <- db:
    default:
        // 池已满，关闭连接
        db.Close()
    }
}

// 异步查询
func (r *AccountRepository) GetAccountAsync(accountID string, callback func(*models.Account, error)) {
    go func() {
        account, err := r.GetAccount(accountID)
        callback(account, err)
    }()
}
```

#### 玩家性能优化

1. **对象池** - 使用对象池管理玩家对象
2. **组件预创建** - 预创建玩家组件
3. **批量同步** - 批量同步玩家数据
4. **状态压缩** - 压缩玩家状态数据
5. **事件合并** - 合并玩家事件

```go
// 玩家对象池
type PlayerPool struct {
    pool  chan *Player
    newFunc func() *Player
    mu    sync.Mutex
}

func (pp *PlayerPool) Get() *Player {
    select {
    case player := <-pp.pool:
        return player
    default:
        return pp.newFunc()
    }
}

func (pp *PlayerPool) Put(player *Player) {
    select {
    case pp.pool <- player:
    default:
        // 池已满，丢弃
    }
}

// 批量同步玩家数据
func (pa *PlayerActor) syncPlayerData() {
    // 构建玩家数据
    playerData := &protocol.PlayerData{
        PlayerID:  pa.Player.GetPlayerId(),
        Level:     int32(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetLevel()),
        Exp:       int64(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetExp()),
        Gold:      int64(pa.Player.GetComponent("baseinfo").(player.IBaseInfo).GetGold()),
        PositionX: Player.GetComponent("position").(player.IPosition).GetX(),
        PositionY: Player.GetComponent("position").(player.IPosition).GetY(),
        PositionZ: Player.GetComponent("position").(player.IPosition).GetZ(),
    }

    // 发送到客户端
    pa.sendToClient(&protocol.PlayerSyncResponse{
        Success:  true,
        PlayerData: playerData,
    })
}
```

### 2. 安全性最佳实践

#### 网络安全

1. **连接限制** - 限制每个IP的连接数
2. **流量限制** - 限制每个连接的流量
3. **数据包验证** - 验证数据包格式和内容
4. **加密通信** - 使用TLS加密数据传输

```go
// DDoS防护
type DdosProtector struct {
    ipMap map[string]*IpInfo
    mu    sync.RWMutex
}

func (dp *DdosProtector) Allow(ip string) bool {
    dp.mu.RLock()
    info, exists := dp.ipMap[ip]
    dp.mu.RUnlock()

    if !exists {
        return true
    }

    // 检查连接数
    if info.ConnCount > 100 {
        return false
    }

    // 检查流量
    if info.Traffic > 100*1024*1024 {
        return false
    }

    return true
}
```

#### 代码安全

1. **输入验证** - 验证所有外部输入
2. **SQL注入防护** - 使用参数化查询
3. **XSS防护** - 过滤用户输入
4. **CSRF防护** - 使用Token验证

```go
// 输入验证
func ValidateInput(input string) error {
    if input == "" {
        return errors.New("输入不能为空")
    }

    if len(input) > 100 {
        return errors.New("输入过长")
    }

    // 正则表达式验证
    match := regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(input)
    if !match {
        return errors.New("输入包含非法字符")
    }

    return nil
}

// 参数化查询
func (r *AccountRepository) GetAccountAsync(accountID string, callback func(*models.Account, error)) {
    go func() {
        var account models.Account
        err := r.db.QueryRow("SELECT * FROM accounts WHERE account_id = ?", accountID).
            Scan(&account.AccountID, &account.Password, &account.Status, &account.CreateTime)
        if err != nil {
            callback(nil, err)
            return
        }
        callback(&account, nil)
    }()
}
```

### 3. 可维护性最佳实践

#### 代码组织

1. **模块化** - 按功能划分模块
2. **单一职责** - 每个函数/类只做一件事
3. **接口抽象** - 使用接口，不依赖实现
4. **依赖注入** - 通过依赖注入实现解耦

```go
// 模块化示例
package game

import (
    "github.com/pzqf/zEngine/zActor"
    "github.com/pzqf/zEngine/zEvent"
)

// 战斗模块
type CombatModule struct {
    actor *zActor.Actor
    eventBus *zEvent.EventBus
}

// 移动模块
type MovementModule struct {
    actor *zActor.Actor
    eventBus *zEvent.EventBus
}

// 技能模块
type SkillModule struct {
    actor *zActor.Actor
    eventBus *zEvent.EventBus
}
```

#### 文档编写

1. **代码注释** - 注释复杂逻辑和关键算法
2. **API文档** - 使用godoc生成API文档
3. **README** - 项目功能和使用说明
4. **CHANGELOG** - 版本变更记录

```go
// 战斗系统实现
// 
// 战斗系统负责处理玩家和怪物之间的战斗逻辑，
// 包括攻击、防御、技能释放、状态效果等。
// 
// 核心功能：
// 1. 攻击计算 - 根据属性和技能计算伤害
// 2. 状态管理 - 管理BUFF、DEBUFF等状态
// 3. 仇恨系统 - 管理怪物的目标选择
// 4. 战斗结果 - 处理胜负和奖励
// 
// 使用示例：
//   combat := NewCombatSystem()
//   combat.Attack(attacker, target)
// 
type CombatSystem struct {
    // 技能管理器
    skillMgr *SkillManager
  
    // 状态管理器
    statusMgr *StatusManager
  
    // 仇恨系统
    aggroSystem *AggroSystem
}
```

### 4. 监控和调试

#### 性能监控

```go
// 性能监控
type PerformanceMonitor struct {
    counters map[string]atomic.Int64
    timers   map[string]*Timer
}

func (pm *PerformanceMonitor) Count(name string) {
    pm.counters[name].Add(1)
}

func (pm *PerformanceMonitor) Time(name string) *Timer {
    timer := NewTimer()
    pm.timers[name] = timer
    return timer
}

// 性能统计
type PerformanceStats struct {
    Requests    int64
    Errors      int64
    AvgLatency  time.Duration
    P95Latency  time.Duration
    P99Latency  time.Duration
}

func (pm *PerformanceMonitor) GetStats() *PerformanceStats {
    return &PerformanceStats{
        Requests:    pm.counters["requests"].Load(),
        Errors:      pm.counters["errors"].Load(),
        AvgLatency:  pm.calculateAvgLatency(),
        P95Latency:  pm.calculateP95Latency(),
        P99Latency:  pm.calculateP99Latency(),
    }
}
```

#### 日志监控

```go
// 日志监控
type LogMonitor struct {
    errorChan    chan *LogEntry
    warningChan  chan *LogEntry
    stats        map[string]int64
}

func (lm *LogMonitor) Monitor(logger *zLog.Logger) {
    // 订阅日志事件
    logger.Subscribe(func(entry *LogEntry) {
        switch entry.Level {
        case zap.ErrorLevel:
            lm.errorChan <- entry
        case zap.WarnLevel:
            lm.warningChan <- entry
        }
    })

    // 处理错误日志
    go func() {
        for entry := <-lm.errorChan {
            lm.stats["errors"]++
            // 发送告警
            lm.sendAlert(entry)
        }
    }()
}
```

---

## 📖 参考资源

### 官方文档

- [Go官方文档](https://golang.org/doc/)
- [zEngine项目文档](https://github.com/pzqf/zEngine)
- [zGameServer项目文档](https://github.com/pzqf/zGameServer)

### 学习资源

- **游戏服务器开发**：

  - 《游戏服务器架构设计》
  - 《多人在线游戏开发》
- **网络编程**：

  - 《Go网络编程》
  - 《高性能网络编程》
- **数据库**：

  - 《高性能MySQL》
  - 《MongoDB权威指南》

### 相关项目

- **游戏引擎**：

  - [zEngine](https://github.com/pzqf/zEngine) - 基于zEngine的游戏服务器框架
  - [zUtil](https://github.com/pzqf/zUtil) - 工具库
- **网络库**：

  - [zNet](https://github.com/pzqf/zEngine/tree/master/zNet) - 网络模块
  - [zEvent](https://github.com/pzqf/zEngine/tree/master/zEvent) - 事件模块
- **工具库**：

  - [zLog](https://github.com/pzqf/zEngine/tree/master/zLog) - 日志模块
  - [zActor](https://github.com/pzqf/zEngine/tree/master/zActor) - Actor模块

---

## 🤝 贡献指南

### 如何贡献

1. **Fork项目**
2. **创建分支**
3. **提交更改**
4. **推送分支**
5. **提交Pull Request**

### 代码规范

- 遵循Go代码规范
- 使用gofmt格式化代码
- 添加适当的注释
- 编写单元测试

### 提交规范

- 清晰的提交信息
- 参考conventional commits
- 关联相关Issue

---

## 📄 许可证

本项目采用MIT许可证。详情请参考LICENSE文件。

---

## 📞 联系方式

如有问题或建议，欢迎通过以下方式联系：

- **Issue**：https://github.com/pzqf/zGameServer/issues
- **邮件**：pzqf@example.com
- **Gitee**：https://gitee.com/pzqf/zGameServer

---

**zGameServer** - 高性能可扩展的MMO游戏服务器框架
