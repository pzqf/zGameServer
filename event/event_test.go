package event

import (
	"sync"
	"testing"
	"time"

	"github.com/pzqf/zEngine/zEvent"
)

// 测试事件总线的基本功能
func TestEventBusBasic(t *testing.T) {
	// 创建事件总线实例
	bus := zEvent.NewEventBus()
	defer bus.Close()

	// 定义事件处理函数
	var receivedEvent *zEvent.Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *zEvent.Event) {
		receivedEvent = event
		wg.Done()
	}

	// 订阅事件
	bus.Subscribe(EventPlayerExpAdd, handler)

	// 发布事件
	eventData := &PlayerExpEventData{
		PlayerID: 1001,
		Exp:      100,
	}
	event := NewEvent(EventPlayerExpAdd, "test_source", eventData)
	bus.Publish(event)

	// 等待事件处理完成
	wg.Wait()

	// 验证事件是否被正确处理
	if receivedEvent == nil {
		t.Error("Event was not received")
		return
	}

	if receivedEvent.Type != EventPlayerExpAdd {
		t.Errorf("Expected event type %d, got %d", EventPlayerExpAdd, receivedEvent.Type)
	}

	if receivedEvent.Source != "test_source" {
		t.Errorf("Expected event source 'test_source', got %v", receivedEvent.Source)
	}

	if data, ok := receivedEvent.Data.(*PlayerExpEventData); ok {
		if data.PlayerID != 1001 {
			t.Errorf("Expected playerID 1001, got %d", data.PlayerID)
		}
		if data.Exp != 100 {
			t.Errorf("Expected exp 100, got %d", data.Exp)
		}
	} else {
		t.Error("Expected PlayerExpEventData type")
	}
}

// 测试多协程环境下的事件处理
func TestEventBusConcurrent(t *testing.T) {
	// 创建事件总线实例
	bus := zEvent.NewEventBus()
	defer bus.Close()

	// 定义事件处理函数
	var receivedCount int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(100)

	handler := func(event *zEvent.Event) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
	}

	// 订阅事件
	bus.Subscribe(EventPlayerGoldAdd, handler)

	// 多协程发布事件
	for i := 0; i < 100; i++ {
		go func(index int) {
			eventData := &PlayerGoldEventData{
				PlayerID: 1001,
				Gold:     int64(index),
			}
			// 发布事件
			event := zEvent.NewEvent(EventPlayerGoldAdd, "test_source", eventData)
			bus.Publish(event)
		}(i)
	}

	// 等待所有事件处理完成
	wg.Wait()

	// 验证所有事件是否被正确处理
	mu.Lock()
	if receivedCount != 100 {
		t.Errorf("Expected 100 events, got %d", receivedCount)
	}
	mu.Unlock()
}

// 测试全局事件总线
func TestGlobalEventBus(t *testing.T) {
	// 获取全局事件总线实例
	bus := GetGlobalEventBus()

	// 定义事件处理函数
	var receivedEvent *zEvent.Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *zEvent.Event) {
		receivedEvent = event
		wg.Done()
	}

	// 订阅事件
	bus.Subscribe(EventPlayerLogin, handler)

	// 发布事件
	event := zEvent.NewEvent(EventPlayerLogin, "test_source", nil)
	bus.Publish(event)

	// 等待事件处理完成
	wg.Wait()

	// 验证事件是否被正确处理
	if receivedEvent == nil {
		t.Error("Event was not received")
		return
	}

	if receivedEvent.Type != EventPlayerLogin {
		t.Errorf("Expected event type %d, got %d", EventPlayerLogin, receivedEvent.Type)
	}
}

// 测试事件总线的运行状态
func TestEventBusRunningStatus(t *testing.T) {
	// 创建事件总线实例
	bus := zEvent.NewEventBus()
	defer bus.Close()

	// 定义事件处理函数
	var receivedEvent *zEvent.Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *zEvent.Event) {
		receivedEvent = event
		wg.Done()
	}

	// 订阅事件
	bus.Subscribe(EventPlayerLogout, handler)

	// 关闭事件总线
	bus.Close()

	// 发布事件
	event := zEvent.NewEvent(EventPlayerLogout, "test_source", nil)
	bus.Publish(event)

	// 等待一小段时间，确保事件不会被处理
	time.Sleep(100 * time.Millisecond)

	// 验证事件是否被正确忽略
	if receivedEvent != nil {
		t.Error("Event should not be received after bus is closed")
	}
}
