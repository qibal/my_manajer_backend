package repository

import (
	"context"
	"fmt"
	"time"

	"backend_my_manajer/config"
	"backend_my_manajer/model"
	"backend_my_manajer/utils"

	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseRepository adalah interface untuk operasi database entitas Database.
type DatabaseRepository interface {
	CreateDatabase(ctx context.Context, database *model.Database) error
	GetDatabaseByID(ctx context.Context, id primitive.ObjectID) (*model.Database, error)
	GetDatabasesByChannelID(ctx context.Context, channelID primitive.ObjectID) ([]model.Database, error)
	UpdateDatabase(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Database, error)
	DeleteDatabase(ctx context.Context, id primitive.ObjectID) error

	// Operasi untuk baris data (rows)
	AddRowToDatabase(ctx context.Context, databaseID primitive.ObjectID, row *model.DatabaseRow) (*model.Database, error)
	UpdateRowInDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID, updatedValues model.DatabaseRowValue) (*model.Database, error)
	DeleteRowFromDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID) (*model.Database, error)

	// Operasi untuk kolom
	UpdateColumnInDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID, updateData bson.M) (*model.Database, error)
	DeleteColumnFromDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID) (*model.Database, error)
	GetColumnInDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID) (*model.DatabaseColumn, error)

	// Operasi untuk select options dalam kolom
	AddSelectOptionToColumn(ctx context.Context, databaseID, columnID primitive.ObjectID, option *model.SelectOption) (*model.Database, error)
	UpdateSelectOptionInColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID, updateData bson.M) (*model.Database, error)
	DeleteSelectOptionFromColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID) (*model.Database, error)
	GetSelectOptionInColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID) (*model.SelectOption, error)
	GetRowInDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID) (*model.DatabaseRow, error)
	GetRowsByDatabaseID(ctx context.Context, databaseID primitive.ObjectID) ([]model.DatabaseRow, error)
}

// databaseRepositoryImpl adalah implementasi dari DatabaseRepository.
type databaseRepositoryImpl struct {
	collection *mongo.Collection
}

// NewDatabaseRepository membuat instance baru dari DatabaseRepository.
func NewDatabaseRepository(dbClient *mongo.Client) DatabaseRepository {
	collection := config.GetCollection(dbClient, "Databases")
	return &databaseRepositoryImpl{collection: collection}
}

// CreateDatabase menyimpan objek Database baru ke database.
func (r *databaseRepositoryImpl) CreateDatabase(ctx context.Context, database *model.Database) error {
	database.CreatedAt = time.Now()
	database.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, database)
	if err != nil {
		utils.LogError(err, "Gagal membuat database baru di database")
		return err
	}
	utils.LogInfo("Berhasil membuat database baru: %s", database.ID.Hex())
	return nil
}

// GetDatabaseByID mengambil objek Database berdasarkan ID.
func (r *databaseRepositoryImpl) GetDatabaseByID(ctx context.Context, id primitive.ObjectID) (*model.Database, error) {
	var database model.Database
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&database)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s tidak ditemukan", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal mengambil database berdasarkan ID: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil database dengan ID: %s", id.Hex())
	return &database, nil
}

// GetDatabasesByChannelID mengambil semua database berdasarkan ChannelID.
func (r *databaseRepositoryImpl) GetDatabasesByChannelID(ctx context.Context, channelID primitive.ObjectID) ([]model.Database, error) {
	var databases []model.Database
	filter := bson.M{"channelId": channelID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal mengambil database berdasarkan channelId")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &databases); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen database by channelId")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil database by channelId. Total: %d", len(databases))
	return databases, nil
}

// UpdateDatabase memperbarui objek Database berdasarkan ID.
func (r *databaseRepositoryImpl) UpdateDatabase(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Database, error) {
	// Tambahkan updatedAt ke updateData
	updateData["$set"].(bson.M)["updatedAt"] = time.Now()

	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, updateData, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s tidak ditemukan untuk diperbarui", id.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal memperbarui database di database: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil memperbarui database dengan ID: %s", id.Hex())
	return &updatedDatabase, nil
}

