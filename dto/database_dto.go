package dto

import "time"

// DatabaseColumnRequest merepresentasikan definisi kolom untuk request.
type DatabaseColumnRequest struct {
	ID      string   `json:"id,omitempty"`
	Name    string   `json:"name" validate:"required"`
	Type    string   `json:"type" validate:"required,oneof=date text select boolean number"` // Tipe kolom yang diizinkan
	Options []string `json:"options,omitempty"`                                              // Hanya untuk tipe "select"
	Order   int      `json:"order,omitempty"`
}

// SelectOptionResponse merepresentasikan opsi untuk kolom bertipe "select" untuk response.
type SelectOptionResponse struct {
	ID        string    `json:"id"`
	Value     string    `json:"value"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
}

// DatabaseRowValueRequest merepresentasikan nilai-nilai dalam satu baris data untuk request.
// Menggunakan map[string]interface{} karena nilai bisa bervariasi.
type DatabaseRowValueRequest map[string]interface{}

// DatabaseRowRequest merepresentasikan satu baris data untuk request.
type DatabaseRowRequest struct {
	ID     string                  `json:"id,omitempty"`
	Values DatabaseRowValueRequest `json:"values" validate:"required"`
}

// DatabaseCreateRequest merepresentasikan data yang diterima saat membuat database baru.
type DatabaseCreateRequest struct {
	ChannelID string                  `json:"channelId" validate:"required"` // ID channel tempat database berada
	AuthorID  string                  `json:"authorId" validate:"required"`  // ID user yang membuat database
	Title     string                  `json:"title" validate:"required,min=3"`
	Columns   []DatabaseColumnRequest `json:"columns" validate:"required,min=1,dive"` // Minimal satu kolom, dengan validasi rekursif
}

// DatabaseUpdateRequest merepresentasikan data yang diterima saat memperbarui database.
type DatabaseUpdateRequest struct {
	Title   string                  `json:"title,omitempty" validate:"omitempty,min=3"`
	Columns []DatabaseColumnRequest `json:"columns,omitempty" validate:"omitempty,min=1,dive"`
	Rows    []DatabaseRowRequest    `json:"rows,omitempty" validate:"omitempty,dive"`
}

// DatabaseColumnUpdateRequest merepresentasikan data yang diterima saat memperbarui kolom.
type DatabaseColumnUpdateRequest struct {
	Name    string   `json:"name,omitempty" validate:"omitempty,min=1"`
	Type    string   `json:"type,omitempty" validate:"omitempty,oneof=date text select boolean number"`
	Options []string `json:"options,omitempty"` // Untuk update massal atau ganti tipe
	Order   int      `json:"order,omitempty"`
}

// SelectOptionRequest merepresentasikan data yang diterima saat membuat opsi select baru.
type SelectOptionRequest struct {
	Value string `json:"value" validate:"required"`
	Order int    `json:"order,omitempty"`
}

// SelectOptionUpdateRequest merepresentasikan data yang diterima saat memperbarui opsi select.
type SelectOptionUpdateRequest struct {
	Value string `json:"value,omitempty" validate:"omitempty,min=1"`
	Order int    `json:"order,omitempty"`
}

// DatabaseColumnResponse merepresentasikan definisi kolom untuk response.
type DatabaseColumnResponse struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Options []SelectOptionResponse `json:"options,omitempty"`
	Order   int                    `json:"order"`
}

// DatabaseRowValueResponse merepresentasikan nilai-nilai dalam satu baris data untuk response.
type DatabaseRowValueResponse map[string]interface{}

// DatabaseRowResponse merepresentasikan satu baris data untuk response.
type DatabaseRowResponse struct {
	ID     string                   `json:"id"`
	Values DatabaseRowValueResponse `json:"values"`
}

// DatabaseResponse merepresentasikan data database yang dikirimkan sebagai respons API.
type DatabaseResponse struct {
	ID        string                   `json:"id"`
	ChannelID string                   `json:"channelId"`
	AuthorID  string                   `json:"authorId"`
	Title     string                   `json:"title"`
	Columns   []DatabaseColumnResponse `json:"columns"`
	Rows      []DatabaseRowResponse    `json:"rows"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}
