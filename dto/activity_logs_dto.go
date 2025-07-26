package dto

import "time"

// CreateActivityLogDTO adalah DTO untuk membuat entri log aktivitas baru.
type CreateActivityLogDTO struct {
	UserID     string `json:"user_id"`
	Action     string `json:"action"`
	Method     string `json:"method"`
	Endpoint   string `json:"endpoint"`
	StatusCode int    `json:"status_code"`
	IPAddress  string `json:"ip_address"`
}

// ActivityLogResponse adalah DTO untuk respons log aktivitas.
type ActivityLogResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Action     string    `json:"action"`
	Method     string    `json:"method"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}
