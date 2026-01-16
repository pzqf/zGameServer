package tables

import (
	"sync"

	"github.com/pzqf/zGameServer/config/models"
)

// PetTableLoader 宠物表加载器
type PetTableLoader struct {
	mu   sync.RWMutex
	pets map[int32]*models.Pet
}

// NewPetTableLoader 创建宠物表加载器
func NewPetTableLoader() *PetTableLoader {
	return &PetTableLoader{
		pets: make(map[int32]*models.Pet),
	}
}

// Load 加载宠物表数据
func (ptl *PetTableLoader) Load(dir string) error {
	config := ExcelConfig{
		FileName:   "pet.xlsx",
		SheetName:  "Sheet1",
		MinColumns: 10,
		TableName:  "pets",
	}

	// 使用临时map批量加载数据，减少锁竞争
	tempPets := make(map[int32]*models.Pet)

	err := ReadExcelFile(config, dir, func(row []string) error {
		pet := &models.Pet{
			PetID:         StrToInt32(row[0]),
			Name:          row[1],
			Type:          StrToInt32(row[2]),
			BaseHP:        StrToInt32(row[3]),
			BaseAttack:    StrToInt32(row[4]),
			BaseDefense:   StrToInt32(row[5]),
			GrowthRate:    StrToFloat32(row[6]),
			SkillID:       StrToInt32(row[7]),
			ObtainMethod:  row[8],
			Rarity:        StrToInt32(row[9]),
		}

		tempPets[pet.PetID] = pet
		return nil
	})

	// 批量写入到目标map，只需加一次锁
	if err == nil {
		ptl.mu.Lock()
		ptl.pets = tempPets
		ptl.mu.Unlock()
	}

	return err
}

// GetTableName 获取表格名称
func (ptl *PetTableLoader) GetTableName() string {
	return "pets"
}

// GetPet 根据ID获取宠物配置
func (ptl *PetTableLoader) GetPet(petID int32) (*models.Pet, bool) {
	ptl.mu.RLock()
	pet, ok := ptl.pets[petID]
	ptl.mu.RUnlock()
	return pet, ok
}

// GetAllPets 获取所有宠物配置
func (ptl *PetTableLoader) GetAllPets() map[int32]*models.Pet {
	ptl.mu.RLock()
	// 创建一个副本，避免外部修改内部数据
	petsCopy := make(map[int32]*models.Pet, len(ptl.pets))
	for id, pet := range ptl.pets {
		petsCopy[id] = pet
	}
	ptl.mu.RUnlock()
	return petsCopy
}
