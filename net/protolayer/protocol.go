package protolayer

import (
	"github.com/pzqf/zEngine/zNet"
)

// Protocol 定义协议接口，实现协议的编解码与zNet层解耦
type Protocol interface {
	// Encode 将应用层消息编码为二进制数据
	Encode(protoId int32, version int32, data interface{}) (*zNet.NetPacket, error)
	// Decode 将二进制数据解码为应用层消息
	Decode(packet *zNet.NetPacket) (interface{}, error)
}

// ProtocolType 协议类型枚举
type ProtocolType int

const (
	ProtocolTypeProtobuf ProtocolType = iota
	ProtocolTypeJSON
	ProtocolTypeXML
)

// NewProtocol 根据类型创建协议实例
func NewProtocol(protocolType ProtocolType) Protocol {
	switch protocolType {
	case ProtocolTypeProtobuf:
		return NewProtobufProtocol()
	case ProtocolTypeJSON:
		return NewJSONProtocol()
	case ProtocolTypeXML:
		return NewXMLProtocol()
	default:
		return NewProtobufProtocol() // 默认使用protobuf
	}
}
