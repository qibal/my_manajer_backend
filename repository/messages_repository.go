package repository

import (
	"context"
	"fmt"
	"time"

	"backend_my_manajer/config"
	"backend_my_manajer/model"
	"backend_my_manajer/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MessageRepository adalah interface untuk operasi database Message.
type MessageRepository interface {
	CreateMessage(ctx context.Context, message *model.Message) error
	GetMessageByID(ctx context.Context, id primitive.ObjectID) (*model.Message, error)
	GetMessagesByChannelID(ctx context.Context, channelID primitive.ObjectID, limit, skip int64) ([]model.Message, error)
	UpdateMessage(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Message, error)
	DeleteMessage(ctx context.Context, id primitive.ObjectID) error
	AddMessageReaction(ctx context.Context, messageID, userID primitive.ObjectID, emoji string) (*model.Message, error)
	RemoveMessageReaction(ctx context.Context, messageID, userID primitive.ObjectID, emoji string) (*model.Message, error)
}

// messageRepositoryImpl adalah implementasi dari MessageRepository.
type messageRepositoryImpl struct {
	collection *mongo.Collection
}

// NewMessageRepository membuat instance baru dari MessageRepository.
func NewMessageRepository(dbClient *mongo.Client) MessageRepository {
	collection := config.GetCollection(dbClient, "Messages")
	return &messageRepositoryImpl{
		collection: collection,
	}
}

// CreateMessage menyimpan objek Message baru ke database.
func (r *messageRepositoryImpl) CreateMessage(ctx context.Context, message *model.Message) error {
	message.CreatedAt = time.Now()
	message.IsPinned = false                      // Default
	message.Reactions = []model.MessageReaction{} // Default

	result, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		utils.LogError(err, "Gagal membuat pesan baru di database")
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		message.ID = oid
	} else {
		utils.LogWarning("ID pesan yang dimasukkan bukan ObjectID: %v", result.InsertedID)
	}

	utils.LogInfo("Berhasil membuat pesan baru: %s", message.ID.Hex())
	return nil
}

// GetMessageByID mengambil objek Message berdasarkan ID.
func (r *messageRepositoryImpl) GetMessageByID(ctx context.Context, id primitive.ObjectID) (*model.Message, error) {
	var message model.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Pesan dengan ID %s tidak ditemukan", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil pesan berdasarkan ID: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil pesan dengan ID: %s", id.Hex())
	return &message, nil
}

// GetMessagesByChannelID mengambil semua pesan berdasarkan ChannelID.
func (r *messageRepositoryImpl) GetMessagesByChannelID(ctx context.Context, channelID primitive.ObjectID, limit, skip int64) ([]model.Message, error) {
	var messages []model.Message
	filter := bson.M{"channelId": channelID}

	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Pesan terbaru di atas
	if limit > 0 {
		findOptions.SetLimit(limit)
	}
	if skip > 0 {
		findOptions.SetSkip(skip)
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		utils.LogError(err, "Gagal mengambil pesan berdasarkan channelId")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &messages); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen pesan by channelId")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil pesan by channelId. Total: %d", len(messages))
	return messages, nil
}

// UpdateMessage memperbarui objek Message berdasarkan ID.
func (r *messageRepositoryImpl) UpdateMessage(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Message, error) {
	now := time.Now()
	if setMap, ok := updateData["$set"].(bson.M); ok {
		setMap["updatedAt"] = now
	} else {
		updateData["$set"] = bson.M{"updatedAt": now}
	}

	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedMessage model.Message
	err := r.collection.FindOneAndUpdate(ctx, filter, updateData, opts).Decode(&updatedMessage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Pesan dengan ID %s tidak ditemukan untuk diperbarui", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil dokumen pesan setelah update untuk ID: %s", id.Hex())
		return nil, err
	}

	utils.LogInfo("Berhasil memperbarui pesan dengan ID: %s", id.Hex())
	return &updatedMessage, nil
}

// DeleteMessage menghapus objek Message berdasarkan ID.
func (r *messageRepositoryImpl) DeleteMessage(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal menghapus pesan dengan ID: %s", id.Hex())
		return err
	}

	if result.DeletedCount == 0 {
		utils.LogWarning("Pesan dengan ID %s tidak ditemukan untuk dihapus", id.Hex())
		return mongo.ErrNoDocuments
	}
	utils.LogInfo("Berhasil menghapus pesan dengan ID: %s", id.Hex())
	return nil
}

