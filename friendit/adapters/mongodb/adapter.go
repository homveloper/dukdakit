package frienditmongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/homveloper/dukdakit/friendit"
)

// ============================================================================
// MongoDB Adapter Example - 사용자가 참고할 수 있는 구현 예제
// ============================================================================

// MongoAdapter provides MongoDB implementation of repositories
// 이것은 예제 구현입니다. 사용자가 자신의 스키마에 맞게 수정할 수 있습니다.
type MongoAdapter struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoAdapter creates a new MongoDB adapter
func NewMongoAdapter(connectionString, databaseName string) (*MongoAdapter, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoAdapter{
		client:   client,
		database: client.Database(databaseName),
	}, nil
}

// Close closes the MongoDB connection
func (ma *MongoAdapter) Close() error {
	return ma.client.Disconnect(context.Background())
}

// ============================================================================
// User Repository Implementation
// ============================================================================

// MongoUserRepository implements UserRepository for MongoDB with BasicUser
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoDB user repository
func (ma *MongoAdapter) NewMongoUserRepository() *MongoUserRepository {
	return &MongoUserRepository{
		collection: ma.database.Collection("users"),
	}
}

// Create implements UserRepository.Create for BasicUser
func (r *MongoUserRepository) Create(ctx context.Context, user friendit.BasicUser) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID implements UserRepository.GetByID for BasicUser
func (r *MongoUserRepository) GetByID(ctx context.Context, id friendit.UserID) (friendit.BasicUser, error) {
	var user friendit.BasicUser

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, fmt.Errorf("user not found: %s", id)
		}
		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update implements UserRepository.Update for BasicUser
func (r *MongoUserRepository) Update(ctx context.Context, user friendit.BasicUser) error {
	user.UpdatedAt = time.Now()

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete implements UserRepository.Delete for BasicUser
func (r *MongoUserRepository) Delete(ctx context.Context, id friendit.UserID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// GetByIDs implements UserRepository.GetByIDs for BasicUser
func (r *MongoUserRepository) GetByIDs(ctx context.Context, ids []friendit.UserID) ([]friendit.BasicUser, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []friendit.BasicUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// FindByStatus implements UserRepository.FindByStatus for BasicUser
func (r *MongoUserRepository) FindByStatus(ctx context.Context, status string) ([]friendit.BasicUser, error) {
	filter := bson.M{"status": status}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by status: %w", err)
	}
	defer cursor.Close(ctx)

	var users []friendit.BasicUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// Search implements UserRepository.Search for BasicUser
func (r *MongoUserRepository) Search(ctx context.Context, query string, limit int) ([]friendit.BasicUser, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"display_name": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []friendit.BasicUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// UpdateStatus implements UserRepository.UpdateStatus for BasicUser
func (r *MongoUserRepository) UpdateStatus(ctx context.Context, id friendit.UserID, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"last_seen":  time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// GetOnlineUsers implements UserRepository.GetOnlineUsers for BasicUser
func (r *MongoUserRepository) GetOnlineUsers(ctx context.Context) ([]friendit.BasicUser, error) {
	filter := bson.M{"status": "online"}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find online users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []friendit.BasicUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// ============================================================================
// Friendship Repository Implementation
// ============================================================================

// MongoFriendshipRepository implements FriendshipRepository for MongoDB with BasicFriendship
type MongoFriendshipRepository struct {
	collection *mongo.Collection
}

// NewMongoFriendshipRepository creates a new MongoDB friendship repository
func (ma *MongoAdapter) NewMongoFriendshipRepository() *MongoFriendshipRepository {
	return &MongoFriendshipRepository{
		collection: ma.database.Collection("friendships"),
	}
}

// Create implements FriendshipRepository.Create for BasicFriendship
func (r *MongoFriendshipRepository) Create(ctx context.Context, friendship friendit.BasicFriendship) error {
	friendship.CreatedAt = time.Now()
	friendship.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, friendship)
	if err != nil {
		return fmt.Errorf("failed to create friendship: %w", err)
	}
	return nil
}

// GetByID implements FriendshipRepository.GetByID for BasicFriendship
func (r *MongoFriendshipRepository) GetByID(ctx context.Context, id friendit.FriendshipID) (friendit.BasicFriendship, error) {
	var friendship friendit.BasicFriendship

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return friendship, fmt.Errorf("friendship not found: %s", id)
		}
		return friendship, fmt.Errorf("failed to get friendship: %w", err)
	}

	return friendship, nil
}

// GetFriendship implements FriendshipRepository.GetFriendship for BasicFriendship
func (r *MongoFriendshipRepository) GetFriendship(ctx context.Context, user1ID, user2ID friendit.UserID) (friendit.BasicFriendship, error) {
	var friendship friendit.BasicFriendship

	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": user1ID, "user2_id": user2ID},
			{"user1_id": user2ID, "user2_id": user1ID},
		},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return friendship, fmt.Errorf("friendship not found between %s and %s", user1ID, user2ID)
		}
		return friendship, fmt.Errorf("failed to get friendship: %w", err)
	}

	return friendship, nil
}

