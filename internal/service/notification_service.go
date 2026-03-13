package service

import (
	"context"
	"log"
	"time"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/push"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService interface {
	SendVoidReminder(ctx context.Context, userID primitive.ObjectID) error
	SendFriendRequest(ctx context.Context, receiverID primitive.ObjectID, senderNickname string) error
	SendFriendAccept(ctx context.Context, originalSenderID primitive.ObjectID, accepterNickname string) error
	SendFriendNudge(ctx context.Context, targetID primitive.ObjectID, senderNickname string) error

	GetNotifications(ctx context.Context, userID string, limit int, offset int) (*dto.NotificationListResponse, error)
	MarkAsRead(ctx context.Context, userID string, notifID string) error
	GetUnreadCount(ctx context.Context, userID string) (*dto.UnreadCountResponse, error)
}

type notificationService struct {
	notifRepo       repository.NotificationRepository
	deviceTokenRepo repository.DeviceTokenRepository
	userRepo        repository.UserRepository
	pushClient      push.PushClient
}

func NewNotificationService(
	nr repository.NotificationRepository,
	dr repository.DeviceTokenRepository,
	ur repository.UserRepository,
	pc push.PushClient,
) NotificationService {
	return &notificationService{
		notifRepo:       nr,
		deviceTokenRepo: dr,
		userRepo:        ur,
		pushClient:      pc,
	}
}

func (s *notificationService) sendNotification(ctx context.Context, userID primitive.ObjectID, notifType model.NotificationType, title string, body string, data map[string]string, pushEnabled bool) {
	notif := &model.Notification{
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		Data:      data,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		log.Printf("[NOTIF] failed to create notification: %v\n", err)
		return
	}

	if !pushEnabled {
		return
	}

	tokens, err := s.deviceTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("[NOTIF] failed to find device tokens: %v\n", err)
		return
	}

	for _, t := range tokens {
		if err := s.pushClient.Send(ctx, t.Token, title, body, data); err != nil {
			log.Printf("[NOTIF] failed to send push to %s: %v\n", t.Token, err)
		}
	}
}

func (s *notificationService) isPushEnabled(ctx context.Context, userID primitive.ObjectID, notifType model.NotificationType) bool {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return false
	}

	switch notifType {
	case model.NotifVoidReminder:
		return user.NotificationSettings.VoidReminder
	case model.NotifFriendRequest, model.NotifFriendAccept:
		return user.NotificationSettings.FriendRequest
	case model.NotifFriendNudge:
		return user.NotificationSettings.FriendNudge
	default:
		return false
	}
}

func (s *notificationService) SendVoidReminder(ctx context.Context, userID primitive.ObjectID) error {
	pushEnabled := s.isPushEnabled(ctx, userID, model.NotifVoidReminder)
	s.sendNotification(ctx, userID, model.NotifVoidReminder,
		"오랜 공백 알림",
		"설정한 시간이 지났어요. 공백을 확인해보세요.",
		nil, pushEnabled,
	)
	return nil
}

func (s *notificationService) SendFriendRequest(ctx context.Context, receiverID primitive.ObjectID, senderNickname string) error {
	pushEnabled := s.isPushEnabled(ctx, receiverID, model.NotifFriendRequest)
	s.sendNotification(ctx, receiverID, model.NotifFriendRequest,
		"친구 요청",
		senderNickname+"님이 친구 요청을 보냈어요.",
		map[string]string{"senderNickname": senderNickname}, pushEnabled,
	)
	return nil
}

func (s *notificationService) SendFriendAccept(ctx context.Context, originalSenderID primitive.ObjectID, accepterNickname string) error {
	pushEnabled := s.isPushEnabled(ctx, originalSenderID, model.NotifFriendAccept)
	s.sendNotification(ctx, originalSenderID, model.NotifFriendAccept,
		"친구 수락",
		accepterNickname+"님이 친구 요청을 수락했어요.",
		map[string]string{"accepterNickname": accepterNickname}, pushEnabled,
	)
	return nil
}

func (s *notificationService) SendFriendNudge(ctx context.Context, targetID primitive.ObjectID, senderNickname string) error {
	pushEnabled := s.isPushEnabled(ctx, targetID, model.NotifFriendNudge)
	s.sendNotification(ctx, targetID, model.NotifFriendNudge,
		"친구 알림",
		senderNickname+"님이 알림을 보냈어요.",
		map[string]string{"senderNickname": senderNickname}, pushEnabled,
	)
	return nil
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string, limit int, offset int) (*dto.NotificationListResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	if limit <= 0 {
		limit = 20
	}

	notifications, err := s.notifRepo.FindByUserID(ctx, oid, limit+1, offset)
	if err != nil {
		return nil, domain.NewInternal("failed to find notifications: " + err.Error())
	}

	hasMore := len(notifications) > limit
	if hasMore {
		notifications = notifications[:limit]
	}

	items := make([]dto.NotificationItem, len(notifications))
	for i, n := range notifications {
		items[i] = dto.NotificationItem{
			ID:        n.ID.Hex(),
			Type:      string(n.Type),
			Title:     n.Title,
			Body:      n.Body,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
	}

	return &dto.NotificationListResponse{
		Notifications: items,
		HasMore:       hasMore,
	}, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID string, notifID string) error {
	userOid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	notifOid, err := primitive.ObjectIDFromHex(notifID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrNotificationNotFound, "invalid notification id")
	}

	modifiedCount, err := s.notifRepo.MarkAsRead(ctx, notifOid, userOid)
	if err != nil {
		return domain.NewInternal("failed to mark as read: " + err.Error())
	}
	if modifiedCount == 0 {
		return domain.NewBadRequest(domain.ErrNotificationNotFound, "notification not found")
	}
	return nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (*dto.UnreadCountResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	count, err := s.notifRepo.CountUnread(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to count unread: " + err.Error())
	}

	return &dto.UnreadCountResponse{Count: count}, nil
}
