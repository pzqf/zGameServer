package util

import (
	"sync"
)

// ObjectPool 对象池接口
type ObjectPool interface {
	Get() interface{}
	Put(obj interface{})
	Size() int
}

// PoolItem 对象池项接口，用于重置对象状态
type PoolItem interface {
	Reset()
}

// genericObjectPool 泛型对象池实现
type genericObjectPool struct {
	pool    sync.Pool
	size    int
	maxSize int
	mu      sync.Mutex
}

// NewObjectPool 创建对象池实例
func NewObjectPool(newFunc func() interface{}, maxSize int) ObjectPool {
	return &genericObjectPool{
		pool: sync.Pool{
			New: newFunc,
		},
		maxSize: maxSize,
	}
}

// Get 从对象池获取对象
func (p *genericObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put 将对象放回对象池
func (p *genericObjectPool) Put(obj interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.size < p.maxSize {
		// 如果对象实现了PoolItem接口，重置对象状态
		if item, ok := obj.(PoolItem); ok {
			item.Reset()
		}
		p.pool.Put(obj)
		p.size++
	}
}

// Size 获取对象池大小
func (p *genericObjectPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.size
}
