package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/protocol"
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

	fmt.Printf("Account login response: %+v\n", loginResp)

	if loginResp.Success {
		// 检查是否有角色
		var characterId int64
		if len(loginResp.Characters) == 0 {
			// 创建新角色
			fmt.Println("No characters found, creating new character...")
			charCreateReq := protocol.CharacterCreateRequest{
				Name: "TestChar7",
				Sex:  1,
				Age:  20,
			}

			charCreateData, err := proto.Marshal(&charCreateReq)
			if err != nil {
				fmt.Printf("Failed to marshal character create request: %v\n", err)
				return
			}

			// 创建角色创建数据包
			charCreatePacket := &zNet.NetPacket{
				ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_CREATE),
				DataSize: int32(len(charCreateData)),
				Version:  0,
				Data:     charCreateData,
			}

			// 发送角色创建请求
			if _, err := conn.Write(charCreatePacket.Marshal()); err != nil {
				fmt.Printf("Failed to send character create request: %v\n", err)
				return
			}

			fmt.Println("Sent character create request")

			// 接收角色创建响应
			charCreateRespPacket, err := readPacket(conn)
			if err != nil {
				fmt.Printf("Failed to read character create response: %v\n", err)
				return
			}

			// 解析角色创建响应
			var charCreateResp protocol.CharacterCreateResponse
			if err := proto.Unmarshal(charCreateRespPacket.Data, &charCreateResp); err != nil {
				fmt.Printf("Failed to unmarshal character create response: %v\n", err)
				return
			}

			fmt.Printf("Character create response: %+v\n", charCreateResp)

			if charCreateResp.Success {
				characterId = charCreateResp.Character.CharacterId
			}
		} else {
			// 使用第一个角色
			characterId = loginResp.Characters[0].CharacterId
		}

		// 角色登录
		charLoginReq := protocol.CharacterLoginRequest{
			CharacterId: characterId,
		}

		charLoginData, err := proto.Marshal(&charLoginReq)
		if err != nil {
			fmt.Printf("Failed to marshal character login request: %v\n", err)
			return
		}

		// 创建角色登录数据包
		charLoginPacket := &zNet.NetPacket{
			ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGIN),
			DataSize: int32(len(charLoginData)),
			Version:  0,
			Data:     charLoginData,
		}

		// 发送角色登录请求
		if _, err := conn.Write(charLoginPacket.Marshal()); err != nil {
			fmt.Printf("Failed to send character login request: %v\n", err)
			return
		}

		fmt.Println("Sent character login request")

		// 接收角色登录响应
		charLoginRespPacket, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Failed to read character login response: %v\n", err)
			return
		}

		// 解析角色登录响应
		var charLoginResp protocol.CharacterLoginResponse
		if err := proto.Unmarshal(charLoginRespPacket.Data, &charLoginResp); err != nil {
			fmt.Printf("Failed to unmarshal character login response: %v\n", err)
			return
		}

		fmt.Printf("Character login response: %+v\n", charLoginResp)

		// 等待一段时间
		time.Sleep(2 * time.Second)

		// 测试角色登出请求
		charLogoutReq := protocol.CharacterLogoutRequest{
			CharacterId: characterId,
		}

		charLogoutData, err := proto.Marshal(&charLogoutReq)
		if err != nil {
			fmt.Printf("Failed to marshal character logout request: %v\n", err)
			return
		}

		// 创建角色登出数据包
		charLogoutPacket := &zNet.NetPacket{
			ProtoId:  int32(protocol.PlayerMsgId_MSG_PLAYER_CHARACTER_LOGOUT),
			DataSize: int32(len(charLogoutData)),
			Version:  0,
			Data:     charLogoutData,
		}

		// 发送角色登出请求
		if _, err := conn.Write(charLogoutPacket.Marshal()); err != nil {
			fmt.Printf("Failed to send character logout request: %v\n", err)
			return
		}

		fmt.Println("Sent character logout request")

		// 接收角色登出响应
		charLogoutRespPacket, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Failed to read character logout response: %v\n", err)
			return
		}

		// 解析角色登出响应
		var charLogoutResp protocol.CharacterLogoutResponse
		if err := proto.Unmarshal(charLogoutRespPacket.Data, &charLogoutResp); err != nil {
			fmt.Printf("Failed to unmarshal character logout response: %v\n", err)
			return
		}

		fmt.Printf("Character logout response: %+v\n", charLogoutResp)
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
