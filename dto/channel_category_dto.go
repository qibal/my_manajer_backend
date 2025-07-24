package dto

import "time"

// ChannelCategoryCreateRequest merepresentasikan data yang diterima saat membuat kategori channel baru.
type ChannelCategoryCreateRequest struct {
	BusinessID string `json:"businessId" validate:"required"` // ID Bisnis induk, wajib
	Name       string `json:"name" validate:"required,min=3,max=100"`
}

// ChannelCategoryUpdateRequest merepresentasikan data yang diterima saat memperbarui kategori channel.
type ChannelCategoryUpdateRequest struct {
	Name string `json:"name,omitempty" validate:"min=3,max=100"`
}

// ChannelCategoryResponse merepresentasikan data kategori channel yang dikirimkan sebagai respons API.
type ChannelCategoryResponse struct {
	ID         string    `json:"id"`
	BusinessID string    `json:"businessId"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