// AddMessageReaction menambahkan reaksi ke pesan.
func (r *messageRepositoryImpl) AddMessageReaction(ctx context.Context, messageID, userID primitive.ObjectID, emoji string) (*model.Message, error) {
	// Filter pesan berdasarkan ID
	filter := bson.M{"_id": messageID}

	// Menyiapkan update: menambahkan reaksi baru jika belum ada, atau menambahkan userID ke reaksi yang sudah ada.
	update := bson.M{
		"$addToSet": bson.M{
			"reactions": bson.M{
				"emoji":   emoji,
				"userIds": userID,
			},
		},
	}

	// Jika reaksi dengan emoji yang sama sudah ada, hanya tambahkan userID ke array userIds
	// Ini membutuhkan operasi array yang lebih kompleks atau dua langkah.
	// Untuk saat ini, kita akan melakukan ini dalam dua langkah atau menggunakan $push/$each.
	// Untuk menyederhanakan, kita asumsikan $addToSet akan bekerja untuk array of structs.
	// Cara yang lebih akurat adalah mencari dulu, lalu update.

	// Cek apakah reaksi sudah ada untuk emoji tertentu
	var existingMessage model.Message
	err := r.collection.FindOne(ctx, filter).Decode(&existingMessage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Pesan dengan ID %s tidak ditemukan untuk menambahkan reaksi", messageID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil pesan untuk menambahkan reaksi: %s", messageID.Hex())
		return nil, err
	}

	reactionFound := false
	for i, reaction := range existingMessage.Reactions {
		if reaction.Emoji == emoji {
			// Cek apakah userID sudah ada di dalam userIds untuk emoji ini
			userAlreadyReacted := false
			for _, existingUserID := range reaction.UserIDs {
				if existingUserID == userID {
					userAlreadyReacted = true
					break
				}
			}
			if !userAlreadyReacted {
				// Tambahkan userID ke array jika belum ada
				update = bson.M{
					"$addToSet": bson.M{
						fmt.Sprintf("reactions.%d.userIds", i): userID,
					},
				}
			}
			reactionFound = true
			break
		}
	}

	if !reactionFound {
		// Jika emoji belum ada, tambahkan reaksi baru
		newReaction := model.MessageReaction{
			Emoji:   emoji,
			UserIDs: []primitive.ObjectID{userID},
		}
		update = bson.M{
			"$push": bson.M{
				"reactions": newReaction,
			},
		}
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedMessage model.Message
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedMessage)
	if err != nil {
		utils.LogError(err, "Gagal memperbarui reaksi pesan di database: %s", messageID.Hex())
		return nil, err
	}

	utils.LogInfo("Berhasil menambahkan reaksi '%s' ke pesan ID: %s oleh user ID: %s", emoji, messageID.Hex(), userID.Hex())
	return &updatedMessage, nil
}

// RemoveMessageReaction menghapus reaksi dari pesan.
func (r *messageRepositoryImpl) RemoveMessageReaction(ctx context.Context, messageID, userID primitive.ObjectID, emoji string) (*model.Message, error) {
	filter := bson.M{"_id": messageID, "reactions.emoji": emoji}

	// Menghapus userID dari array userIds dalam reaksi yang cocok
	update := bson.M{
		"$pull": bson.M{
			"reactions.$[elem].userIds": userID,
		},
	}
	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem.emoji": emoji}},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetArrayFilters(arrayFilters)
	var updatedMessage model.Message
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedMessage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Pesan dengan ID %s atau reaksi '%s' tidak ditemukan untuk dihapus", messageID.Hex(), emoji)
			return nil, nil
		}
		utils.LogError(err, "Gagal menghapus reaksi pesan di database: %s", messageID.Hex())
		return nil, err
	}

	// Opsional: Jika setelah penghapusan, array userIds kosong, hapus seluruh objek reaksi tersebut.
	// Ini bisa dilakukan dengan operasi terpisah atau aggregation pipeline jika lebih kompleks.
	_, err = r.collection.UpdateOne(ctx, bson.M{
		"_id":               messageID,
		"reactions.userIds": bson.M{"$size": 0},
	}, bson.M{
		"$pull": bson.M{
			"reactions": bson.M{
				"userIds": bson.M{"$size": 0},
			},
		},
	})
	if err != nil {
		utils.LogError(err, "Gagal menghapus objek reaksi kosong setelah user dihapus: %s", messageID.Hex())
		// Tidak mengembalikan error fatal karena pesan sudah terupdate dengan baik.
	}

	utils.LogInfo("Berhasil menghapus reaksi '%s' dari pesan ID: %s oleh user ID: %s", emoji, messageID.Hex(), userID.Hex())
	return &updatedMessage, nil
}
