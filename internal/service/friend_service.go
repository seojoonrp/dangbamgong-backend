package service

import (
	"context"
	"regexp"
	"time"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendService interface {
	Search(ctx context.Context, userID string, tagPrefix string) (*dto.FriendSearchResponse, error)
	GetFriends(ctx context.Context, userID string) (*dto.FriendListResponse, error)
	RemoveFriend(ctx context.Context, userID string, targetID string) error
	GetRequests(ctx context.Context, userID string, requestType string) (interface{}, error)
	SendRequest(ctx context.Context, userID string, req dto.SendFriendRequestRequest) (*dto.SendFriendRequestResponse, error)
	AcceptRequest(ctx context.Context, userID string, requestID string) error
	RejectRequest(ctx context.Context, userID string, requestID string) error
	Nudge(ctx context.Context, userID string, targetID string) error
}

type friendService struct {
	userRepo          repository.UserRepository
	blockRepo         repository.BlockRepository
	friendshipRepo    repository.FriendshipRepository
	friendRequestRepo repository.FriendRequestRepository
}

func NewFriendService(
	ur repository.UserRepository,
	br repository.BlockRepository,
	fr repository.FriendshipRepository,
	frr repository.FriendRequestRepository,
) FriendService {
	return &friendService{
		userRepo:          ur,
		blockRepo:         br,
		friendshipRepo:    fr,
		friendRequestRepo: frr,
	}
}

func (s *friendService) Search(ctx context.Context, userID string, tagPrefix string) (*dto.FriendSearchResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	if tagPrefix == "" {
		return nil, domain.NewBadRequest(domain.ErrBadRequest, "tag prefix is required")
	}

	// 내가 차단한 유저
	myBlocks, err := s.blockRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find blocks: " + err.Error())
	}

	// 나를 차단한 유저
	blockedMe, err := s.blockRepo.FindByBlockedID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find blocks: " + err.Error())
	}

	excludeIDs := []primitive.ObjectID{oid}
	for _, b := range myBlocks {
		excludeIDs = append(excludeIDs, b.BlockedID)
	}
	for _, b := range blockedMe {
		excludeIDs = append(excludeIDs, b.UserID)
	}

	sanitized := regexp.QuoteMeta(tagPrefix)
	users, err := s.userRepo.SearchByTagPrefix(ctx, sanitized, excludeIDs, 20)
	if err != nil {
		return nil, domain.NewInternal("failed to search users: " + err.Error())
	}

	items := make([]dto.FriendSearchItem, len(users))
	for i, u := range users {
		items[i] = dto.FriendSearchItem{
			UserID:   u.ID.Hex(),
			Nickname: u.Nickname,
			Tag:      u.Tag,
		}
	}

	return &dto.FriendSearchResponse{Users: items}, nil
}

func (s *friendService) GetFriends(ctx context.Context, userID string) (*dto.FriendListResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	friendships, err := s.friendshipRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find friendships: " + err.Error())
	}

	if len(friendships) == 0 {
		return &dto.FriendListResponse{Friends: []dto.FriendItem{}}, nil
	}

	friendIDs := make([]primitive.ObjectID, len(friendships))
	for i, f := range friendships {
		friendIDs[i] = f.FriendID
	}

	users, err := s.userRepo.FindByIDs(ctx, friendIDs)
	if err != nil {
		return nil, domain.NewInternal("failed to find users: " + err.Error())
	}

	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	items := make([]dto.FriendItem, 0, len(friendships))
	for _, f := range friendships {
		u, ok := userMap[f.FriendID]
		if !ok {
			continue
		}
		items = append(items, dto.FriendItem{
			UserID:    u.ID.Hex(),
			Nickname:  u.Nickname,
			Tag:       u.Tag,
			IsInVoid:  u.IsInVoid,
			CreatedAt: f.CreatedAt,
		})
	}

	return &dto.FriendListResponse{Friends: items}, nil
}

func (s *friendService) RemoveFriend(ctx context.Context, userID string, targetID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	targetOid, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrNotFriends, "invalid target user id")
	}

	existing, err := s.friendshipRepo.FindOne(ctx, oid, targetOid)
	if err != nil {
		return domain.NewInternal("failed to check friendship: " + err.Error())
	}
	if existing == nil {
		return domain.NewBadRequest(domain.ErrNotFriends, "not friends")
	}

	if err := s.friendshipRepo.DeleteByUserPair(ctx, oid, targetOid); err != nil {
		return domain.NewInternal("failed to delete friendship: " + err.Error())
	}

	return nil
}

