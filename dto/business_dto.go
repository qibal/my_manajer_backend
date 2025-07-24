package dto

import "time"

// BusinessCreateRequest merepresentasikan data yang diterima saat membuat bisnis baru.
// Field 'json' digunakan untuk mapping dari body request JSON.
type BusinessCreateRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=100"` // Nama bisnis, wajib, min 3 karakter, max 100
	Avatar string `json:"avatar,omitempty"`                       // URL avatar, opsional
	// OwnerID akan diambil dari token JWT, tidak dari request body
	// Settings juga bisa opsional atau memiliki DTO tersendiri jika lebih kompleks
}

// BusinessUpdateRequest merepresentasikan data yang diterima saat memperbarui bisnis.
// Semua field bersifat opsional karena hanya field yang ada yang akan diupdate.
type BusinessUpdateRequest struct {
	Name   string `json:"name,omitempty" validate:"min=3,max=100"`
	Avatar string `json:"avatar,omitempty"`
	// Settings bisa berupa struct terpisah jika ingin mengizinkan update parsial pada pengaturan
}

/*
Cara Penggunaan:

// Dalam handler Fiber:
// var req dto.BusinessCreateRequest
// if err := c.BodyParser(&req); err != nil {
//     return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
// }
//
// if err := validate.Struct(req); err != nil { // Asumsi ada instance validator
//     return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", err.Error())
// }
*/

// BusinessResponse merepresentasikan data bisnis yang dikirimkan sebagai respons API.
// Field 'json' digunakan untuk mapping ke JSON response.
type BusinessResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	OwnerID   string           `json:"ownerId"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Settings  BusinessSettings `json:"settings"` // Menggunakan BusinessSettings dari model
	Avatar    string           `json:"avatar"`
}

// BusinessSettings digunakan di dalam BusinessResponse.
// Ini adalah duplikasi dari model.BusinessSettings, tetapi bisa disesuaikan
// jika ada kebutuhan untuk menyembunyikan atau mengubah field tertentu untuk response saja.
// Untuk saat ini, sama persis dengan model.BusinessSettings.
type BusinessSettings struct {
	Theme         string `json:"theme"`
	Notifications string `json:"notifications"`
}

/*
Cara Penggunaan:

// Dalam handler Fiber setelah mendapatkan data dari repository:
// businessModel := &model.Business{...} // Data dari database
//
// resp := dto.BusinessResponse{
//     ID:        businessModel.ID,
//     Name:      businessModel.Name,
//     OwnerID:   businessModel.OwnerID,
//     CreatedAt: businessModel.CreatedAt,
//     UpdatedAt: businessModel.UpdatedAt,
//     Settings:  dto.BusinessSettings{ // Mapping settings
//         Theme: businessModel.Settings.Theme,
//         Notifications: businessModel.Settings.Notifications,
//     },
//     Avatar: businessModel.Avatar,
// }
//
// return utils.SendSuccessResponse(c, fiber.StatusOK, "Business data retrieved", resp)
*/