// DeleteDatabase menghapus objek Database berdasarkan ID.
func (r *databaseRepositoryImpl) DeleteDatabase(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal menghapus database dengan ID: %s", id.Hex())
		return err
	}
	if result.DeletedCount == 0 {
		utils.LogWarning("Database dengan ID %s tidak ditemukan untuk dihapus", id.Hex())
		return mongo.ErrNoDocuments
	}
	utils.LogInfo("Berhasil menghapus database dengan ID: %s", id.Hex())
	return nil
}

// AddRowToDatabase menambahkan baris baru ke array Rows dalam dokumen database.
func (r *databaseRepositoryImpl) AddRowToDatabase(ctx context.Context, databaseID primitive.ObjectID, row *model.DatabaseRow) (*model.Database, error) {
	row.ID = primitive.NewObjectID()

	filter := bson.M{"_id": databaseID}
	update := bson.M{
		"$push": bson.M{"databaseData.rows": row},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s tidak ditemukan untuk menambahkan baris", databaseID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal menambahkan baris ke database: %s", databaseID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil menambahkan baris ke database ID: %s. Row ID: %s", databaseID.Hex(), row.ID.Hex())
	return &updatedDatabase, nil
}

// UpdateColumnInDatabase memperbarui sebuah kolom tertentu dalam dokumen database.
func (r *databaseRepositoryImpl) UpdateColumnInDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID, updateData bson.M) (*model.Database, error) {
	filter := bson.M{"_id": databaseID, "databaseData.columns._id": columnID}
	update := bson.M{
		"$set": bson.M{
			"databaseData.columns.$": updateData, // Memperbarui seluruh objek kolom yang cocok
			"updatedAt":              time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s atau Kolom ID %s tidak ditemukan untuk diperbarui", databaseID.Hex(), columnID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal memperbarui kolom dalam database: %s, Kolom ID: %s", databaseID.Hex(), columnID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil memperbarui kolom ID: %s di database ID: %s", columnID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}

// DeleteColumnFromDatabase menghapus kolom tertentu dari array Columns dalam dokumen database.
func (r *databaseRepositoryImpl) DeleteColumnFromDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID) (*model.Database, error) {
	filter := bson.M{"_id": databaseID}
	update := bson.M{
		"$pull": bson.M{
			"databaseData.columns": bson.M{"_id": columnID},
		},
		"$set": bson.M{"updatedAt": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s atau Kolom ID %s tidak ditemukan untuk dihapus", databaseID.Hex(), columnID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal menghapus kolom dari database: %s, Kolom ID: %s", databaseID.Hex(), columnID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil menghapus kolom ID: %s dari database ID: %s", columnID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}

// AddSelectOptionToColumn menambahkan opsi baru ke array Options dalam sebuah kolom.
func (r *databaseRepositoryImpl) AddSelectOptionToColumn(ctx context.Context, databaseID, columnID primitive.ObjectID, option *model.SelectOption) (*model.Database, error) {
	option.ID = primitive.NewObjectID()
	option.CreatedAt = time.Now()

	filter := bson.M{"_id": databaseID, "databaseData.columns._id": columnID}
	update := bson.M{
		"$push": bson.M{"databaseData.columns.$.options": option}, // Menambahkan opsi ke kolom yang cocok
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s atau Kolom ID %s tidak ditemukan untuk menambahkan opsi select", databaseID.Hex(), columnID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal menambahkan opsi select ke kolom: %s, Kolom ID: %s", databaseID.Hex(), columnID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil menambahkan opsi select ke kolom ID: %s di database ID: %s", columnID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}

// UpdateSelectOptionInColumn memperbarui opsi tertentu dalam array Options di sebuah kolom.
func (r *databaseRepositoryImpl) UpdateSelectOptionInColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID, updateData bson.M) (*model.Database, error) {
	// Filter untuk menemukan dokumen database dan elemen kolom yang sesuai
	filter := bson.M{
		"_id":                              databaseID,
		"databaseData.columns._id":         columnID,
		"databaseData.columns.options._id": optionID, // Filter untuk opsi spesifik
	}

	// Tambahkan updatedAt ke updateData
	updateData["$set"].(bson.M)["updatedAt"] = time.Now()

	update := bson.M{
		// Menggunakan positional operator $[] untuk memperbarui elemen dalam array of arrays
		// Ini akan memperbarui 'value' dari opsi yang cocok di dalam kolom yang cocok
		"$set": bson.M{
			"databaseData.columns.$[col].options.$[opt].value": updateData["$set"].(bson.M)["value"],
			"databaseData.columns.$[col].options.$[opt].order": updateData["$set"].(bson.M)["order"],
			"updatedAt": time.Now(),
		},
	}
	// ArrayFilters untuk mengidentifikasi kolom dan opsi yang akan diperbarui
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"col._id": columnID},
			bson.M{"opt._id": optionID},
		},
	})

	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s, Kolom ID %s, atau Opsi ID %s tidak ditemukan untuk diperbarui", databaseID.Hex(), columnID.Hex(), optionID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal memperbarui opsi select dalam kolom: %s, Kolom ID: %s, Opsi ID: %s", databaseID.Hex(), columnID.Hex(), optionID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil memperbarui opsi select ID: %s di kolom ID: %s dalam database ID: %s", optionID.Hex(), columnID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}

// DeleteSelectOptionFromColumn menghapus opsi tertentu dari array Options di sebuah kolom.
func (r *databaseRepositoryImpl) DeleteSelectOptionFromColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID) (*model.Database, error) {
	filter := bson.M{
		"_id": databaseID,
		"databaseData.columns": bson.M{
			"$elemMatch": bson.M{
				"_id": columnID,
			},
		},
	}

	update := bson.M{
		"$pull": bson.M{
			"databaseData.columns.$.options": bson.M{"_id": optionID},
		},
	}

	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedDatabase)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus select option dari kolom: %w", err)
	}

	return &updatedDatabase, nil
}

