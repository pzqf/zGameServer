package global

import (
	"testing"
	"go.uber.org/zap"
)

func TestGuildSystem(t *testing.T) {
	// 初始化日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建公会服务
	guildService := NewGuildService(logger)
	if guildService == nil {
		t.Fatal("Failed to create guild service")
	}

	// 初始化公会服务
	if err := guildService.Init(); err != nil {
		t.Fatalf("Failed to initialize guild service: %v", err)
	}
	defer guildService.Close()

	// 测试1: 创建公会
	guildId := int64(1)
	guildName := "TestGuild"
	leaderId := int64(1001)
	leaderName := "Leader"

	guild, err := guildService.CreateGuild(guildId, guildName, leaderId, leaderName)
	if err != nil {
		t.Fatalf("Failed to create guild: %v", err)
	}
	if guild == nil {
		t.Fatal("Created guild is nil")
	}
	if guild.name != guildName {
		t.Errorf("Guild name mismatch: expected %s, got %s", guildName, guild.name)
	}
	if guild.leaderId != leaderId {
		t.Errorf("Guild leader mismatch: expected %d, got %d", leaderId, guild.leaderId)
	}

	// 测试2: 加入公会
	playerId := int64(1002)
	playerName := "Member"
	if err := guildService.JoinGuild(playerId, playerName, guildId); err != nil {
		t.Fatalf("Failed to join guild: %v", err)
	}

	// 测试3: 检查玩家是否在公会中
	guildByPlayer, exists := guildService.GetGuildByPlayer(playerId)
	if !exists {
		t.Fatal("Player should be in guild")
	}
	if guildByPlayer.guildId != guildId {
		t.Errorf("Guild ID mismatch: expected %d, got %d", guildId, guildByPlayer.guildId)
	}

	// 测试4: 设置成员职位
	newPosition := GuildPositionVice
	if err := guildService.SetGuildMemberPosition(leaderId, playerId, newPosition); err != nil {
		t.Fatalf("Failed to set guild member position: %v", err)
	}

	// 测试5: 检查权限（副会长应该有踢人权限）
	if err := guildService.CheckGuildPermission(playerId, GuildPermissionKickMember); err != nil {
		t.Fatalf("Vice leader should have kick permission: %v", err)
	}

	// 测试6: 副会长尝试升级公会（应该没有权限）
	if err := guildService.UpgradeGuild(playerId); err == nil {
		t.Fatal("Vice leader should not have upgrade permission")
	}

	// 测试7: 会长升级公会
	// 先增加公会经验
	if err := guildService.UpdateGuildMemberContribution(leaderId, 10000); err != nil {
		t.Fatalf("Failed to update guild member contribution: %v", err)
	}
	// 升级公会
	if err := guildService.UpgradeGuild(leaderId); err != nil {
		t.Fatalf("Failed to upgrade guild: %v", err)
	}
	// 检查公会等级是否提升
	updatedGuild, exists := guildService.GetGuild(guildId)
	if !exists {
		t.Fatal("Guild should exist after upgrade")
	}
	if updatedGuild.level != 2 {
		t.Errorf("Guild level mismatch: expected 2, got %d", updatedGuild.level)
	}

	// 测试8: 玩家离开公会
	if err := guildService.LeaveGuild(playerId); err != nil {
		t.Fatalf("Failed to leave guild: %v", err)
	}

	// 测试9: 检查玩家是否已离开公会
	_, exists = guildService.GetGuildByPlayer(playerId)
	if exists {
		t.Fatal("Player should not be in guild after leaving")
	}

	// 测试10: 解散公会
	if err := guildService.DisbandGuild(guildId); err != nil {
		t.Fatalf("Failed to disband guild: %v", err)
	}

	// 测试11: 检查公会是否已解散
	_, exists = guildService.GetGuild(guildId)
	if exists {
		t.Fatal("Guild should not exist after disbanding")
	}

	// 测试12: 检查会长是否已离开公会
	_, exists = guildService.GetGuildByPlayer(leaderId)
	if exists {
		t.Fatal("Leader should not be in guild after disbanding")
	}

	t.Log("All guild system tests passed!")
}
