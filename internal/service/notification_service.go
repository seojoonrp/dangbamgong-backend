package service

import (
	"context"
	"fmt"
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
	// Send* 는 푸시 알림을 유발하는 fire-and-forget 메서드다.
	// 요청 생명주기와 분리하기 위해 내부에서 dispatch(별도 goroutine + background ctx)로 실행되며,
	// 따라서 ctx를 받지 않고 error도 반환하지 않는다 (실패는 내부에서 로깅).
	SendVoidReminder(userID primitive.ObjectID, hours int)
	SendVoidAutoCancel(userID primitive.ObjectID)
	SendFriendRequest(receiverID primitive.ObjectID, senderNickname string)
	SendFriendAccept(originalSenderID primitive.ObjectID, accepterNickname string)
	SendFriendNudge(targetID primitive.ObjectID, senderNickname string)

	GetNotifications(ctx context.Context, userID string, limit int, offset int) (*dto.NotificationListResponse, error)
	MarkAsRead(ctx context.Context, userID string, notifID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (*dto.UnreadCountResponse, error)
	DeleteNotification(ctx context.Context, userID string, notifID string) error
	DeleteAllRead(ctx context.Context, userID string) error
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

func (s *notificationService) SendVoidReminder(userID primitive.ObjectID, hours int) {
	s.dispatch(func(ctx context.Context) {
		pushEnabled := s.isPushEnabled(ctx, userID, model.NotifVoidReminder)
		s.sendNotification(ctx, userID, model.NotifVoidReminder,
			"당밤공 알림",
			fmt.Sprintf("공백을 시작한 지 %d시간이 지났어요", hours),
			nil, pushEnabled,
		)
	})
}

func (s *notificationService) SendVoidAutoCancel(userID primitive.ObjectID) {
	s.dispatch(func(ctx context.Context) {
		s.sendNotification(ctx, userID, model.NotifVoidAutoCancel,
			"당밤공 알림",
			"새로운 하루가 시작되어 공백이 자동 취소되었어요",
			nil, true,
		)
	})
}

func (s *notificationService) SendFriendRequest(receiverID primitive.ObjectID, senderNickname string) {
	s.dispatch(func(ctx context.Context) {
		pushEnabled := s.isPushEnabled(ctx, receiverID, model.NotifFriendRequest)
		s.sendNotification(ctx, receiverID, model.NotifFriendRequest,
			senderNickname,
			"친구 요청을 보냈어요",
			map[string]string{"senderNickname": senderNickname}, pushEnabled,
		)
	})
}

func (s *notificationService) SendFriendAccept(originalSenderID primitive.ObjectID, accepterNickname string) {
	s.dispatch(func(ctx context.Context) {
		pushEnabled := s.isPushEnabled(ctx, originalSenderID, model.NotifFriendAccept)
		s.sendNotification(ctx, originalSenderID, model.NotifFriendAccept,
			accepterNickname,
			"친구 요청을 수락했어요.",
			map[string]string{"accepterNickname": accepterNickname}, pushEnabled,
		)
	})
}

func (s *notificationService) SendFriendNudge(targetID primitive.ObjectID, senderNickname string) {
	s.dispatch(func(ctx context.Context) {
		pushEnabled := s.isPushEnabled(ctx, targetID, model.NotifFriendNudge)
		s.sendNotification(ctx, targetID, model.NotifFriendNudge,
			senderNickname,
			"알림을 보냈어요",
			map[string]string{"senderNickname": senderNickname}, pushEnabled,
		)
	})
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

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	_, err = s.notifRepo.MarkAllAsRead(ctx, oid)
	if err != nil {
		return domain.NewInternal("failed to mark all as read: " + err.Error())
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

func (s *notificationService) DeleteNotification(ctx context.Context, userID string, notifID string) error {
	userOid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	notifOid, err := primitive.ObjectIDFromHex(notifID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrNotificationNotFound, "invalid notification id")
	}

	deletedCount, err := s.notifRepo.Delete(ctx, notifOid, userOid)
	if err != nil {
		return domain.NewInternal("failed to delete notification: " + err.Error())
	}
	if deletedCount == 0 {
		return domain.NewNotFound(domain.ErrNotificationNotFound, "notification not found")
	}
	return nil
}

func (s *notificationService) DeleteAllRead(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	_, err = s.notifRepo.DeleteAllRead(ctx, oid)
	if err != nil {
		return domain.NewInternal("failed to delete read notifications: " + err.Error())
	}
	return nil
}

func (s *notificationService) dispatch(fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[NOTIF] panic in async dispatch: %v\n", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fn(ctx)
	}()
}
