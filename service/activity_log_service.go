package service

import (
	"context"
	"time"

	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityLogService adalah antarmuka untuk layanan log aktivitas.
type ActivityLogService interface {
	LogActivity(ctx context.Context, userID, action, method, endpoint string, statusCode int, ipAddress string)
}

type activityLogServiceImpl struct {
	repo repository.ActivityLogRepository
}

// NewActivityLogService membuat instance baru dari ActivityLogService.
func NewActivityLogService(repo repository.ActivityLogRepository) ActivityLogService {
	return &activityLogServiceImpl{repo: repo}
}

// LogActivity membuat entri log aktivitas baru.
func (s *activityLogServiceImpl) LogActivity(ctx context.Context, userID, action, method, endpoint string, statusCode int, ipAddress string) {
	logEntry := &model.ActivityLog{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		Action:     action,
		Method:     method,
		Endpoint:   endpoint,
		StatusCode: statusCode,
		IPAddress:  ipAddress,
		CreatedAt:  time.Now(),
	}

	err := s.repo.CreateActivityLog(ctx, logEntry)
	if err != nil {
		utils.LogError(err, "Gagal membuat log aktivitas dari service")
	}
}