// GetByUserID implements FriendshipRepository.GetByUserID for BasicFriendship
func (r *MongoFriendshipRepository) GetByUserID(ctx context.Context, userID friendit.UserID) ([]friendit.BasicFriendship, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": userID},
			{"user2_id": userID},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find friendships: %w", err)
	}
	defer cursor.Close(ctx)

	var friendships []friendit.BasicFriendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, fmt.Errorf("failed to decode friendships: %w", err)
	}

	return friendships, nil
}

// Update implements FriendshipRepository.Update for BasicFriendship
func (r *MongoFriendshipRepository) Update(ctx context.Context, friendship friendit.BasicFriendship) error {
	friendship.UpdatedAt = time.Now()

	filter := bson.M{"_id": friendship.ID}
	update := bson.M{"$set": friendship}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update friendship: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("friendship not found: %s", friendship.ID)
	}

	return nil
}

// Delete implements FriendshipRepository.Delete for BasicFriendship
func (r *MongoFriendshipRepository) Delete(ctx context.Context, id friendit.FriendshipID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete friendship: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("friendship not found: %s", id)
	}

	return nil
}

// GetMutualFriends implements FriendshipRepository.GetMutualFriends for BasicFriendship
func (r *MongoFriendshipRepository) GetMutualFriends(ctx context.Context, user1ID, user2ID friendit.UserID) ([]friendit.BasicFriendship, error) {
	// MongoDB aggregation pipeline to find mutual friends
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"user1_id": user1ID},
					{"user2_id": user1ID},
				},
				"status": "active",
			},
		},
		{
			"$lookup": bson.M{
				"from": "friendships",
				"let": bson.M{
					"friend_id": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$user1_id", user1ID}},
							"then": "$user2_id",
							"else": "$user1_id",
						},
					},
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$or": []bson.M{
											{
												"$and": []bson.M{
													{"$eq": []interface{}{"$user1_id", user2ID}},
													{"$eq": []interface{}{"$user2_id", "$$friend_id"}},
												},
											},
											{
												"$and": []bson.M{
													{"$eq": []interface{}{"$user1_id", "$$friend_id"}},
													{"$eq": []interface{}{"$user2_id", user2ID}},
												},
											},
										},
									},
									{"$eq": []interface{}{"$status", "active"}},
								},
							},
						},
					},
				},
				"as": "mutual",
			},
		},
		{
			"$match": bson.M{
				"mutual": bson.M{"$ne": []interface{}{}},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to find mutual friends: %w", err)
	}
	defer cursor.Close(ctx)

	var friendships []friendit.BasicFriendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, fmt.Errorf("failed to decode mutual friends: %w", err)
	}

	return friendships, nil
}

// GetByStatus implements FriendshipRepository.GetByStatus for BasicFriendship
func (r *MongoFriendshipRepository) GetByStatus(ctx context.Context, userID friendit.UserID, status string) ([]friendit.BasicFriendship, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": userID},
			{"user2_id": userID},
		},
		"status": status,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find friendships by status: %w", err)
	}
	defer cursor.Close(ctx)

	var friendships []friendit.BasicFriendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, fmt.Errorf("failed to decode friendships: %w", err)
	}

	return friendships, nil
}

// UpdateStatus implements FriendshipRepository.UpdateStatus for BasicFriendship
func (r *MongoFriendshipRepository) UpdateStatus(ctx context.Context, id friendit.FriendshipID, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update friendship status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("friendship not found: %s", id)
	}

	return nil
}

// GetFriendships implements FriendshipRepository.GetFriendships for BasicFriendship
func (r *MongoFriendshipRepository) GetFriendships(ctx context.Context, userIDs []friendit.UserID) ([]friendit.BasicFriendship, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": bson.M{"$in": userIDs}},
			{"user2_id": bson.M{"$in": userIDs}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find friendships: %w", err)
	}
	defer cursor.Close(ctx)

	var friendships []friendit.BasicFriendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, fmt.Errorf("failed to decode friendships: %w", err)
	}

	return friendships, nil
}

// DeleteByUserID implements FriendshipRepository.DeleteByUserID for BasicFriendship
func (r *MongoFriendshipRepository) DeleteByUserID(ctx context.Context, userID friendit.UserID) error {
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": userID},
			{"user2_id": userID},
		},
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete friendships by user ID: %w", err)
	}

	return nil
}
