package dto

// CreateSuperAdminRequest defines the structure for creating the initial super admin user.
// This endpoint should be used ONLY ONCE to set up the first admin.
type CreateSuperAdminRequest struct {
	Username string `json:"username" example:"superadmin"`
	Email    string `json:"email,omitempty" example:"superadmin@example.com"`
	Password string `json:"password" example:"StrongPassword123!"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/superadmin.png"`
}

// CreateSuperAdminResponse defines the response structure after creating a super admin.
type CreateSuperAdminResponse struct {
	ID       string `json:"id" example:"60d5ec49f8a3c5d6c8e7e1f2"`
	Username string `json:"username" example:"superadmin"`
	Email    string `json:"email" example:"superadmin@example.com"`
}
