package handler

import (
	"crypto/sha256"
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/db"
	"github.com/pzqf/zGameServer/db/models"
	"github.com/pzqf/zGameServer/game/player"
	"github.com/pzqf/zGameServer/net/protocol"
	"github.com/pzqf/zGameServer/net/router"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// playerHandler 玩家请求处理器
type playerHandler struct {
	playerService *player.PlayerService
	dbManager     *db.DBManager
	// 会话管理，用于保存账号和角色的会话信息
	accountSession map[string]int64 // key: account, value: sessionId
	sessionAccount map[int64]string // key: sessionId, value: account
	charSession    map[int64]int64  // key: characterId, value: sessionId
	sessionChar    map[int64]int64  // key: sessionId, value: characterId
}

// NewPlayerHandler 创建玩家请求处理器
func NewPlayerHandler(playerService *player.PlayerService) *playerHandler {
	return &playerHandler{
		playerService:  playerService,
		accountSession: make(map[string]int64),
		sessionAccount: make(map[int64]string),
		charSession:    make(map[int64]int64),
		sessionChar:    make(map[int64]int64),
	}
}

// SetDBManager 设置数据库管理器
func (h *playerHandler) SetDBManager(dbManager *db.DBManager) {
	h.dbManager = dbManager
}

// RegisterPlayerHandlers 注册玩家相关的消息处理器
func RegisterPlayerHandlers(router *router.PacketRouter, playerService *player.PlayerService, dbManager *db.DBManager) {
	handler := NewPlayerHandler(playerService)
	handler.SetDBManager(dbManager)

	// 注册账号相关处理器
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), handler.handleAccountCreate)
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), handler.handleAccountLogin)
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), handler.handleCharacterCreate)
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), handler.handleCharacterLogin)
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGOUT), handler.handleCharacterLogout)

	// 注册玩家信息处理器
	router.RegisterHandler(int32(protocol.PlayerMsgId_MSG_PLAYER_GET_INFO), handler.handlePlayerGetInfo)
}

// handlePlayerGetInfo 处理获取玩家信息请求
func (h *playerHandler) handlePlayerGetInfo(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received get info request", zap.Int64("sessionId", int64(session.GetSid())))

	// 根据sessionId获取玩家
	player := h.playerService.GetPlayerBySession(int64(session.GetSid()))
	if player == nil {
		zLog.Warn("Player not found for session", zap.Int64("sessionId", int64(session.GetSid())))
		return nil
	}

	// 构建玩家信息响应
	info := protocol.PlayerBasicInfo{
		PlayerId:   player.GetPlayerId(),
		Name:       player.GetName(),
		Level:      int32(player.GetBasicInfo().Level),
		Exp:        player.GetBasicInfo().Exp.Load(),
		Gold:       player.GetBasicInfo().Gold.Load(),
		VipLevel:   int32(player.GetBasicInfo().VipLevel),
		ServerId:   1,
		CreateTime: player.GetBasicInfo().CreateTime,
	}

	// 发送响应
	respData, err := proto.Marshal(&info)
	if err != nil {
		zLog.Error("Failed to marshal player info response", zap.Error(err), zap.Int64("playerId", player.GetPlayerId()))
		return err
	}

	return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_GET_INFO), respData)
}

// validateLogin 验证登录信息（模拟）
func (h *playerHandler) validateLogin(account, password string) bool {
	// 简单的登录验证逻辑（模拟）
	// 实际应该连接数据库或认证服务进行验证
	// 这里允许任何非空的账号密码组合
	return account != "" && password != ""
}

