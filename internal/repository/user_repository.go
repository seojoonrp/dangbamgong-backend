package repository

import (
	"context"
	"strings"
	"time"

	"dangbamgong-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	FindBySocial(ctx context.Context, provider model.SocialProvider, socialID string) (*model.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	FindByTag(ctx context.Context, tag string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	UpdateNickname(ctx context.Context, id primitive.ObjectID, nickname string) error
	UpdateSettings(ctx context.Context, id primitive.ObjectID, settings model.NotificationSettings) error
	SetVoidState(ctx context.Context, id primitive.ObjectID, isInVoid bool, startedAt *time.Time, lastVoidEndedAt *time.Time) error
	UpdateAppleRefreshToken(ctx context.Context, id primitive.ObjectID, token string) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	SearchByTagPrefix(ctx context.Context, prefix string, excludeIDs []primitive.ObjectID, limit int) ([]model.User, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.User, error)
	FindUsersInVoid(ctx context.Context) ([]model.User, error)
	CancelAllVoidStates(ctx context.Context) (int64, error)
	UpdateFriendRequestLastReadAt(ctx context.Context, id primitive.ObjectID, t time.Time) error
}

type userRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{coll: db.Collection("users")}
}

func (r *userRepository) FindBySocial(ctx context.Context, provider model.SocialProvider, socialID string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.coll.FindOne(ctx, bson.M{
		"social_provider": provider,
		"social_id":       socialID,
	}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByTag(ctx context.Context, tag string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"tag": tag}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *userRepository) UpdateNickname(ctx context.Context, id primitive.ObjectID, nickname string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"nickname": nickname, "updated_at": time.Now()},
	})
	return err
}

func (r *userRepository) UpdateSettings(ctx context.Context, id primitive.ObjectID, settings model.NotificationSettings) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"notification_settings": settings, "updated_at": time.Now()},
	})
	return err
}

func (r *userRepository) SetVoidState(ctx context.Context, id primitive.ObjectID, isInVoid bool, startedAt *time.Time, lastVoidEndedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fields := bson.M{
		"is_in_void":              isInVoid,
		"current_void_started_at": startedAt,
		"updated_at":              time.Now(),
	}
	if lastVoidEndedAt != nil {
		fields["last_void_ended_at"] = lastVoidEndedAt
	}

	_, err := r.coll.UpdateByID(ctx, id, bson.M{"$set": fields})
	return err
}

func (r *userRepository) UpdateAppleRefreshToken(ctx context.Context, id primitive.ObjectID, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"apple_refresh_token": token, "updated_at": time.Now()},
	})
	return err
}

func (r *userRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *userRepository) SearchByTagPrefix(ctx context.Context, prefix string, excludeIDs []primitive.ObjectID, limit int) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"tag": bson.M{"$regex": "^" + strings.ToUpper(prefix)},
	}
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindUsersInVoid(ctx context.Context) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"is_in_void": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []model.User
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepository) CancelAllVoidStates(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.coll.UpdateMany(ctx,
		bson.M{"is_in_void": true},
		bson.M{"$set": bson.M{
			"is_in_void":              false,
			"current_void_started_at": nil,
			"updated_at":              time.Now(),
		}},
	)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (r *userRepository) UpdateFriendRequestLastReadAt(ctx context.Context, id primitive.ObjectID, t time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// TODO: UpdateByID를 사용해서 friend_request_last_read_at과 updated_at을 $set으로 업데이트하세요.
	// 힌트: UpdateNickname 메서드의 패턴을 참고하세요.
	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"friend_request_last_read_at": t, "updated_at": time.Now()},
	})
	return err
}
