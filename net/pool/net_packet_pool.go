package pool

import (
	"github.com/pzqf/zEngine/zNet"
	"github.com/pzqf/zGameServer/util"
)

// PacketTaskPool packetTask对象池
type PacketTaskPool struct {
	util.ObjectPool
}

// NewPacketTaskPool 创建PacketTask对象池
func NewPacketTaskPool(maxSize int) *PacketTaskPool {
	return &PacketTaskPool{
		ObjectPool: util.NewObjectPool(func() interface{} {
			return &PacketTask{}
		}, maxSize),
	}
}

// GetPacketTask 从对象池获取PacketTask
func (p *PacketTaskPool) GetPacketTask() *PacketTask {
	return p.Get().(*PacketTask)
}

// PutPacketTask 将PacketTask放回对象池
func (p *PacketTaskPool) PutPacketTask(task *PacketTask) {
	p.Put(task)
}

// PacketTask 数据包处理任务，实现PoolItem接口
type PacketTask struct {
	Session interface{}
	Packet  *zNet.NetPacket
}

// Reset 重置PacketTask状态
func (t *PacketTask) Reset() {
	t.Session = nil
	t.Packet = nil
}