// handleAccountCreate 处理账号创建请求
func (h *playerHandler) handleAccountCreate(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received account create request", zap.Int64("sessionId", int64(session.GetSid())))

	// 解析请求数据
	var req protocol.AccountCreateRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal account create request", zap.Error(err))
		return err
	}

	// 验证账号和密码
	if req.Account == "" || req.Password == "" {
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "账号或密码不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), respData)
	}

	// 检查账号是否已存在
	h.dbManager.AccountDAO.GetAccountByName(req.Account, func(account *models.Account, err error) {
		if err != nil {
			zLog.Error("Failed to check account existence", zap.Error(err))
			resp := protocol.AccountCreateResponse{
				Success:  false,
				ErrorMsg: "服务器错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), respData)
			return
		}

		if account != nil {
			// 账号已存在
			resp := protocol.AccountCreateResponse{
				Success:  false,
				ErrorMsg: "账号已存在",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), respData)
			return
		}

		// 创建新账号
		now := time.Now()
		newAccount := &models.Account{
			AccountID:   time.Now().UnixNano() / 1000000,
			AccountName: req.Account,
			Password:    req.Password, // 实际应该进行加密存储
			Status:      1,            // 1表示正常
			CreatedAt:   now,
			LastLoginAt: now,
		}

		// 保存账号到数据库
		h.dbManager.AccountDAO.CreateAccount(newAccount, func(accountID int64, err error) {
			if err != nil {
				zLog.Error("Failed to create account", zap.Error(err))
				resp := protocol.AccountCreateResponse{
					Success:  false,
					ErrorMsg: "服务器错误",
				}
				respData, _ := proto.Marshal(&resp)
				_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), respData)
				return
			}

			// 账号创建成功
			resp := protocol.AccountCreateResponse{
				Success:  true,
				ErrorMsg: "",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_CREATE), respData)
		})
	})

	return nil
}

// handleAccountLogin 处理账号登录请求
func (h *playerHandler) handleAccountLogin(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received account login request", zap.Int64("sessionId", int64(session.GetSid())))

	// 解析请求数据
	var req protocol.AccountLoginRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal account login request", zap.Error(err))
		return err
	}

	// 验证账号和密码
	if req.Account == "" || req.Password == "" {
		resp := protocol.AccountLoginResponse{
			Success:  false,
			ErrorMsg: "账号或密码不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), respData)
	}

	// 检查账号是否存在
	h.dbManager.AccountDAO.GetAccountByName(req.Account, func(account *models.Account, err error) {
		if err != nil {
			zLog.Error("Failed to get account", zap.Error(err))
			resp := protocol.AccountLoginResponse{
				Success:  false,
				ErrorMsg: "服务器错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), respData)
			return
		}

		if account == nil || account.Password != req.Password {
			// 账号不存在或密码错误
			resp := protocol.AccountLoginResponse{
				Success:  false,
				ErrorMsg: "账号或密码错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), respData)
			return
		}

		// 更新账号的最后登录时间
		account.LastLoginAt = time.Now()
		h.dbManager.AccountDAO.UpdateAccount(account, func(success bool, err error) {
			if err != nil {
				zLog.Error("Failed to update last login time", zap.Error(err))
			}
		})

		// 获取该账号下的所有角色
		h.dbManager.CharacterDAO.GetCharactersByAccountID(account.AccountID, func(characters []*models.Character, err error) {
			if err != nil {
				zLog.Error("Failed to get characters", zap.Error(err))
				resp := protocol.AccountLoginResponse{
					Success:  false,
					ErrorMsg: "服务器错误",
				}
				respData, _ := proto.Marshal(&resp)
				_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), respData)
				return
			}

			// 转换为协议格式的角色信息
			var characterInfos []*protocol.CharacterInfo
			for _, character := range characters {
				characterInfos = append(characterInfos, &protocol.CharacterInfo{
					CharacterId: character.CharID,
					Name:        character.CharName,
					Level:       int32(character.Level),
					Sex:         int32(character.Sex),
					Age:         int32(character.Age),
				})
			}

			// 保存账号和会话的对应关系
			h.accountSession[req.Account] = int64(session.GetSid())
			h.sessionAccount[int64(session.GetSid())] = req.Account

			// 发送登录成功响应，包含角色列表
			resp := protocol.AccountLoginResponse{
				Success:    true,
				ErrorMsg:   "",
				Characters: characterInfos,
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN), respData)
		})
	})

	return nil
}

