package protolayer

import (
	"github.com/pzqf/zEngine/zNet"
	"google.golang.org/protobuf/proto"
)

// ProtobufProtocol Protobuf协议实现
type ProtobufProtocol struct {}

// NewProtobufProtocol 创建Protobuf协议实例
func NewProtobufProtocol() *ProtobufProtocol {
	return &ProtobufProtocol{}
}

// Encode 将应用层消息编码为二进制数据
func (pp *ProtobufProtocol) Encode(protoId int32, version int32, data interface{}) (*zNet.NetPacket, error) {
	// 将data转换为proto.Message
	msg, ok := data.(proto.Message)
	if !ok {
		return nil, ErrInvalidProtobufMessage
	}

	// 序列化Protobuf数据
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// 创建NetPacket
	packet := &zNet.NetPacket{
		ProtoId:  protoId,
		Version:  version,
		DataSize: int32(len(body)),
		Data:     body,
	}

	return packet, nil
}

// Decode 将二进制数据解码为应用层消息
// 注意：这里需要根据protoId创建对应的具体消息类型，目前简单实现，实际使用时需要完善
func (pp *ProtobufProtocol) Decode(packet *zNet.NetPacket) (interface{}, error) {
	// 这里只返回原始数据，实际使用时需要根据protoId解析为具体的消息类型
	// 例如：使用消息注册表根据protoId创建对应的消息实例，然后进行Unmarshal
	return packet.Data, nil
}