func (repository *databaseRepositoryImpl) GetColumnInDatabase(ctx context.Context, databaseID, columnID primitive.ObjectID) (*model.DatabaseColumn, error) {
	filter := bson.M{
		"_id":                      databaseID,
		"databaseData.columns._id": columnID,
	}

	projection := bson.M{
		"databaseData.columns.$": 1,
	}

	var database model.Database
	err := repository.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&database)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("kolom tidak ditemukan di database dengan ID %s atau database tidak ditemukan dengan ID %s", columnID.Hex(), databaseID.Hex())
		}
		return nil, fmt.Errorf("gagal mengambil kolom dari database: %w", err)
	}

	if len(database.DatabaseData.Columns) == 0 {
		return nil, fmt.Errorf("kolom tidak ditemukan di database dengan ID %s", columnID.Hex())
	}

	return &database.DatabaseData.Columns[0], nil
}

func (repository *databaseRepositoryImpl) GetSelectOptionInColumn(ctx context.Context, databaseID, columnID, optionID primitive.ObjectID) (*model.SelectOption, error) {
	filter := bson.M{
		"_id": databaseID,
	}

	var database model.Database
	err := repository.collection.FindOne(ctx, filter).Decode(&database)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("database dengan ID %s tidak ditemukan", databaseID.Hex())
		}
		return nil, fmt.Errorf("gagal mengambil database: %w", err)
	}

	// Cari kolom yang relevan di dalam dokumen database
	var targetColumn *model.DatabaseColumn
	for _, col := range database.DatabaseData.Columns {
		if col.ID == columnID {
			targetColumn = &col
			break
		}
	}

	if targetColumn == nil {
		return nil, fmt.Errorf("kolom dengan ID %s tidak ditemukan di database ID %s", columnID.Hex(), databaseID.Hex())
	}

	// Cari opsi yang relevan di dalam kolom
	var targetOption *model.SelectOption
	for _, opt := range targetColumn.Options {
		if opt.ID == optionID {
			targetOption = &opt
			break
		}
	}

	if targetOption == nil {
		return nil, fmt.Errorf("opsi pilihan dengan ID %s tidak ditemukan di kolom ID %s", optionID.Hex(), columnID.Hex())
	}

	return targetOption, nil
}