// handleCharacterCreate 处理角色创建请求
func (h *playerHandler) handleCharacterCreate(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character create request", zap.Int64("sessionId", int64(session.GetSid())))

	// 获取账号信息
	account, ok := h.sessionAccount[int64(session.GetSid())]
	if !ok {
		// 未找到账号信息，需要重新登录
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
	}

	// 解析请求数据
	var req protocol.CharacterCreateRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal character create request", zap.Error(err))
		return err
	}

	// 验证角色名称
	if req.Name == "" {
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "角色名称不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
	}

	// 检查角色名称是否已存在
	h.dbManager.CharacterDAO.GetCharacterByName(req.Name, func(character *models.Character, err error) {
		if err != nil {
			zLog.Error("Failed to check character name", zap.Error(err))
			resp := protocol.CharacterCreateResponse{
				Success:  false,
				ErrorMsg: "服务器错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
			return
		}

		if character != nil {
			// 角色名称已存在
			resp := protocol.CharacterCreateResponse{
				Success:  false,
				ErrorMsg: "角色名称已存在",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
			return
		}

		// 获取账号ID
		h.dbManager.AccountDAO.GetAccountByName(account, func(account *models.Account, err error) {
			if err != nil || account == nil {
				zLog.Error("Failed to get account", zap.Error(err))
				resp := protocol.CharacterCreateResponse{
					Success:  false,
					ErrorMsg: "服务器错误",
				}
				respData, _ := proto.Marshal(&resp)
				_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
				return
			}

			// 创建新角色
			now := time.Now()
			newCharacter := &models.Character{
				CharID:    time.Now().UnixNano() / 1000000,
				AccountID: account.AccountID,
				CharName:  req.Name,
				Sex:       int(req.Sex),
				Age:       int(req.Age),
				Level:     1,
				CreatedAt: now,
				UpdatedAt: now,
			}

			// 保存角色到数据库
			h.dbManager.CharacterDAO.CreateCharacter(newCharacter, func(characterID int64, err error) {
				if err != nil {
					zLog.Error("Failed to create character", zap.Error(err))
					resp := protocol.CharacterCreateResponse{
						Success:  false,
						ErrorMsg: "服务器错误",
					}
					respData, _ := proto.Marshal(&resp)
					_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
					return
				}

				// 发送角色创建成功响应
				resp := protocol.CharacterCreateResponse{
					Success:  true,
					ErrorMsg: "",
					Character: &protocol.CharacterInfo{
						CharacterId: newCharacter.CharID,
						Name:        newCharacter.CharName,
						Level:       int32(newCharacter.Level),
						Sex:         int32(newCharacter.Sex),
						Age:         int32(newCharacter.Age),
					},
				}
				respData, _ := proto.Marshal(&resp)
				_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE), respData)
			})
		})
	})

	return nil
}

