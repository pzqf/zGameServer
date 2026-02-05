package handler

import (
	"time"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/db"
	"github.com/pzqf/zGameServer/db/models"
	"github.com/pzqf/zGameServer/game/common"
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
	playerSession  map[common.PlayerIdType]zNet.SessionIdType
	sessionPlayer  map[zNet.SessionIdType]common.PlayerIdType
}

func NewPlayerNetHandler(playerService *player.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService:  playerService,
		accountSession: make(map[string]zNet.SessionIdType),
		sessionAccount: make(map[zNet.SessionIdType]string),
		playerSession:  make(map[common.PlayerIdType]zNet.SessionIdType),
		sessionPlayer:  make(map[zNet.SessionIdType]common.PlayerIdType),
	}
}

func RegisterPlayerNetHandlers(router *router.PacketRouter, playerService *player.PlayerService) {
	// 创建player_handler

	handler := NewPlayerNetHandler(playerService)

	router.RegisterHandler(1001, handler.handleAccountCreate)
	router.RegisterHandler(1002, handler.handleAccountLogin)
	router.RegisterHandler(1003, handler.handlePlayerCreate)
	router.RegisterHandler(1004, handler.handlePlayerLogin)
	router.RegisterHandler(1005, handler.handlePlayerLogout)

	//todo 将来可优化，明确哪些消息需转发给玩家协程处理
	for i := 1006; i <= 2000; i++ {
		router.RegisterHandler(int32(i), handler.handlePlayerMessage)
	}
}

func (h *PlayerHandler) handlePlayerMessage(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	playerId, ok := h.sessionPlayer[session.GetSid()]
	if !ok {
		zLog.Warn("Player not found for session", zap.Uint64("sessionId", session.GetSid()))
		return nil
	}

	playerActor := h.playerService.GetPlayerActor(common.PlayerIdType(playerId))
	if playerActor == nil {
		zLog.Warn("Player actor not found", zap.Int64("playerId", int64(playerId)))
		return nil
	}

	msg := player.NewPlayerActorNetworkMessage(int64(playerId), packet)
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

	accountID, err := common.GenerateAccountID()
	if err != nil {
		zLog.Error("Failed to generate account ID", zap.Error(err))
		resp := protocol.AccountCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1001, respData)
	}

	now := time.Now()
	newAccount := &models.Account{
		AccountID:   int64(accountID),
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

	//todo 查询玩家列表

	resp := protocol.AccountLoginResponse{
		Success:  true,
		ErrorMsg: "",
		Players:  []*protocol.PlayerInfo{},
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1002, respData)
}

func (h *PlayerHandler) handlePlayerCreate(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received player create request", zap.Int64("sessionId", int64(session.GetSid())))

	account, ok := h.sessionAccount[session.GetSid()]
	if !ok {
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	var req protocol.PlayerCreateRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal player create request", zap.Error(err))
		return err
	}

	if req.Name == "" {
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "玩家名称不能为空",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	accountObj, err := db.GetDBManager().AccountRepository.GetByName(account)
	if err != nil || accountObj == nil {
		zLog.Error("Failed to get account", zap.Error(err))
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	charID, err := common.GenerateCharID()
	if err != nil {
		zLog.Error("Failed to generate character ID", zap.Error(err))
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	now := time.Now()
	newCharacter := &models.Character{
		CharID:    int64(charID),
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
		zLog.Error("Failed to create player", zap.Error(err))
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	if id <= 0 {
		resp := protocol.PlayerCreateResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1003, respData)
	}

	resp := protocol.PlayerCreateResponse{
		Success:  true,
		ErrorMsg: "",
		Player: &protocol.PlayerInfo{
			PlayerId: newCharacter.CharID,
			Name:     newCharacter.CharName,
			Level:    int32(newCharacter.Level),
			Sex:      int32(newCharacter.Sex),
			Age:      int32(newCharacter.Age),
		},
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1003, respData)
}

func (h *PlayerHandler) handlePlayerLogin(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received player login request", zap.Int64("sessionId", int64(session.GetSid())))

	_, ok := h.sessionAccount[session.GetSid()]
	if !ok {
		resp := protocol.PlayerLoginResponse{
			Success:  false,
			ErrorMsg: "请先登录账号",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	var req protocol.PlayerLoginRequest
	if err := proto.Unmarshal(packet.Data, &req); err != nil {
		zLog.Error("Failed to unmarshal player login request", zap.Error(err))
		return err
	}

	character, err := db.GetDBManager().CharacterRepository.GetByID(req.PlayerId)
	if err != nil || character == nil {
		zLog.Error("Failed to get player", zap.Error(err))
		resp := protocol.PlayerLoginResponse{
			Success:  false,
			ErrorMsg: "玩家不存在",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	_, err = h.playerService.CreatePlayerActor(session, common.PlayerIdType(character.CharID), character.CharName)
	if err != nil {
		zLog.Error("Failed to create player", zap.Error(err))
		resp := protocol.PlayerLoginResponse{
			Success:  false,
			ErrorMsg: "服务器错误",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1004, respData)
	}

	h.playerSession[common.PlayerIdType(req.PlayerId)] = session.GetSid()
	h.sessionPlayer[session.GetSid()] = common.PlayerIdType(req.PlayerId)

	resp := protocol.PlayerLoginResponse{
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

func (h *PlayerHandler) handlePlayerLogout(session *zNet.TcpServerSession, packet *zNet.NetPacket) error {
	zLog.Debug("Received player logout request", zap.Int64("sessionId", int64(session.GetSid())))

	playerId, ok := h.sessionPlayer[session.GetSid()]
	if !ok {
		resp := protocol.PlayerLogoutResponse{
			Success:  false,
			ErrorMsg: "玩家未登录",
		}
		respData, _ := proto.Marshal(&resp)
		return session.Send(1005, respData)
	}

	playerActor := h.playerService.GetPlayerActor(common.PlayerIdType(playerId))
	if playerActor != nil {
		disconnectMsg := player.NewPlayerActorMessage(int64(playerId), "disconnect", nil)
		playerActor.SendMessage(disconnectMsg)

		h.playerService.RemovePlayer(common.PlayerIdType(playerId))
		zLog.Info("Player logged out", zap.Int64("playerId", int64(playerId)), zap.String("name", playerActor.Player.GetName()))
	}

	delete(h.playerSession, common.PlayerIdType(playerId))
	delete(h.sessionPlayer, session.GetSid())

	resp := protocol.PlayerLogoutResponse{
		Success:  true,
		ErrorMsg: "",
	}
	respData, _ := proto.Marshal(&resp)
	return session.Send(1005, respData)
}