func (repository *databaseRepositoryImpl) GetRowInDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID) (*model.DatabaseRow, error) {
	filter := bson.M{
		"_id":                   databaseID,
		"databaseData.rows._id": rowID,
	}

	projection := bson.M{
		"databaseData.rows.$": 1,
	}

	var database model.Database
	err := repository.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&database)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("baris tidak ditemukan dengan ID %s di database ID %s atau database tidak ditemukan", rowID.Hex(), databaseID.Hex())
		}
		return nil, fmt.Errorf("gagal mengambil baris dari database: %w", err)
	}

	if len(database.DatabaseData.Rows) == 0 {
		return nil, fmt.Errorf("baris tidak ditemukan dengan ID %s di database ID %s", rowID.Hex(), databaseID.Hex())
	}

	return &database.DatabaseData.Rows[0], nil
}

func (repository *databaseRepositoryImpl) GetRowsByDatabaseID(ctx context.Context, databaseID primitive.ObjectID) ([]model.DatabaseRow, error) {
	filter := bson.M{
		"_id": databaseID,
	}

	projection := bson.M{
		"databaseData.rows": 1,
	}

	var database model.Database
	err := repository.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&database)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("database tidak ditemukan dengan ID %s", databaseID.Hex())
		}
		return nil, fmt.Errorf("gagal mengambil baris berdasarkan ID database: %w", err)
	}

	return database.DatabaseData.Rows, nil
}

// UpdateRowInDatabase memperbarui nilai-nilai dari baris tertentu dalam dokumen database.
func (r *databaseRepositoryImpl) UpdateRowInDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID, updatedValues model.DatabaseRowValue) (*model.Database, error) {
	filter := bson.M{"_id": databaseID, "databaseData.rows._id": rowID}
	// Gunakan $set dengan posisi array yang diidentifikasi oleh operator posisi $.
	update := bson.M{
		"$set": bson.M{
			"databaseData.rows.$.values": updatedValues, // Perbarui field values di baris yang cocok
			"updatedAt":                  time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s atau Baris ID %s tidak ditemukan untuk diperbarui", databaseID.Hex(), rowID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal memperbarui baris dalam database: %s, Row ID: %s", databaseID.Hex(), rowID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil memperbarui baris ID: %s di database ID: %s", rowID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}

// DeleteRowFromDatabase menghapus baris tertentu dari array Rows dalam dokumen database.
func (r *databaseRepositoryImpl) DeleteRowFromDatabase(ctx context.Context, databaseID, rowID primitive.ObjectID) (*model.Database, error) {
	filter := bson.M{"_id": databaseID}
	update := bson.M{
		"$pull": bson.M{
			"databaseData.rows": bson.M{"_id": rowID},
		},
		"$set": bson.M{"updatedAt": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDatabase model.Database
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Database dengan ID %s atau Baris ID %s tidak ditemukan untuk dihapus", databaseID.Hex(), rowID.Hex())
			return nil, nil
		}
		utils.LogError(err, "Gagal menghapus baris dari database: %s, Row ID: %s", databaseID.Hex(), rowID.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil menghapus baris ID: %s dari database ID: %s", rowID.Hex(), databaseID.Hex())
	return &updatedDatabase, nil
}
