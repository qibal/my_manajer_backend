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

// ChannelRepository adalah interface untuk operasi database Channel.
type ChannelRepository interface {
	CreateChannel(ctx context.Context, channel *model.Channel) error
	GetChannelByID(ctx context.Context, id primitive.ObjectID) (*model.Channel, error)
	GetAllChannels(ctx context.Context) ([]model.Channel, error)
	UpdateChannel(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Channel, error)
	DeleteChannel(ctx context.Context, id primitive.ObjectID) error
	GetChannelsByBusinessID(ctx context.Context, businessID primitive.ObjectID) ([]model.Channel, error) // Tambahan
}

// channelRepositoryImpl adalah implementasi dari ChannelRepository.
type channelRepositoryImpl struct {
	collection *mongo.Collection
}

// NewChannelRepository membuat instance baru dari ChannelRepository.
func NewChannelRepository(dbClient *mongo.Client) ChannelRepository {
	collection := config.GetCollection(dbClient, "Channels") // Menggunakan fungsi GetCollection dari paket config
	return &channelRepositoryImpl{
		collection: collection,
	}
}

// CreateChannel menyimpan objek Channel baru ke database.
func (r *channelRepositoryImpl) CreateChannel(ctx context.Context, channel *model.Channel) error {
	channel.CreatedAt = time.Now()
	// Unread dihapus

	result, err := r.collection.InsertOne(ctx, channel)
	if err != nil {
		utils.LogError(err, "Gagal membuat channel baru di database")
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		channel.ID = oid
	} else {
		utils.LogWarning("ID channel yang dimasukkan bukan ObjectID: %v", result.InsertedID)
	}

	utils.LogInfo("Berhasil membuat channel baru: %s", channel.ID.Hex())
	return nil
}

// GetChannelByID mengambil objek Channel berdasarkan ID.
func (r *channelRepositoryImpl) GetChannelByID(ctx context.Context, id primitive.ObjectID) (*model.Channel, error) {
	var channel model.Channel
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&channel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Channel dengan ID %s tidak ditemukan", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil channel berdasarkan ID: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil channel dengan ID: %s", id.Hex())
	return &channel, nil
}

// GetAllChannels mengambil semua objek Channel dari database.
func (r *channelRepositoryImpl) GetAllChannels(ctx context.Context) ([]model.Channel, error) {
	var channels []model.Channel
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "categoryId", Value: 1}, {Key: "order", Value: 1}}))
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua channel")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &channels); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen channel")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil semua channel. Total: %d", len(channels))
	return channels, nil
}

// UpdateChannel memperbarui objek Channel berdasarkan ID.
func (r *channelRepositoryImpl) UpdateChannel(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Channel, error) {
	if setMap, ok := updateData["$set"].(bson.M); ok {
		setMap["updatedAt"] = time.Now()
	} else {
		updateData["$set"] = bson.M{"updatedAt": time.Now()}
	}

	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedChannel model.Channel
	err := r.collection.FindOneAndUpdate(ctx, filter, updateData, opts).Decode(&updatedChannel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Channel dengan ID %s tidak ditemukan untuk diperbarui", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil dokumen channel setelah update untuk ID: %s", id.Hex())
		return nil, err
	}

	utils.LogInfo("Berhasil memperbarui channel dengan ID: %s", id.Hex())
	return &updatedChannel, nil
}

// DeleteChannel menghapus objek Channel berdasarkan ID.
func (r *channelRepositoryImpl) DeleteChannel(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal menghapus channel dengan ID: %s", id.Hex())
		return err
	}

	if result.DeletedCount == 0 {
		utils.LogWarning("Channel dengan ID %s tidak ditemukan untuk dihapus", id.Hex())
		return mongo.ErrNoDocuments
	}
	utils.LogInfo("Berhasil menghapus channel dengan ID: %s", id.Hex())
	return nil
}

// GetChannelsByBusinessID mengambil semua channel berdasarkan businessID.
func (r *channelRepositoryImpl) GetChannelsByBusinessID(ctx context.Context, businessID primitive.ObjectID) ([]model.Channel, error) {
	var channels []model.Channel
	filter := bson.M{"businessId": businessID}
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "categoryId", Value: 1}, {Key: "order", Value: 1}}))
	if err != nil {
		utils.LogError(err, "Gagal mengambil channel berdasarkan businessId")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &channels); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen channel by businessId")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil channel by businessId. Total: %d", len(channels))
	return channels, nil
}
