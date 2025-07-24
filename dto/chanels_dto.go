package dto

import "time"

// ChannelCreateRequest merepresentasikan data yang diterima saat membuat channel baru.
type ChannelCreateRequest struct {
	BusinessID string `json:"businessId" validate:"required"` // ID Bisnis induk, wajib
	Name       string `json:"name" validate:"required,min=3,max=100"`
	Type       string `json:"type" validate:"required" enums:"messages,voices,drawings,documents,databases,reports"` // e.g., "messages", "documents"
	CategoryID string `json:"categoryId,omitempty"`                                                                  // Opsional
	Order      int    `json:"order,omitempty"`                                                                       // Opsional
}

// ChannelUpdateRequest merepresentasikan data yang diterima saat memperbarui channel.
type ChannelUpdateRequest struct {
	Name       string `json:"name,omitempty" validate:"min=3,max=100"`
	Type       string `json:"type,omitempty" enums:"messages,voices,drawings,documents,databases,reports"`
	CategoryID string `json:"categoryId,omitempty"`
	Order      int    `json:"order,omitempty"`
	// Unread dihapus
}

// ChannelResponse merepresentasikan data channel yang dikirimkan sebagai respons API.
type ChannelResponse struct {
	ID         string    `json:"id"`
	BusinessID string    `json:"businessId"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	CategoryID string    `json:"categoryId"` // Mengubah dari Category
	Order      int       `json:"order"`
	CreatedAt  time.Time `json:"createdAt"`
	// Unread dihapus
}
