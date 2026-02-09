package db

import (
	"sync"

	"github.com/pzqf/zEngine/zInject"
	"github.com/pzqf/zGameServer/config"
	"github.com/pzqf/zGameServer/db/connector"
	"github.com/pzqf/zGameServer/db/di"
	"github.com/pzqf/zGameServer/db/models"
	"github.com/pzqf/zGameServer/db/repository"
)

type DBManager struct {
	container             zInject.Container
	connectors            map[string]connector.DBConnector
	PlayerRepository      repository.PlayerRepository
	AccountRepository     repository.AccountRepository
	PlayerItemRepository  repository.PlayerItemRepository
	PlayerSkillRepository repository.PlayerSkillRepository
	PlayerMailRepository  repository.PlayerMailRepository
	PlayerQuestRepository repository.PlayerQuestRepository
	PlayerPetRepository   repository.PlayerPetRepository
	PlayerBuffRepository  repository.PlayerBuffRepository
	GuildRepository       repository.GuildRepository
	GuildMemberRepository repository.GuildMemberRepository
	AuctionRepository     repository.AuctionRepository
	LoginLogRepository    repository.LoginLogRepository
	MailLogRepository     repository.MailLogRepository
	QuestLogRepository    repository.QuestLogRepository
	AuctionLogRepository  repository.AuctionLogRepository
}

var (
	dbManager *DBManager
	dbOnce    sync.Once
)

func GetMgr() *DBManager {
	return dbManager
}

func InitDBManager() error {
	var err error
	dbOnce.Do(func() {
		dbManager = &DBManager{
			container:  zInject.NewContainer(),
			connectors: make(map[string]connector.DBConnector),
		}
		err = dbManager.Init()
	})
	return err
}

func (manager *DBManager) Init() error {
	dbConfigs := config.GetAllDBConfigs()

	for dbName, dbConfig := range dbConfigs {
		conn := connector.NewDBConnector(dbName, dbConfig.Driver, 1000)
		conn.Init(dbConfig)
		if err := conn.Start(); err != nil {
			return err
		}
		manager.connectors[dbName] = conn
	}

	di.RegisterConnectors(manager.container, manager.connectors)
	di.RegisterDAOs(manager.container)
	di.RegisterRepositories(manager.container)

	manager.initRepositories()
	return nil
}

func (manager *DBManager) initRepositories() {
	manager.AccountRepository = di.ResolveRepo[repository.AccountRepository](manager.container, di.RepoAccount)
	manager.PlayerRepository = di.ResolveRepo[repository.PlayerRepository](manager.container, di.RepoPlayer)
	manager.PlayerItemRepository = di.ResolveRepo[repository.PlayerItemRepository](manager.container, di.RepoPlayerItem)
	manager.PlayerSkillRepository = di.ResolveRepo[repository.PlayerSkillRepository](manager.container, di.RepoPlayerSkill)
	manager.PlayerMailRepository = di.ResolveRepo[repository.PlayerMailRepository](manager.container, di.RepoPlayerMail)
	manager.PlayerQuestRepository = di.ResolveRepo[repository.PlayerQuestRepository](manager.container, di.RepoPlayerQuest)
	manager.PlayerPetRepository = di.ResolveRepo[repository.PlayerPetRepository](manager.container, di.RepoPlayerPet)
	manager.PlayerBuffRepository = di.ResolveRepo[repository.PlayerBuffRepository](manager.container, di.RepoPlayerBuff)
	manager.GuildRepository = di.ResolveRepo[repository.GuildRepository](manager.container, di.RepoGuild)
	manager.GuildMemberRepository = di.ResolveRepo[repository.GuildMemberRepository](manager.container, di.RepoGuildMember)
	manager.AuctionRepository = di.ResolveRepo[repository.AuctionRepository](manager.container, di.RepoAuction)
	manager.LoginLogRepository = di.ResolveRepo[repository.LoginLogRepository](manager.container, di.RepoLoginLog)
	manager.MailLogRepository = di.ResolveRepo[repository.MailLogRepository](manager.container, di.RepoMailLog)
	manager.QuestLogRepository = di.ResolveRepo[repository.QuestLogRepository](manager.container, di.RepoQuestLog)
	manager.AuctionLogRepository = di.ResolveRepo[repository.AuctionLogRepository](manager.container, di.RepoAuctionLog)
}

func (manager *DBManager) Close() error {
	for _, conn := range manager.connectors {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (manager *DBManager) GetConnector(dbName string) connector.DBConnector {
	return manager.connectors[dbName]
}

func (manager *DBManager) GetAllConnectors() map[string]connector.DBConnector {
	return manager.connectors
}

func (manager *DBManager) GetContainer() zInject.Container {
	return manager.container
}

func ValidateModelTags() error {
	return models.ValidateModelTags()
}
