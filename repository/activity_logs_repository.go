package repository

import (
	"backend_my_manajer/config"
	"backend_my_manajer/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ActivityLogRepository mendefinisikan antarmuka untuk operasi database log aktivitas.
type ActivityLogRepository interface {
	CreateActivityLog(ctx context.Context, log *model.ActivityLog) error
	GetAllActivityLogs(ctx context.Context) ([]model.ActivityLog, error)
}

type activityLogRepository struct {
	collection *mongo.Collection
}

// NewActivityLogRepository membuat instance baru dari ActivityLogRepository.
func NewActivityLogRepository(dbClient *mongo.Client) ActivityLogRepository {
	collection := config.GetCollection(dbClient, "ActivityLogs")
	return &activityLogRepository{
		collection: collection,
	}
}

// CreateActivityLog menyimpan log aktivitas baru ke database.
func (r *activityLogRepository) CreateActivityLog(ctx context.Context, log *model.ActivityLog) error {
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// GetAllActivityLogs mengambil semua log aktivitas dari database.
func (r *activityLogRepository) GetAllActivityLogs(ctx context.Context) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
