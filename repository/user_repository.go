package repository

import (
	"context"

	"backend_my_manajer/config"
	"backend_my_manajer/model"
	"backend_my_manajer/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository adalah interface untuk operasi database User.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetAllUsers(ctx context.Context) ([]model.User, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.User, error)
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	IsSuperAdminExists(ctx context.Context) (bool, error)
}

// userRepositoryImpl adalah implementasi dari UserRepository.
type userRepositoryImpl struct {
	collection *mongo.Collection
}

// NewUserRepository membuat instance baru dari UserRepository.
func NewUserRepository(dbClient *mongo.Client) UserRepository {
	collection := config.GetCollection(dbClient, "Users")
	return &userRepositoryImpl{collection: collection}
}

// CreateUser menyimpan user baru ke database.
func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		utils.LogError(err, "Gagal membuat user di database")
		return err
	}
	return nil
}

// FindUserByUsername mencari user berdasarkan username.
func (r *userRepositoryImpl) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		utils.LogError(err, "Error finding user by username")
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail mencari user berdasarkan email.
func (r *userRepositoryImpl) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		utils.LogError(err, "Error finding user by email")
		return nil, err
	}
	return &user, nil
}

// FindUserByID mencari user berdasarkan ID.
func (r *userRepositoryImpl) FindUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetAllUsers mengambil semua pengguna dari database.
func (r *userRepositoryImpl) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua pengguna")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen pengguna")
		return nil, err
	}
	return users, nil
}

// UpdateUser memperbarui pengguna berdasarkan ID.
func (r *userRepositoryImpl) UpdateUser(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.User, error) {
	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedUser model.User
	err := r.collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": updateData}, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &updatedUser, nil
}

// DeleteUser menghapus pengguna berdasarkan ID.
func (r *userRepositoryImpl) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		utils.LogError(err, "Gagal menghapus pengguna dengan ID: %s", id.Hex())
		return err
	}
	return nil
}

// IsSuperAdminExists memeriksa apakah sudah ada super admin di database.
func (r *userRepositoryImpl) IsSuperAdminExists(ctx context.Context) (bool, error) {
	// Asumsi super admin memiliki peran 'super_admin' di salah satu bisnis
	// atau ada field khusus `isSuperAdmin`.
	// Untuk sederhana, kita akan cari user dengan username 'superadmin'.
	filter := bson.M{"username": "superadmin"}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
