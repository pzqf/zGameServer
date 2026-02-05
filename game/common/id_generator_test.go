package common

import (
	"sync"
	"testing"
)

func TestSnowflake(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}

	ids := make(map[int64]bool)
	var mu sync.Mutex

	var wg sync.WaitGroup
	numGoroutines := 100
	idsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := sf.NextID()
				if err != nil {
					t.Errorf("Failed to generate ID: %v", err)
					return
				}

				mu.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID generated: %d", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	expectedIDs := numGoroutines * idsPerGoroutine
	if len(ids) != expectedIDs {
		t.Errorf("Expected %d unique IDs, got %d", expectedIDs, len(ids))
	}
}

func TestSnowflakeParse(t *testing.T) {
	sf, err := NewSnowflake(5, 3)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}

	id, err := sf.NextID()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	_, datacenterID, workerID, sequence := sf.ParseID(id)

	if workerID != 5 {
		t.Errorf("Expected workerID 5, got %d", workerID)
	}
	if datacenterID != 3 {
		t.Errorf("Expected datacenterID 3, got %d", datacenterID)
	}
	if sequence < 0 || sequence > 4095 {
		t.Errorf("Invalid sequence number: %d", sequence)
	}
}

func TestSnowflakeInvalidWorkerID(t *testing.T) {
	_, err := NewSnowflake(32, 1)
	if err != ErrInvalidWorkerID {
		t.Errorf("Expected ErrInvalidWorkerID, got %v", err)
	}
}

func TestSnowflakeInvalidDatacenterID(t *testing.T) {
	_, err := NewSnowflake(1, 32)
	if err != ErrInvalidDatacenter {
		t.Errorf("Expected ErrInvalidDatacenter, got %v", err)
	}
}

func TestGeneratePlayerID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GeneratePlayerID()
	if err != nil {
		t.Fatalf("Failed to generate player ID: %v", err)
	}

	id2, err := GeneratePlayerID()
	if err != nil {
		t.Fatalf("Failed to generate player ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different player IDs, got both %d", id1)
	}
}

func TestGenerateMapID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateMapID()
	if err != nil {
		t.Fatalf("Failed to generate map ID: %v", err)
	}

	id2, err := GenerateMapID()
	if err != nil {
		t.Fatalf("Failed to generate map ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different map IDs, got both %d", id1)
	}
}

func TestGenerateObjectID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateObjectID()
	if err != nil {
		t.Fatalf("Failed to generate object ID: %v", err)
	}

	id2, err := GenerateObjectID()
	if err != nil {
		t.Fatalf("Failed to generate object ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different object IDs, got both %d", id1)
	}
}

func TestGenerateAccountID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateAccountID()
	if err != nil {
		t.Fatalf("Failed to generate account ID: %v", err)
	}

	id2, err := GenerateAccountID()
	if err != nil {
		t.Fatalf("Failed to generate account ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different account IDs, got both %d", id1)
	}
}

func TestGenerateCharID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateCharID()
	if err != nil {
		t.Fatalf("Failed to generate char ID: %v", err)
	}

	id2, err := GenerateCharID()
	if err != nil {
		t.Fatalf("Failed to generate char ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different char IDs, got both %d", id1)
	}
}

func TestGenerateGroupID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateGroupID()
	if err != nil {
		t.Fatalf("Failed to generate group ID: %v", err)
	}

	id2, err := GenerateGroupID()
	if err != nil {
		t.Fatalf("Failed to generate group ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different group IDs, got both %d", id1)
	}
}

func TestGenerateComboID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateComboID()
	if err != nil {
		t.Fatalf("Failed to generate combo ID: %v", err)
	}

	id2, err := GenerateComboID()
	if err != nil {
		t.Fatalf("Failed to generate combo ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different combo IDs, got both %d", id1)
	}
}

func TestGenerateVisualID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateVisualID()
	if err != nil {
		t.Fatalf("Failed to generate visual ID: %v", err)
	}

	id2, err := GenerateVisualID()
	if err != nil {
		t.Fatalf("Failed to generate visual ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different visual IDs, got both %d", id1)
	}
}

func TestGenerateLogID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateLogID()
	if err != nil {
		t.Fatalf("Failed to generate log ID: %v", err)
	}

	id2, err := GenerateLogID()
	if err != nil {
		t.Fatalf("Failed to generate log ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different log IDs, got both %d", id1)
	}
}

func TestGenerateItemID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateItemID()
	if err != nil {
		t.Fatalf("Failed to generate item ID: %v", err)
	}

	id2, err := GenerateItemID()
	if err != nil {
		t.Fatalf("Failed to generate item ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different item IDs, got both %d", id1)
	}
}

func TestGenerateMailID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateMailID()
	if err != nil {
		t.Fatalf("Failed to generate mail ID: %v", err)
	}

	id2, err := GenerateMailID()
	if err != nil {
		t.Fatalf("Failed to generate mail ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different mail IDs, got both %d", id1)
	}
}

func TestGenerateSkillID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateSkillID()
	if err != nil {
		t.Fatalf("Failed to generate skill ID: %v", err)
	}

	id2, err := GenerateSkillID()
	if err != nil {
		t.Fatalf("Failed to generate skill ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different skill IDs, got both %d", id1)
	}
}

func TestGenerateTaskID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateTaskID()
	if err != nil {
		t.Fatalf("Failed to generate task ID: %v", err)
	}

	id2, err := GenerateTaskID()
	if err != nil {
		t.Fatalf("Failed to generate task ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different task IDs, got both %d", id1)
	}
}

func TestGenerateGuildID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateGuildID()
	if err != nil {
		t.Fatalf("Failed to generate guild ID: %v", err)
	}

	id2, err := GenerateGuildID()
	if err != nil {
		t.Fatalf("Failed to generate guild ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different guild IDs, got both %d", id1)
	}
}

func TestGenerateApplyID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateApplyID()
	if err != nil {
		t.Fatalf("Failed to generate apply ID: %v", err)
	}

	id2, err := GenerateApplyID()
	if err != nil {
		t.Fatalf("Failed to generate apply ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different apply IDs, got both %d", id1)
	}
}

func TestGenerateAuctionID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateAuctionID()
	if err != nil {
		t.Fatalf("Failed to generate auction ID: %v", err)
	}

	id2, err := GenerateAuctionID()
	if err != nil {
		t.Fatalf("Failed to generate auction ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different auction IDs, got both %d", id1)
	}
}

func TestGenerateBidID(t *testing.T) {
	err := InitIDGenerator(1, 1)
	if err != nil {
		t.Fatalf("Failed to initialize ID generator: %v", err)
	}

	id1, err := GenerateBidID()
	if err != nil {
		t.Fatalf("Failed to generate bid ID: %v", err)
	}

	id2, err := GenerateBidID()
	if err != nil {
		t.Fatalf("Failed to generate bid ID: %v", err)
	}

	if id1 == id2 {
		t.Errorf("Expected different bid IDs, got both %d", id1)
	}
}