// handleCharacterLogin 处理角色选择登录请求
func (h *playerHandler) handleCharacterLogin(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character login request", zap.Int64("sessionId", int64(session.GetSid())))

	// 获取账号信息
	_, ok := h.sessionAccount[int64(session.GetSid())]
	if !ok {
		// 未找到账号信息，需要重新登录
		resp := protocol.CharacterLoginResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), respData)
	}

	// 解析请求数据
	var req protocol.CharacterLoginRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal character login request", zap.Error(err))
		return err
	}

	// 获取角色信息
	h.dbManager.CharacterDAO.GetCharacterByID(req.CharacterId, func(character *models.Character, err error) {
		if err != nil {
			zLog.Error("Failed to get character", zap.Error(err))
			resp := protocol.CharacterLoginResponse{
				Success:  false,
				ErrorMsg: "服务器错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), respData)
			return
		}

		if character == nil {
			// 角色不存在
			resp := protocol.CharacterLoginResponse{
				Success:  false,
				ErrorMsg: "角色不存在",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), respData)
			return
		}

		// 检查角色是否已经在线
		if existingSessionId, ok := h.charSession[req.CharacterId]; ok {
			// 如果角色已经在线，踢掉旧会话
			// 这里需要通过sessionManager获取旧会话并关闭它
			// 暂时注释掉这部分代码，因为需要sessionManager的支持
			zLog.Info("Character already online, kicking old session", zap.Int64("characterId", req.CharacterId), zap.Int64("oldSessionId", existingSessionId))
			// oldSession := sessionManager.GetSessionByID(existingSessionId)
			// if oldSession != nil {
			//     oldSession.Close()
			//     delete(h.sessionChar, existingSessionId)
			// }
		}

		// 创建玩家对象
		player, err := h.playerService.CreatePlayer(session, character.CharID, character.CharName)
		if err != nil {
			zLog.Error("Failed to create player", zap.Error(err))
			resp := protocol.CharacterLoginResponse{
				Success:  false,
				ErrorMsg: "服务器错误",
			}
			respData, _ := proto.Marshal(&resp)
			_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), respData)
			return
		}

		// 保存角色和会话的对应关系
		h.charSession[req.CharacterId] = int64(session.GetSid())
		h.sessionChar[int64(session.GetSid())] = req.CharacterId

		// 记录角色登录日志
		loginLog := &models.LoginLog{
			CharID:    character.CharID,
			CharName:  character.CharName,
			OpType:    1, // 1表示登录
			CreatedAt: time.Now(),
		}
		h.dbManager.LoginLogDAO.CreateLoginLog(loginLog, func(logID int64, err error) {
			if err != nil {
				zLog.Error("Failed to record login log", zap.Error(err))
			}
		})

		// 发送角色登录成功响应
		resp := protocol.CharacterLoginResponse{
			Success:  true,
			ErrorMsg: "",
			PlayerId: player.GetPlayerId(),
			Name:     player.GetName(),
			Level:    int32(player.GetBasicInfo().Level),
			Gold:     player.GetBasicInfo().Gold.Load(),
		}
		respData, _ := proto.Marshal(&resp)
		_ = session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN), respData)
	})

	return nil
}

// handleCharacterLogout 处理角色登出请求
func (h *playerHandler) handleCharacterLogout(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character logout request", zap.Int64("sessionId", int64(session.GetSid())))

	// 获取角色ID
	characterId, ok := h.sessionChar[int64(session.GetSid())]
	if !ok {
		// 未找到角色信息
		resp := protocol.CharacterLogoutResponse{
			Success:  false,
			ErrorMsg: "角色未登录",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGOUT), respData)
	}

	// 获取玩家对象
	player := h.playerService.GetPlayer(characterId)
	if player != nil {
		// 处理玩家退出逻辑
		player.OnDisconnect()
		h.playerService.RemovePlayer(player.GetPlayerId())
		zLog.Info("Player logged out", zap.Int64("playerId", player.GetPlayerId()), zap.String("name", player.GetName()))

		// 记录角色登出日志
		loginLog := &models.LoginLog{
			CharID:    characterId,
			CharName:  player.GetName(),
			OpType:    2, // 2表示登出
			CreatedAt: time.Now(),
		}
		h.dbManager.LoginLogDAO.CreateLoginLog(loginLog, func(logID int64, err error) {
			if err != nil {
				zLog.Error("Failed to record logout log", zap.Error(err))
			}
		})
	}

	// 清理角色和会话的对应关系
	delete(h.charSession, characterId)
	delete(h.sessionChar, int64(session.GetSid()))

	// 发送登出成功响应
	resp := protocol.CharacterLogoutResponse{
		Success:  true,
		ErrorMsg: "",
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGOUT), respData)
}

// hashString 计算字符串的哈希值（用于生成唯一playerId）
func hashString(s string) int {
	// 使用SHA256哈希算法
	hash := sha256.Sum256([]byte(s))

	// 取前4个字节转换为整数
	result := 0
	for i := 0; i < 4; i++ {
		result = result<<8 + int(hash[i])
	}

	// 确保结果为正整数
	if result < 0 {
		result = -result
	}

	return result
}
