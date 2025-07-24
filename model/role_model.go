package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// RolePermissions merepresentasikan izin-izin untuk sebuah peran.
// Menggunakan map[string][]string karena key (misal: "channels") memiliki array of string (misal: "create", "read").
type RolePermissions map[string][]string

// Role merepresentasikan struktur dokumen peran di database.
type Role struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                // Mengubah tipe ID
	BusinessID  primitive.ObjectID `bson:"businessId,omitempty" json:"businessId"` // Mengubah tipe BusinessID
	Name        string             `json:"name" bson:"name"`
	Permissions RolePermissions    `json:"permissions" bson:"permissions"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Role
// role := model.Role{
//     ID:         "role_admin",
//     BusinessID: "business_abc",
//     Name:       "Administrator",
//     Permissions: model.RolePermissions{
//         "channels": {"create", "read", "update", "delete"},
//         "users":    {"invite", "kick", "manage_roles"},
//     },
// }
*/