func (s *friendService) GetRequests(ctx context.Context, userID string, requestType string) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	switch requestType {
	case "received":
		return s.getReceivedRequests(ctx, oid)
	case "sent":
		return s.getSentRequests(ctx, oid)
	default:
		return nil, domain.NewBadRequest(domain.ErrInvalidRequestType, "type must be 'sent' or 'received'")
	}
}

func (s *friendService) getReceivedRequests(ctx context.Context, userID primitive.ObjectID) (*dto.ReceivedRequestsResponse, error) {
	requests, err := s.friendRequestRepo.FindByReceiverID(ctx, userID, model.FriendRequestPending)
	if err != nil {
		return nil, domain.NewInternal("failed to find requests: " + err.Error())
	}

	if len(requests) == 0 {
		return &dto.ReceivedRequestsResponse{Requests: []dto.ReceivedRequestItem{}}, nil
	}

	senderIDs := make([]primitive.ObjectID, len(requests))
	for i, r := range requests {
		senderIDs[i] = r.SenderID
	}

	users, err := s.userRepo.FindByIDs(ctx, senderIDs)
	if err != nil {
		return nil, domain.NewInternal("failed to find users: " + err.Error())
	}

	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	items := make([]dto.ReceivedRequestItem, 0, len(requests))
	for _, r := range requests {
		u, ok := userMap[r.SenderID]
		if !ok {
			continue
		}
		items = append(items, dto.ReceivedRequestItem{
			RequestID: r.ID.Hex(),
			Sender: dto.FriendSearchItem{
				UserID:   u.ID.Hex(),
				Nickname: u.Nickname,
				Tag:      u.Tag,
			},
			CreatedAt: r.CreatedAt,
		})
	}

	return &dto.ReceivedRequestsResponse{Requests: items}, nil
}

func (s *friendService) getSentRequests(ctx context.Context, userID primitive.ObjectID) (*dto.SentRequestsResponse, error) {
	requests, err := s.friendRequestRepo.FindBySenderID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternal("failed to find requests: " + err.Error())
	}

	if len(requests) == 0 {
		return &dto.SentRequestsResponse{Requests: []dto.SentRequestItem{}}, nil
	}

	receiverIDs := make([]primitive.ObjectID, len(requests))
	for i, r := range requests {
		receiverIDs[i] = r.ReceiverID
	}

	users, err := s.userRepo.FindByIDs(ctx, receiverIDs)
	if err != nil {
		return nil, domain.NewInternal("failed to find users: " + err.Error())
	}

	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	items := make([]dto.SentRequestItem, 0, len(requests))
	for _, r := range requests {
		u, ok := userMap[r.ReceiverID]
		if !ok {
			continue
		}
		items = append(items, dto.SentRequestItem{
			RequestID: r.ID.Hex(),
			Receiver: dto.FriendSearchItem{
				UserID:   u.ID.Hex(),
				Nickname: u.Nickname,
				Tag:      u.Tag,
			},
			Status:    string(r.Status),
			CreatedAt: r.CreatedAt,
		})
	}

	return &dto.SentRequestsResponse{Requests: items}, nil
}

