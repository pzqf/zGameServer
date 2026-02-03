package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/net/protocol"
	"google.golang.org/protobuf/proto"
)

// 测试客户端
func main() {
	// 连接到游戏服务器
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to game server!")

	// 直接进行登录，跳过账号创建
	loginReq := protocol.AccountLoginRequest{
		Account:    "testuser7",
		Password:   "testpass",
		DeviceId:   "testdevice",
		DeviceType: 1,
		Version:    "1.0.0",
	}

	loginData, err := proto.Marshal(&loginReq)
	if err != nil {
		fmt.Printf("Failed to marshal login request: %v\n", err)
		return
	}

	// 创建登录数据包
	loginPacket := &zNet.NetPacket{
		ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_ACCOUNT_LOGIN),
		DataSize: int32(len(loginData)),
		Version:  0,
		Data:     loginData,
	}

	// 发送登录请求
	if _, err := conn.Write(loginPacket.Marshal()); err != nil {
		fmt.Printf("Failed to send login request: %v\n", err)
		return
	}

	fmt.Println("Sent account login request")

	// 接收登录响应
	loginRespPacket, err := readPacket(conn)
	if err != nil {
		fmt.Printf("Failed to read login response: %v\n", err)
		return
	}

	// 解析登录响应
	var loginResp protocol.AccountLoginResponse
	if err := proto.Unmarshal(loginRespPacket.Data, &loginResp); err != nil {
		fmt.Printf("Failed to unmarshal login response: %v\n", err)
		return
	}

	fmt.Printf("Account login response: Success=%v, Message=%s\n", loginResp.Success, loginResp.ErrorMsg)

	if loginResp.Success {
		// 检查是否有角色
		var playerId int64
		if len(loginResp.Players) == 0 {
			// 创建新角色
			fmt.Println("No players found, creating new player...")
			playerCreateReq := protocol.PlayerCreateRequest{
				Name: "TestPlayer7",
				Sex:  1,
				Age:  20,
			}

			playerCreateData, err := proto.Marshal(&playerCreateReq)
			if err != nil {
				fmt.Printf("Failed to marshal player create request: %v\n", err)
				return
			}

			// 创建玩家创建数据包
			playerCreatePacket := &zNet.NetPacket{
				ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_PLAYER_CREATE),
				DataSize: int32(len(playerCreateData)),
				Version:  0,
				Data:     playerCreateData,
			}

			// 发送玩家创建请求
			if _, err := conn.Write(playerCreatePacket.Marshal()); err != nil {
				fmt.Printf("Failed to send player create request: %v\n", err)
				return
			}

			fmt.Println("Sent player create request")

			// 接收玩家创建响应
			playerCreateRespPacket, err := readPacket(conn)
			if err != nil {
				fmt.Printf("Failed to read player create response: %v\n", err)
				return
			}

			// 解析玩家创建响应
			var playerCreateResp protocol.PlayerCreateResponse
			if err := proto.Unmarshal(playerCreateRespPacket.Data, &playerCreateResp); err != nil {
				fmt.Printf("Failed to unmarshal player create response: %v\n", err)
				return
			}

			fmt.Printf("Player create response: Success=%v, Message=%s\n", playerCreateResp.Success, playerCreateResp.ErrorMsg)

			if playerCreateResp.Success {
				playerId = playerCreateResp.Player.PlayerId
			}
		} else {
			// 使用第一个玩家
			playerId = loginResp.Players[0].PlayerId
		}

		// 玩家登录
		playerLoginReq := protocol.PlayerLoginRequest{
			PlayerId: playerId,
		}

		playerLoginData, err := proto.Marshal(&playerLoginReq)
		if err != nil {
			fmt.Printf("Failed to marshal player login request: %v\n", err)
			return
		}

		// 创建玩家登录数据包
		playerLoginPacket := &zNet.NetPacket{
			ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_PLAYER_LOGIN),
			DataSize: int32(len(playerLoginData)),
			Version:  0,
			Data:     playerLoginData,
		}

		// 发送玩家登录请求
		if _, err := conn.Write(playerLoginPacket.Marshal()); err != nil {
			fmt.Printf("Failed to send player login request: %v\n", err)
			return
		}

		fmt.Println("Sent player login request")

		// 接收玩家登录响应
		playerLoginRespPacket, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Failed to read player login response: %v\n", err)
			return
		}

		// 解析玩家登录响应
		var playerLoginResp protocol.PlayerLoginResponse
		if err := proto.Unmarshal(playerLoginRespPacket.Data, &playerLoginResp); err != nil {
			fmt.Printf("Failed to unmarshal player login response: %v\n", err)
			return
		}

		fmt.Printf("Player login response: Success=%v, Message=%s\n", playerLoginResp.Success, playerLoginResp.ErrorMsg)

		// 等待一段时间
		time.Sleep(2 * time.Second)

		// 测试玩家登出请求
		playerLogoutReq := protocol.PlayerLogoutRequest{
			PlayerId: playerId,
		}

		playerLogoutData, err := proto.Marshal(&playerLogoutReq)
		if err != nil {
			fmt.Printf("Failed to marshal player logout request: %v\n", err)
			return
		}

		// 创建玩家登出数据包
		playerLogoutPacket := &zNet.NetPacket{
			ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_PLAYER_LOGOUT),
			DataSize: int32(len(playerLogoutData)),
			Version:  0,
			Data:     playerLogoutData,
		}

		// 发送玩家登出请求
		if _, err := conn.Write(playerLogoutPacket.Marshal()); err != nil {
			fmt.Printf("Failed to send player logout request: %v\n", err)
			return
		}

		fmt.Println("Sent player logout request")

		// 接收玩家登出响应
		playerLogoutRespPacket, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Failed to read player logout response: %v\n", err)
			return
		}

		// 解析玩家登出响应
		var playerLogoutResp protocol.PlayerLogoutResponse
		if err := proto.Unmarshal(playerLogoutRespPacket.Data, &playerLogoutResp); err != nil {
			fmt.Printf("Failed to unmarshal player logout response: %v\n", err)
			return
		}

		fmt.Printf("Player logout response: Success=%v, Message=%s\n", playerLogoutResp.Success, playerLogoutResp.ErrorMsg)
	}

	fmt.Println("Test completed!")
}

// readPacket 从连接中读取一个完整的数据包
func readPacket(conn net.Conn) (*zNet.NetPacket, error) {
	// 读取包头部
	headBuf := make([]byte, zNet.NetPacketHeadSize)
	if _, err := io.ReadFull(conn, headBuf); err != nil {
		return nil, err
	}

	// 解析包头部
	packet := &zNet.NetPacket{}
	if err := packet.UnmarshalHead(headBuf); err != nil {
		return nil, err
	}

	// 读取数据部分
	if packet.DataSize > 0 {
		packet.Data = make([]byte, packet.DataSize)
		if _, err := io.ReadFull(conn, packet.Data); err != nil {
			return nil, err
		}
	}

	return packet, nil
}
