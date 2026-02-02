package handler

import (
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

type PlayerHandler struct {
	playerService  *player.PlayerService
	accountSession map[string]zNet.SessionIdType
	sessionAccount map[zNet.SessionIdType]string
	charSession    map[int64]zNet.SessionIdType
	sessionChar    map[zNet.SessionIdType]int64
}

func NewPlayerNetHandler(playerService *player.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService:  playerService,
		accountSession: make(map[string]zNet.SessionIdType),
		sessionAccount: make(map[zNet.SessionIdType]string),
		charSession:    make(map[int64]zNet.SessionIdType),
		sessionChar:    make(map[zNet.SessionIdType]int64),
	}
}

func RegisterPlayerNetHandlers(router *router.PacketRouter, playerService *player.PlayerService) {
	// 创建player_handler

	handler := NewPlayerNetHandler(playerService)

	router.RegisterHandler(1001, handler.handleAccountCreate)
	router.RegisterHandler(1002, handler.handleAccountLogin)
	router.RegisterHandler(1003, handler.handleCharacterCreate)
	router.RegisterHandler(1004, handler.handleCharacterLogin)
	router.RegisterHandler(1005, handler.handleCharacterLogout)

	//todo 将来可优化，明确哪些消息需转发给玩家协程处理
	for i := 1006; i <= 2000; i++ {
		router.RegisterHandler(int32(i), handler.handlePlayerMessage)
	}
}

func (h *PlayerHandler) handlePlayerMessage(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	playerId, ok := h.sessionChar[session.GetSid()]
	if !ok {
		zLog.Warn("Player not found for session", zap.Uint64("sessionId", session.GetSid()))
		return nil
	}

	playerActor := h.playerService.GetPlayerActor(playerId)
	if playerActor == nil {
		zLog.Warn("Player actor not found", zap.Int64("playerId", playerId))
		return nil
	}

	msg := player.NewPlayerActorNetworkMessage(playerId, packet)
	playerActor.SendMessage(msg)
	return nil
}

func (h *PlayerHandler) handleAccountCreate(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received account create request", zap.Uint64("sessionId", session.GetSid()))

	var req protocol.AccountCreateRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal account create request", zap.Error(err))
		return err
	}

	if req.Account == "" || req.Password == "" {
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "账号或密码不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	account, err := db.GetDBManager().AccountRepository.GetByName(req.Account)
	if err != nil {
		zLog.Error("Failed to check account existence", zap.Error(err))
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	if account != nil {
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "账号已存在",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	now := time.Now()
	newAccount := &models.Account{
		AccountID:   time.Now().UnixNano() / 1000000,
		AccountName: req.Account,
		Password:    req.Password,
		Status:      1,
		CreatedAt:   now,
		LastLoginAt: now,
	}

	id, err := db.GetDBManager().AccountRepository.Create(newAccount)
	if err != nil {
		zLog.Error("Failed to create account", zap.Error(err))
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	if id <= 0 {
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	resp := protocol.AccountCreateResponse{
		Success:  true,
		ErrorMsg: "",
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1001, respData)
}

func (h *PlayerHandler) handleAccountLogin(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received account login request", zap.Int64("sessionId", int64(session.GetSid())))

	var req protocol.AccountLoginRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal account login request", zap.Error(err))
		return err
	}

	if req.Account == "" || req.Password == "" {
		resp := protocol.AccountLoginResponse{
			Success:  false,
			ErrorMsg: "账号或密码不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1002, respData)
	}

	account, err := db.GetDBManager().AccountRepository.GetByName(req.Account)
	if err != nil {
		zLog.Error("Failed to get account", zap.Error(err))
		resp := protocol.AccountLoginResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1002, respData)
	}

	if account == nil || account.Password != req.Password {
		resp := protocol.AccountLoginResponse{
			Success:  false,
			ErrorMsg: "账号或密码错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1002, respData)
	}

	account.LastLoginAt = time.Now()
	_, err = db.GetDBManager().AccountRepository.Update(account)
	if err != nil {
		zLog.Error("Failed to update last login time", zap.Error(err))
	}

	h.accountSession[req.Account] = session.GetSid()
	h.sessionAccount[session.GetSid()] = req.Account

	//todo 查询角色列表

	resp := protocol.AccountLoginResponse{
		Success:    true,
		ErrorMsg:   "",
		Characters: []*protocol.CharacterInfo{},
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1002, respData)
}

func (h *PlayerHandler) handleCharacterCreate(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character create request", zap.Int64("sessionId", int64(session.GetSid())))

	account, ok := h.sessionAccount[session.GetSid()]
	if !ok {
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	var req protocol.CharacterCreateRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal character create request", zap.Error(err))
		return err
	}

	if req.Name == "" {
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "角色名称不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	accountObj, err := db.GetDBManager().AccountRepository.GetByName(account)
	if err != nil || accountObj == nil {
		zLog.Error("Failed to get account", zap.Error(err))
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	now := time.Now()
	newCharacter := &models.Character{
		CharID:    time.Now().UnixNano() / 1000000,
		AccountID: accountObj.AccountID,
		CharName:  req.Name,
		Sex:       int(req.Sex),
		Age:       int(req.Age),
		Level:     1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := db.GetDBManager().CharacterRepository.Create(newCharacter)
	if err != nil {
		zLog.Error("Failed to create character", zap.Error(err))
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	if id <= 0 {
		resp := protocol.CharacterCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

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
	return session.Send(1003, respData)
}

func (h *PlayerHandler) handleCharacterLogin(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character login request", zap.Int64("sessionId", int64(session.GetSid())))

	_, ok := h.sessionAccount[session.GetSid()]
	if !ok {
		resp := protocol.CharacterLoginResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	var req protocol.CharacterLoginRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal character login request", zap.Error(err))
		return err
	}

	character, err := db.GetDBManager().CharacterRepository.GetByID(req.CharacterId)
	if err != nil || character == nil {
		zLog.Error("Failed to get character", zap.Error(err))
		resp := protocol.CharacterLoginResponse{
			Success:  false,
			ErrorMsg: "角色不存在",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	_, err = h.playerService.CreatePlayerActor(session, character.CharID, character.CharName)
	if err != nil {
		zLog.Error("Failed to create player", zap.Error(err))
		resp := protocol.CharacterLoginResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	h.charSession[req.CharacterId] = session.GetSid()
	h.sessionChar[session.GetSid()] = req.CharacterId

	resp := protocol.CharacterLoginResponse{
		Success:  true,
		ErrorMsg: "",
		PlayerId: character.CharID,
		Name:     character.CharName,
		Level:    int32(character.Level),
		Gold:     1000,
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1004, respData)
}

func (h *PlayerHandler) handleCharacterLogout(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received character logout request", zap.Int64("sessionId", int64(session.GetSid())))

	characterId, ok := h.sessionChar[session.GetSid()]
	if !ok {
		resp := protocol.CharacterLogoutResponse{
			Success:  false,
			ErrorMsg: "角色未登录",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1005, respData)
	}

	playerActor := h.playerService.GetPlayerActor(characterId)
	if playerActor != nil {
		disconnectMsg := player.NewPlayerActorMessage(characterId, "disconnect", nil)
		playerActor.SendMessage(disconnectMsg)

		h.playerService.RemovePlayer(characterId)
		zLog.Info("Player logged out", zap.Int64("playerId", characterId), zap.String("name", playerActor.Player.GetName()))
	}

	delete(h.charSession, characterId)
	delete(h.sessionChar, session.GetSid())

	resp := protocol.CharacterLogoutResponse{
		Success:  true,
		ErrorMsg: "",
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1005, respData)
}
