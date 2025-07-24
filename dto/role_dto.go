package dto

type CreateRoleRequest struct {
	BusinessID  string              `json:"businessId"` // Mengubah menjadi string
	Name        string              `json:"name" validate:"required"`
	Permissions map[string][]string `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string              `json:"name"`
	Permissions map[string][]string `json:"permissions"`
}

type RoleResponse struct {
	ID          string              `json:"id"`         // Mengubah menjadi string
	BusinessID  string              `json:"businessId"` // Mengubah menjadi string
	Name        string              `json:"name"`
	Permissions map[string][]string `json:"permissions"`
}