func (s *friendService) SendRequest(ctx context.Context, userID string, req dto.SendFriendRequestRequest) (*dto.SendFriendRequestResponse, error) {
	senderOid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	receiverOid, err := primitive.ObjectIDFromHex(req.ReceiverID)
	if err != nil {
		return nil, domain.NewBadRequest(domain.ErrBadRequest, "invalid receiver id")
	}

	if senderOid == receiverOid {
		return nil, domain.NewBadRequest(domain.ErrBadRequest, "cannot send friend request to yourself")
	}

	// 대상 유저 존재 확인
	receiver, err := s.userRepo.FindByID(ctx, receiverOid)
	if err != nil {
		return nil, domain.NewInternal("failed to find user: " + err.Error())
	}
	if receiver == nil {
		return nil, domain.NewNotFound(domain.ErrUserNotFound, "user not found")
	}

	// 이미 친구인지 확인
	existing, err := s.friendshipRepo.FindOne(ctx, senderOid, receiverOid)
	if err != nil {
		return nil, domain.NewInternal("failed to check friendship: " + err.Error())
	}
	if existing != nil {
		return nil, domain.NewConflict(domain.ErrAlreadyFriends, "already friends")
	}

	// 이미 요청이 있는지 확인 (양방향)
	pending, err := s.friendRequestRepo.FindPending(ctx, senderOid, receiverOid)
	if err != nil {
		return nil, domain.NewInternal("failed to check request: " + err.Error())
	}
	if pending != nil {
		return nil, domain.NewConflict(domain.ErrRequestAlreadySent, "request already sent")
	}

	pendingReverse, err := s.friendRequestRepo.FindPending(ctx, receiverOid, senderOid)
	if err != nil {
		return nil, domain.NewInternal("failed to check request: " + err.Error())
	}
	if pendingReverse != nil {
		return nil, domain.NewConflict(domain.ErrRequestAlreadySent, "request already sent")
	}

	// 차단 확인 (양방향)
	block, err := s.blockRepo.FindOne(ctx, senderOid, receiverOid)
	if err != nil {
		return nil, domain.NewInternal("failed to check block: " + err.Error())
	}
	if block != nil {
		return nil, domain.NewForbidden(domain.ErrBlocked, "blocked")
	}

	blockReverse, err := s.blockRepo.FindOne(ctx, receiverOid, senderOid)
	if err != nil {
		return nil, domain.NewInternal("failed to check block: " + err.Error())
	}
	if blockReverse != nil {
		return nil, domain.NewForbidden(domain.ErrBlocked, "blocked")
	}

	now := time.Now()
	friendReq := &model.FriendRequest{
		SenderID:   senderOid,
		ReceiverID: receiverOid,
		Status:     model.FriendRequestPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.friendRequestRepo.Create(ctx, friendReq); err != nil {
		return nil, domain.NewInternal("failed to create friend request: " + err.Error())
	}

	return &dto.SendFriendRequestResponse{RequestID: friendReq.ID.Hex()}, nil
}

func (s *friendService) AcceptRequest(ctx context.Context, userID string, requestID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	reqOid, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrBadRequest, "invalid request id")
	}

	friendReq, err := s.friendRequestRepo.FindByID(ctx, reqOid)
	if err != nil {
		return domain.NewInternal("failed to find request: " + err.Error())
	}
	if friendReq == nil {
		return domain.NewNotFound(domain.ErrRequestNotFound, "request not found")
	}

	if friendReq.ReceiverID != oid {
		return domain.NewNotFound(domain.ErrRequestNotFound, "request not found")
	}

	if friendReq.Status != model.FriendRequestPending {
		return domain.NewBadRequest(domain.ErrRequestNotPending, "request is not pending")
	}

	if err := s.friendRequestRepo.UpdateStatus(ctx, reqOid, model.FriendRequestAccepted); err != nil {
		return domain.NewInternal("failed to update request status: " + err.Error())
	}

	now := time.Now()
	f1 := &model.Friendship{
		UserID:    friendReq.SenderID,
		FriendID:  friendReq.ReceiverID,
		CreatedAt: now,
	}
	f2 := &model.Friendship{
		UserID:    friendReq.ReceiverID,
		FriendID:  friendReq.SenderID,
		CreatedAt: now,
	}

	if err := s.friendshipRepo.Create(ctx, f1); err != nil {
		return domain.NewInternal("failed to create friendship: " + err.Error())
	}
	if err := s.friendshipRepo.Create(ctx, f2); err != nil {
		return domain.NewInternal("failed to create friendship: " + err.Error())
	}

	return nil
}

func (s *friendService) RejectRequest(ctx context.Context, userID string, requestID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	reqOid, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrBadRequest, "invalid request id")
	}

	friendReq, err := s.friendRequestRepo.FindByID(ctx, reqOid)
	if err != nil {
		return domain.NewInternal("failed to find request: " + err.Error())
	}
	if friendReq == nil {
		return domain.NewNotFound(domain.ErrRequestNotFound, "request not found")
	}

	if friendReq.ReceiverID != oid {
		return domain.NewNotFound(domain.ErrRequestNotFound, "request not found")
	}

	if friendReq.Status != model.FriendRequestPending {
		return domain.NewBadRequest(domain.ErrRequestNotPending, "request is not pending")
	}

	if err := s.friendRequestRepo.UpdateStatus(ctx, reqOid, model.FriendRequestRejected); err != nil {
		return domain.NewInternal("failed to update request status: " + err.Error())
	}

	return nil
}

func (s *friendService) Nudge(ctx context.Context, userID string, targetID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	targetOid, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrNotFriends, "invalid target user id")
	}

	existing, err := s.friendshipRepo.FindOne(ctx, oid, targetOid)
	if err != nil {
		return domain.NewInternal("failed to check friendship: " + err.Error())
	}
	if existing == nil {
		return domain.NewBadRequest(domain.ErrNotFriends, "not friends")
	}

	target, err := s.userRepo.FindByID(ctx, targetOid)
	if err != nil {
		return domain.NewInternal("failed to find user: " + err.Error())
	}
	if target == nil {
		return domain.NewNotFound(domain.ErrUserNotFound, "user not found")
	}

	if !target.IsInVoid {
		return domain.NewBadRequest(domain.ErrFriendNotInVoid, "friend is not in void")
	}

	// TODO: 실제 푸시 알림 전송

	return nil
}
