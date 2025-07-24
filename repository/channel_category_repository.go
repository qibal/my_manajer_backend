package repository

import (
	"context"
	"time"

	"backend_my_manajer/config"
	"backend_my_manajer/model"
	"backend_my_manajer/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChannelCategoryRepository adalah interface untuk operasi database ChannelCategory.
type ChannelCategoryRepository interface {
	CreateChannelCategory(ctx context.Context, category *model.ChannelCategory) error
	GetChannelCategoryByID(ctx context.Context, id primitive.ObjectID) (*model.ChannelCategory, error)
	GetAllChannelCategories(ctx context.Context) ([]model.ChannelCategory, error)
	UpdateChannelCategory(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.ChannelCategory, error)
	DeleteChannelCategory(ctx context.Context, id primitive.ObjectID) error
	GetChannelCategoriesByBusinessID(ctx context.Context, businessID primitive.ObjectID) ([]model.ChannelCategory, error) // Tambahan
}

// channelCategoryRepositoryImpl adalah implementasi dari ChannelCategoryRepository.
type channelCategoryRepositoryImpl struct {
	collection *mongo.Collection
}

// NewChannelCategoryRepository membuat instance baru dari ChannelCategoryRepository.
func NewChannelCategoryRepository(dbClient *mongo.Client) ChannelCategoryRepository {
	collection := config.GetCollection(dbClient, "ChannelCategories") // Menggunakan fungsi GetCollection dari paket config
	return &channelCategoryRepositoryImpl{
		collection: collection,
	}
}

// CreateChannelCategory menyimpan objek ChannelCategory baru ke database.
func (r *channelCategoryRepositoryImpl) CreateChannelCategory(ctx context.Context, category *model.ChannelCategory) error {
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, category)
	if err != nil {
		utils.LogError(err, "Gagal membuat kategori channel baru di database")
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		category.ID = oid
	} else {
		utils.LogWarning("ID kategori channel yang dimasukkan bukan ObjectID: %v", result.InsertedID)
	}

	utils.LogInfo("Berhasil membuat kategori channel baru: %s", category.ID.Hex())
	return nil
}

// GetChannelCategoryByID mengambil objek ChannelCategory berdasarkan ID.
func (r *channelCategoryRepositoryImpl) GetChannelCategoryByID(ctx context.Context, id primitive.ObjectID) (*model.ChannelCategory, error) {
	var category model.ChannelCategory
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Kategori channel dengan ID %s tidak ditemukan", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil kategori channel berdasarkan ID: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil kategori channel dengan ID: %s", id.Hex())
	return &category, nil
}

// GetAllChannelCategories mengambil semua objek ChannelCategory dari database.
func (r *channelCategoryRepositoryImpl) GetAllChannelCategories(ctx context.Context) ([]model.ChannelCategory, error) {
	var categories []model.ChannelCategory
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua kategori channel")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &categories); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen kategori channel")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil semua kategori channel. Total: %d", len(categories))
	return categories, nil
}

// GetChannelCategoriesByBusinessID mengambil semua kategori channel berdasarkan businessID.
func (r *channelCategoryRepositoryImpl) GetChannelCategoriesByBusinessID(ctx context.Context, businessID primitive.ObjectID) ([]model.ChannelCategory, error) {
	var categories []model.ChannelCategory
	filter := bson.M{"businessId": businessID}
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		utils.LogError(err, "Gagal mengambil kategori channel berdasarkan businessId")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &categories); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen kategori channel by businessId")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil kategori channel by businessId. Total: %d", len(categories))
	return categories, nil
}

// UpdateChannelCategory memperbarui objek ChannelCategory berdasarkan ID.
func (r *channelCategoryRepositoryImpl) UpdateChannelCategory(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.ChannelCategory, error) {
	if setMap, ok := updateData["$set"].(bson.M); ok {
		setMap["updatedAt"] = time.Now()
	} else {
		updateData["$set"] = bson.M{"updatedAt": time.Now()}
	}

	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedCategory model.ChannelCategory
	err := r.collection.FindOneAndUpdate(ctx, filter, updateData, opts).Decode(&updatedCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Kategori channel dengan ID %s tidak ditemukan untuk diperbarui", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil dokumen kategori channel setelah update untuk ID: %s", id.Hex())
		return nil, err
	}

	utils.LogInfo("Berhasil memperbarui kategori channel dengan ID: %s", id.Hex())
	return &updatedCategory, nil
}

// DeleteChannelCategory menghapus objek ChannelCategory berdasarkan ID.
func (r *channelCategoryRepositoryImpl) DeleteChannelCategory(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal menghapus kategori channel dengan ID: %s", id.Hex())
		return err
	}

	if result.DeletedCount == 0 {
		utils.LogWarning("Kategori channel dengan ID %s tidak ditemukan untuk dihapus", id.Hex())
		return mongo.ErrNoDocuments
	}
	utils.LogInfo("Berhasil menghapus kategori channel dengan ID: %s", id.Hex())
	return nil
}
