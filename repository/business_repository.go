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

// BusinessRepository adalah interface untuk operasi database Business.
type BusinessRepository interface {
	CreateBusiness(ctx context.Context, business *model.Business) error
	GetBusinessByID(ctx context.Context, id primitive.ObjectID) (*model.Business, error)
	GetAllBusinesses(ctx context.Context) ([]model.Business, error)
	UpdateBusiness(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Business, error)
	DeleteBusiness(ctx context.Context, id primitive.ObjectID) error
}

// businessRepositoryImpl adalah implementasi dari BusinessRepository.
type businessRepositoryImpl struct {
	collection *mongo.Collection
}

// NewBusinessRepository membuat instance baru dari BusinessRepository.
// dbClient adalah client MongoDB yang sudah terhubung.
func NewBusinessRepository(dbClient *mongo.Client) BusinessRepository {
	collection := config.GetCollection(dbClient, "Businesses") // Menggunakan fungsi GetCollection dari paket config
	return &businessRepositoryImpl{
		collection: collection,
	}
}

// CreateBusiness menyimpan objek Business baru ke database.
func (r *businessRepositoryImpl) CreateBusiness(ctx context.Context, business *model.Business) error {
	// Set CreatedAt dan UpdatedAt saat membuat
	business.CreatedAt = time.Now()
	business.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, business)
	if err != nil {
		utils.LogError(err, "Gagal membuat bisnis baru di database")
		return err
	}

	// Ambil ID yang dihasilkan oleh MongoDB dan tetapkan ke objek bisnis
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		business.ID = oid
	} else {
		utils.LogWarning("ID yang dimasukkan bukan ObjectID: %v", result.InsertedID)
		// Anda mungkin ingin menangani kasus ini sebagai error atau log lebih lanjut
	}

	utils.LogInfo("Berhasil membuat bisnis baru dengan ID: %s", business.ID.Hex())
	return nil
}

// GetBusinessByID mengambil objek Business berdasarkan ID.
func (r *businessRepositoryImpl) GetBusinessByID(ctx context.Context, id primitive.ObjectID) (*model.Business, error) {
	var business model.Business
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&business)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Bisnis dengan ID %s tidak ditemukan", id.Hex())
			return nil, nil // Mengembalikan nil model, nil error jika tidak ditemukan
		}
		utils.LogError(err, "Gagal mengambil bisnis berdasarkan ID: %s", id.Hex())
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil bisnis dengan ID: %s", id.Hex())
	return &business, nil
}

// GetAllBusinesses mengambil semua objek Business dari database.
func (r *businessRepositoryImpl) GetAllBusinesses(ctx context.Context) ([]model.Business, error) {
	var businesses []model.Business
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find())
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua bisnis")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &businesses); err != nil {
		utils.LogError(err, "Gagal mendekode dokumen bisnis")
		return nil, err
	}
	utils.LogInfo("Berhasil mengambil semua bisnis. Total: %d", len(businesses))
	return businesses, nil
}

// UpdateBusiness memperbarui objek Business berdasarkan ID.
func (r *businessRepositoryImpl) UpdateBusiness(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*model.Business, error) {
	// Set UpdatedAt saat memperbarui
	if setMap, ok := updateData["$set"].(bson.M); ok {
		setMap["updatedAt"] = time.Now()
	} else {
		updateData["$set"] = bson.M{"updatedAt": time.Now()}
	}

	filter := bson.M{"_id": id}
	// Mengambil dokumen yang diperbarui setelah update
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedBusiness model.Business
	err := r.collection.FindOneAndUpdate(ctx, filter, updateData, opts).Decode(&updatedBusiness)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Bisnis dengan ID %s tidak ditemukan untuk diperbarui", id.Hex())
			return nil, nil // Mengembalikan nil model, nil error jika tidak ditemukan
		}
		utils.LogError(err, "Gagal mengambil dokumen bisnis setelah update untuk ID: %s", id.Hex())
		return nil, err
	}

	utils.LogInfo("Berhasil memperbarui bisnis dengan ID: %s", id.Hex())
	return &updatedBusiness, nil
}

// DeleteBusiness menghapus objek Business berdasarkan ID.
func (r *businessRepositoryImpl) DeleteBusiness(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.LogError(err, "Gagal menghapus bisnis dengan ID: %s", id.Hex())
		return err
	}

	if result.DeletedCount == 0 {
		utils.LogWarning("Bisnis dengan ID %s tidak ditemukan untuk dihapus", id.Hex())
		return mongo.ErrNoDocuments // Mengembalikan error spesifik jika tidak ada dokumen yang dihapus
	}
	utils.LogInfo("Berhasil menghapus bisnis dengan ID: %s", id.Hex())
	return nil
}

// Cara Penggunaan:

// Dalam main.go atau service layer:
// import "go.mongodb.org/mongo-driver/mongo"
// import "context"

// client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
// if err != nil { log.Fatal(err) }
// defer client.Disconnect(context.TODO())

// repo := repository.NewBusinessRepository(client)
// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// defer cancel()

// // Contoh penggunaan CreateBusiness
// newBusiness := &model.Business{Name: "New Biz", OwnerID: "user_test" /* ... */}
// err = repo.CreateBusiness(ctx, newBusiness)
// if err != nil { fmt.Println("Error creating business:", err) }

// // Contoh penggunaan GetBusinessByID
// fetchedBusiness, err := repo.GetBusinessByID(ctx, newBusiness.ID)
// if err != nil { fmt.Println("Error getting business:", err) }
// if fetchedBusiness != nil { fmt.Println("Fetched Business:", fetchedBusiness.Name) }

// // Contoh penggunaan GetAllBusinesses
// allBusinesses, err := repo.GetAllBusinesses(ctx)
// if err != nil { fmt.Println("Error getting all businesses:", err) }
// fmt.Printf("Total Businesses: %d\n", len(allBusinesses))

// // Contoh penggunaan UpdateBusiness
// updateData := bson.M{"$set": bson.M{"name": "Updated Biz Name"}}
// updatedBiz, err := repo.UpdateBusiness(ctx, newBusiness.ID, updateData)
// if err != nil { fmt.Println("Error updating business:", err) }
// if updatedBiz != nil { fmt.Println("Updated Business:", updatedBiz.Name) }

// // Contoh penggunaan DeleteBusiness
// err = repo.DeleteBusiness(ctx, newBusiness.ID)
// if err != nil { fmt.Println("Error deleting business:", err) }
