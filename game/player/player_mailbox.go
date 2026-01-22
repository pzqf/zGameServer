package player

import (
	"sync"

	"go.uber.org/zap"

	"github.com/pzqf/zGameServer/event"
	"github.com/pzqf/zGameServer/game/object/component"
	"github.com/pzqf/zUtil/zMap"
)

// 邮件状态定义
const (
	MailStatusUnread  = 1
	MailStatusRead    = 2
	MailStatusDeleted = 3
)

// Mail 邮件结构
type Mail struct {
	mailId       int64
	senderId     int64
	senderName   string
	receiverId   int64
	receiverName string
	title        string
	content      string
	attachments  *zMap.Map // key: int64(itemId), value: int(count)
	sendTime     int64
	status       int
}

// Mailbox 邮箱系统
type Mailbox struct {
	*component.BaseComponent
	playerId int64
	logger   *zap.Logger
	mails    *zMap.Map // key: int64(mailId), value: *Mail
	maxCount int
	mu       sync.RWMutex // 用于保护邮箱操作的互斥锁
}

func NewMailbox(playerId int64, logger *zap.Logger) *Mailbox {
	return &Mailbox{
		BaseComponent: component.NewBaseComponent("mailbox"),
		playerId:      playerId,
		logger:        logger,
		mails:         zMap.NewMap(),
		maxCount:      100, // 邮箱最大容量
	}
}

func (mb *Mailbox) Init() error {
	// 初始化邮箱系统
	mb.logger.Debug("Initializing mailbox", zap.Int64("playerId", mb.playerId))
	return nil
}

// Destroy 销毁邮箱组件
func (mb *Mailbox) Destroy() {
	// 清理邮箱资源
	mb.logger.Debug("Destroying mailbox", zap.Int64("playerId", mb.playerId))
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.mails.Clear()
}

// SendMail 发送邮件
func (mb *Mailbox) SendMail(mail *Mail) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// 检查邮箱是否已满
	if mb.mails.Len() >= int64(mb.maxCount) {
		// 删除最早的已读邮件
		var oldestReadMailId int64
		var oldestReadMailTime int64
		mb.mails.Range(func(key, value interface{}) bool {
			m := value.(*Mail)
			if m.status == MailStatusRead && (oldestReadMailId == 0 || m.sendTime < oldestReadMailTime) {
				oldestReadMailId = key.(int64)
				oldestReadMailTime = m.sendTime
			}
			return true
		})

		if oldestReadMailId != 0 {
			mb.mails.Delete(oldestReadMailId)
		} else {
			return nil // 邮箱已满，无法发送新邮件
		}
	}

	// 添加邮件
	mb.mails.Store(mail.mailId, mail)
	mb.logger.Info("Mail sent", zap.Int64("mailId", mail.mailId), zap.Int64("senderId", mail.senderId), zap.Int64("receiverId", mail.receiverId))

	// 发布邮件接收事件
	eventData := &event.PlayerMailEventData{
		PlayerID: mb.playerId,
		MailID:   mail.mailId,
	}
	event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerMailReceived, mb, eventData))

	return nil
}

// GetMail 获取邮件
func (mb *Mailbox) GetMail(mailId int64) (*Mail, bool) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mail, exists := mb.mails.Get(mailId)
	if !exists {
		return nil, false
	}

	// 标记为已读
	m := mail.(*Mail)
	if m.status == MailStatusUnread {
		m.status = MailStatusRead
		mb.mails.Store(mailId, m)
	}

	return m, true
}

// GetAllMails 获取所有邮件
func (mb *Mailbox) GetAllMails() []*Mail {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	var mails []*Mail
	mb.mails.Range(func(key, value interface{}) bool {
		if value != nil {
			mails = append(mails, value.(*Mail))
		}
		return true
	})
	return mails
}

// GetUnreadMails 获取未读邮件
func (mb *Mailbox) GetUnreadMails() []*Mail {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	var mails []*Mail
	mb.mails.Range(func(key, value interface{}) bool {
		if value != nil {
			m := value.(*Mail)
			if m.status == MailStatusUnread {
				mails = append(mails, m)
			}
		}
		return true
	})
	return mails
}

// DeleteMail 删除邮件
func (mb *Mailbox) DeleteMail(mailId int64) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	_, exists := mb.mails.Get(mailId)
	if !exists {
		return nil // 邮件不存在
	}

	// 删除邮件
	mb.mails.Delete(mailId)
	mb.logger.Info("Mail deleted", zap.Int64("mailId", mailId), zap.Int64("playerId", mb.playerId))
	return nil
}

// ClaimAttachments 领取邮件附件
func (mb *Mailbox) ClaimAttachments(mailId int64) (*zMap.Map, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mail, exists := mb.mails.Get(mailId)
	if !exists {
		return nil, nil // 邮件不存在
	}

	m := mail.(*Mail)
	if m.status == MailStatusDeleted {
		return nil, nil // 邮件已删除
	}

	// 获取附件
	attachments := m.attachments

	// 清空附件
	m.attachments = zMap.NewMap()
	mb.mails.Store(mailId, m)

	mb.logger.Info("Claimed mail attachments", zap.Int64("mailId", mailId), zap.Int64("playerId", mb.playerId))

	// 发布邮件附件领取事件
	eventData := &event.PlayerMailEventData{
		PlayerID: mb.playerId,
		MailID:   mailId,
	}
	event.GetGlobalEventBus().Publish(event.NewEvent(event.EventPlayerMailClaimed, mb, eventData))

	return attachments, nil
}
