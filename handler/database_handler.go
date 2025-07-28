package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DatabaseHandler menangani logika terkait entitas Database.
type DatabaseHandler interface {
	CreateDatabase(c *fiber.Ctx) error
	GetDatabaseByID(c *fiber.Ctx) error
	GetDatabasesByChannelID(c *fiber.Ctx) error
	UpdateDatabase(c *fiber.Ctx) error
	DeleteDatabase(c *fiber.Ctx) error

	AddRowToDatabase(c *fiber.Ctx) error
	UpdateRowInDatabase(c *fiber.Ctx) error
	DeleteRowFromDatabase(c *fiber.Ctx) error

	// Handler untuk operasi Kolom
	UpdateColumnInDatabase(c *fiber.Ctx) error
	DeleteColumnFromDatabase(c *fiber.Ctx) error

	// Handler untuk operasi Select Options
	AddSelectOptionToColumn(c *fiber.Ctx) error
	UpdateSelectOptionInColumn(c *fiber.Ctx) error
	DeleteSelectOptionFromColumn(c *fiber.Ctx) error
	GetColumnInDatabase(c *fiber.Ctx) error
	GetSelectOptionInColumn(c *fiber.Ctx) error
	GetRowInDatabase(c *fiber.Ctx) error
	GetRowsByDatabaseID(c *fiber.Ctx) error
}

type databaseHandlerImpl struct {
	dbRepo   repository.DatabaseRepository
	userRepo repository.UserRepository // Untuk otorisasi, mungkin perlu cek peran
}

// NewDatabaseHandler membuat instance baru dari DatabaseHandler.
func NewDatabaseHandler(dbRepo repository.DatabaseRepository, userRepo repository.UserRepository) DatabaseHandler {
	return &databaseHandlerImpl{dbRepo: dbRepo, userRepo: userRepo}
}

// CreateDatabase creates a new database entry.
// @Summary Create a new database
// @Description Create a new custom database for a specific channel.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param database body dto.DatabaseCreateRequest true "Database Creation Details"
// @Success 201 {object} utils.APIResponse{data=dto.DatabaseResponse} "Database created successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases [post]
func (h *databaseHandlerImpl) CreateDatabase(c *fiber.Ctx) error {
	var req dto.DatabaseCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if req.ChannelID == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "ChannelID is required")
	}
	if req.AuthorID == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "AuthorID is required")
	}
	if req.Title == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Title is required")
	}
	if len(req.Title) < 3 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Title must be at least 3 characters long")
	}
	if len(req.Columns) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "At least one column is required")
	}

	var columns []model.DatabaseColumn
	for i, colReq := range req.Columns {
		if colReq.Name == "" {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Name is required")
		}
		if colReq.Type == "" {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Type is required")
		}
		validTypes := map[string]bool{"date": true, "text": true, "select": true, "boolean": true, "number": true}
		if !validTypes[colReq.Type] {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Invalid Column Type. Allowed types: date, text, select, boolean, number")
		}

		var selectOptions []model.SelectOption
		if colReq.Type == "select" {
			if len(colReq.Options) == 0 {
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Options are required for 'select' type")
			}
			for optIdx, optValue := range colReq.Options {
				selectOptions = append(selectOptions, model.SelectOption{
					ID:        primitive.NewObjectID(),
					Value:     optValue,
					Order:     optIdx + 1,
					CreatedAt: time.Now(),
				})
			}
		}

		columns = append(columns, model.DatabaseColumn{
			ID:      primitive.NewObjectID(),
			Name:    colReq.Name,
			Type:    colReq.Type,
			Options: selectOptions,
			Order:   i + 1,
		})
	}

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	if req.AuthorID != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: AuthorID must match authenticated user ID", nil)
	}

	authorID, err := primitive.ObjectIDFromHex(req.AuthorID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid Author ID format", err.Error())
	}
	channelID, err := primitive.ObjectIDFromHex(req.ChannelID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid Channel ID format", err.Error())
	}

	newDatabase := &model.Database{
		ID:        primitive.NewObjectID(),
		ChannelID: channelID,
		AuthorID:  authorID,
		Title:     req.Title,
		DatabaseData: model.DatabaseData{
			Columns: columns,
			Rows:    []model.DatabaseRow{}, // Awalnya kosong
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.dbRepo.CreateDatabase(ctx, newDatabase); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create database", err.Error())
	}

	respColumns := make([]dto.DatabaseColumnResponse, len(newDatabase.DatabaseData.Columns))
	for i, col := range newDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Database created successfully", dto.DatabaseResponse{
		ID:        newDatabase.ID.Hex(),
		ChannelID: newDatabase.ChannelID.Hex(),
		AuthorID:  newDatabase.AuthorID.Hex(),
		Title:     newDatabase.Title,
		Columns:   respColumns,
		Rows:      []dto.DatabaseRowResponse{}, // Awalnya kosong
		CreatedAt: newDatabase.CreatedAt,
		UpdatedAt: newDatabase.UpdatedAt,
	})
}

// GetDatabaseByID retrieves a database by its ID.
// @Summary Get database by ID
// @Description Get a specific custom database by its ID.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Database retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to access this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id} [get]
func (h *databaseHandlerImpl) GetDatabaseByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	database, err := h.dbRepo.GetDatabaseByID(ctx, objectID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get database", err.Error())
	}
	if database == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Pastikan user yang request adalah author dari database atau memiliki peran yang relevan.
	// Saat ini, kita hanya cek apakah user adalah author. Peran lain (misal: admin bisnis, anggota channel) bisa ditambahkan.
	if database.AuthorID.Hex() != userIDStr {
		// TODO: Tambahkan logika otorisasi yang lebih kompleks (cek peran, anggota channel)
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to access this database", nil)
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(database.DatabaseData.Columns))
	for i, col := range database.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(database.DatabaseData.Rows))
	for i, row := range database.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Database retrieved successfully", dto.DatabaseResponse{
		ID:        database.ID.Hex(),
		ChannelID: database.ChannelID.Hex(),
		AuthorID:  database.AuthorID.Hex(),
		Title:     database.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: database.CreatedAt,
		UpdatedAt: database.UpdatedAt,
	})
}

// GetDatabasesByChannelID retrieves all databases associated with a specific channel ID.
// @Summary Get databases by channel ID
// @Description Get a list of custom databases for a given channel ID.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param channelId path string true "Channel ID"
// @Success 200 {object} utils.APIResponse{data=[]dto.DatabaseResponse} "Databases retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid Channel ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to access this channel's databases"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/channel/{channelId} [get]
func (h *databaseHandlerImpl) GetDatabasesByChannelID(c *fiber.Ctx) error {
	channelIDStr := c.Params("channelId")
	channelID, err := primitive.ObjectIDFromHex(channelIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid Channel ID format", err.Error())
	}

	// Ambil UserID dan Roles dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}
	// userRoles, _ := c.Locals("userRoles").(map[string][]string)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// TODO: Implementasi otorisasi yang lebih kompleks di sini.
	// Ini bisa melibatkan:
	// 1. Memastikan `channelID` memang milik bisnis yang user miliki aksesnya.
	// 2. Memastikan user adalah anggota channel tersebut.
	// Untuk saat ini, kita akan melewati otorisasi channel, namun penting untuk ditambahkan.

	databases, err := h.dbRepo.GetDatabasesByChannelID(ctx, channelID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get databases by channel ID", err.Error())
	}

	var respDatabases []dto.DatabaseResponse
	for _, db := range databases {
		respColumns := make([]dto.DatabaseColumnResponse, len(db.DatabaseData.Columns))
		for i, col := range db.DatabaseData.Columns {
			respOptions := make([]dto.SelectOptionResponse, len(col.Options))
			for j, opt := range col.Options {
				respOptions[j] = dto.SelectOptionResponse{
					ID:        opt.ID.Hex(),
					Value:     opt.Value,
					Order:     opt.Order,
					CreatedAt: opt.CreatedAt,
				}
			}
			respColumns[i] = dto.DatabaseColumnResponse{
				ID:      col.ID.Hex(),
				Name:    col.Name,
				Type:    col.Type,
				Options: respOptions,
				Order:   col.Order,
			}
		}

		respRows := make([]dto.DatabaseRowResponse, len(db.DatabaseData.Rows))
		for i, row := range db.DatabaseData.Rows {
			respRows[i] = dto.DatabaseRowResponse{
				ID:     row.ID.Hex(),
				Values: dto.DatabaseRowValueResponse(row.Values),
			}
		}

		respDatabases = append(respDatabases, dto.DatabaseResponse{
			ID:        db.ID.Hex(),
			ChannelID: db.ChannelID.Hex(),
			AuthorID:  db.AuthorID.Hex(),
			Title:     db.Title,
			Columns:   respColumns,
			Rows:      respRows,
			CreatedAt: db.CreatedAt,
			UpdatedAt: db.UpdatedAt,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Databases retrieved successfully", respDatabases)
}

// UpdateDatabase updates an existing database entry.
// @Summary Update a database
// @Description Update an existing custom database by its ID. Only author or authorized users can update.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param database body dto.DatabaseUpdateRequest true "Database Update Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Database updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to update this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id} [put]
func (h *databaseHandlerImpl) UpdateDatabase(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}

	var req dto.DatabaseUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if req.Title != "" && len(req.Title) < 3 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Title must be at least 3 characters long if provided")
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, objectID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for update", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author yang dapat mengupdate
	// TODO: Tambahkan otorisasi lebih kompleks (misal: admin bisnis, anggota channel dengan izin)
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to update this database", nil)
	}

	updateMap := bson.M{
		"$set": bson.M{"updatedAt": time.Now()},
	}
	setMap := updateMap["$set"].(bson.M)

	if req.Title != "" {
		setMap["title"] = req.Title
	}

	// Update Columns
	if req.Columns != nil {
		var newColumns []model.DatabaseColumn
		for i, colReq := range req.Columns {
			if colReq.Name == "" {
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Name is required if Columns are provided")
			}
			if colReq.Type == "" {
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Type is required if Columns are provided")
			}
			validTypes := map[string]bool{"date": true, "text": true, "select": true, "boolean": true, "number": true}
			if !validTypes[colReq.Type] {
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Invalid Column Type. Allowed types: date, text, select, boolean, number")
			}

			// if colReq.Type == "select" && len(colReq.Options) == 0 {
			// 	return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Options are required for 'select' type")
			// }

			colID, err := primitive.ObjectIDFromHex(colReq.ID)
			if err != nil {
				colID = primitive.NewObjectID()
			}

			var selectOptions []model.SelectOption
			if colReq.Type == "select" && len(colReq.Options) > 0 {
				for optIdx, optValue := range colReq.Options {
					selectOptions = append(selectOptions, model.SelectOption{
						ID:        primitive.NewObjectID(), // Akan diganti jika ID opsi disediakan
						Value:     optValue,
						Order:     optIdx + 1,
						CreatedAt: time.Now(),
					})
				}
			}

			newColumns = append(newColumns, model.DatabaseColumn{
				ID:      colID,
				Name:    colReq.Name,
				Type:    colReq.Type,
				Options: selectOptions,
				Order:   i + 1,
			})
		}
		setMap["databaseData.columns"] = newColumns
	}

	// Update Rows (ini akan menimpa seluruh array rows jika disediakan)
	// Biasanya, update rows dilakukan melalui endpoint terpisah (AddRow, UpdateRow, DeleteRow)
	// Jika ini dimaksudkan untuk mengganti seluruh array rows, ini bisa berbahaya.
	// Saya akan asumsikan req.Rows hanya untuk update/replace seluruh array. Jika tidak, perlu endpoint terpisah.
	if req.Rows != nil {
		var newRows []model.DatabaseRow
		for _, rowReq := range req.Rows {
			rowID, err := primitive.ObjectIDFromHex(rowReq.ID) // Asumsikan ID baris dikirim jika sudah ada
			if err != nil {
				rowID = primitive.NewObjectID() // Generate baru jika tidak valid atau tidak ada
			}
			newRows = append(newRows, model.DatabaseRow{
				ID:     rowID,
				Values: model.DatabaseRowValue(rowReq.Values),
			})
		}
		setMap["databaseData.rows"] = newRows
	}

	// Periksa apakah ada data yang akan diupdate selain updatedAt
	// if len(setMap) == 1 && setMap["updatedAt"] != nil { // Hanya updatedAt yang ada
	// return utils.SendErrorResponse(c, fiber.StatusBadRequest, "No valid update data provided", nil)
	// }

	updatedDatabase, err := h.dbRepo.UpdateDatabase(ctx, objectID, updateMap)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update database", err.Error())
	}
	if updatedDatabase == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found after update (unlikely)", nil)
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Database updated successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// DeleteDatabase deletes a database entry.
// @Summary Delete a database
// @Description Delete a custom database by its ID. Only author or authorized users can delete.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Success 200 {object} utils.APIResponse "Database deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to delete this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id} [delete]
func (h *databaseHandlerImpl) DeleteDatabase(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, objectID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for deletion", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author yang dapat menghapus
	// TODO: Tambahkan otorisasi lebih kompleks (misal: admin bisnis, anggota channel dengan izin)
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to delete this database", nil)
	}

	if err := h.dbRepo.DeleteDatabase(ctx, objectID); err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found for deletion", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete database", err.Error())
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Database deleted successfully", nil)
}

// AddRowToDatabase adds a new row to a specific database.
// @Summary Add a row to database
// @Description Add a new data row to an existing custom database.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param row body dto.DatabaseRowRequest true "New Row Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Row added successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to add rows to this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/rows [post]
func (h *databaseHandlerImpl) AddRowToDatabase(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}

	var req dto.DatabaseRowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	// if len(req.Values) == 0 { // Baris ini akan dihapus/dikomentari
	// 	return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Row values are required")
	// }

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for adding row", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat menambahkan baris
	// TODO: Tambahkan otorisasi lebih kompleks (misal: admin bisnis, anggota channel dengan izin)
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to add rows to this database", nil)
	}

	newRow := &model.DatabaseRow{
		Values: model.DatabaseRowValue(req.Values),
	}

	updatedDatabase, err := h.dbRepo.AddRowToDatabase(ctx, databaseID, newRow)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to add row to database", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Row added successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// UpdateRowInDatabase updates a specific row within a database.
// @Summary Update a row in database
// @Description Update an existing data row within a custom database.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param rowId path string true "Row ID"
// @Param row body dto.DatabaseRowRequest true "Updated Row Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Row updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to update rows in this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database or Row not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/rows/{rowId} [put]
func (h *databaseHandlerImpl) UpdateRowInDatabase(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	rowIDStr := c.Params("rowId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	rowID, err := primitive.ObjectIDFromHex(rowIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid row ID format", err.Error())
	}

	var req dto.DatabaseRowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if len(req.Values) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Row values are required")
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for updating row", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat memperbarui baris
	// TODO: Tambahkan otorisasi lebih kompleks (misal: admin bisnis, anggota channel dengan izin)
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to update rows in this database", nil)
	}

	updatedDatabase, err := h.dbRepo.UpdateRowInDatabase(ctx, databaseID, rowID, model.DatabaseRowValue(req.Values))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database or Row not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update row in database", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Row updated successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// DeleteRowFromDatabase deletes a specific row from a database.
// @Summary Delete a row from database
// @Description Delete a specific data row from an existing custom database.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param rowId path string true "Row ID"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Row deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to delete rows from this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database or Row not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/rows/{rowId} [delete]
func (h *databaseHandlerImpl) DeleteRowFromDatabase(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	rowIDStr := c.Params("rowId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	rowID, err := primitive.ObjectIDFromHex(rowIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid row ID format", err.Error())
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for deleting row", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat menghapus baris
	// TODO: Tambahkan otorisasi lebih kompleks (misal: admin bisnis, anggota channel dengan izin)
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to delete rows from this database", nil)
	}

	updatedDatabase, err := h.dbRepo.DeleteRowFromDatabase(ctx, databaseID, rowID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database or Row not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete row from database", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Row deleted successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// UpdateColumnInDatabase memperbarui kolom tertentu dalam database.
// @Summary Update a column in database
// @Description Update an existing column within a custom database.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param columnId path string true "Column ID"
// @Param column body dto.DatabaseColumnUpdateRequest true "Updated Column Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Column updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to update columns in this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database or Column not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/columns/{columnId} [put]
func (h *databaseHandlerImpl) UpdateColumnInDatabase(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	columnIDStr := c.Params("columnId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	columnID, err := primitive.ObjectIDFromHex(columnIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid column ID format", err.Error())
	}

	var req dto.DatabaseColumnUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if req.Name != "" && len(req.Name) < 1 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Column Name cannot be empty")
	}

	if req.Type != "" {
		validTypes := map[string]bool{"date": true, "text": true, "select": true, "boolean": true, "number": true}
		if !validTypes[req.Type] {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Invalid Column Type. Allowed types: date, text, select, boolean, number")
		}
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for updating column", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat memperbarui kolom
	// TODO: Tambahkan otorisasi lebih kompleks
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to update columns in this database", nil)
	}

	// Cari kolom yang akan diperbarui
	var existingColumn *model.DatabaseColumn
	for _, col := range existingDB.DatabaseData.Columns {
		if col.ID == columnID {
			existingColumn = &col
			break
		}
	}

	if existingColumn == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Column not found", nil)
	}

	// Buat map updateData untuk repository
	updateData := bson.M{
		"$set": bson.M{},
	}
	setMap := updateData["$set"].(bson.M)

	if req.Name != "" {
		setMap["databaseData.columns.$.name"] = req.Name
	}
	if req.Type != "" {
		setMap["databaseData.columns.$.type"] = req.Type
	}
	if req.Order != 0 {
		setMap["databaseData.columns.$.order"] = req.Order
	}

	// Jika tipe kolom berubah menjadi 'select' dan ada opsi, atau tipe tetap 'select' dan opsi diperbarui
	if req.Type == "select" || (existingColumn.Type == "select" && req.Options != nil) {
		var selectOptions []model.SelectOption
		for optIdx, optValue := range req.Options {
			selectOptions = append(selectOptions, model.SelectOption{
				ID:        primitive.NewObjectID(), // Akan diganti jika ID opsi disediakan
				Value:     optValue,
				Order:     optIdx + 1,
				CreatedAt: time.Now(),
			})
		}
		setMap["databaseData.columns.$.options"] = selectOptions
	} else if req.Type != "select" && existingColumn.Type == "select" && req.Options == nil {
		// Jika tipe berubah dari 'select' ke non-select, hapus opsi yang ada
		setMap["databaseData.columns.$.options"] = []model.SelectOption{}
	}

	updatedDatabase, err := h.dbRepo.UpdateColumnInDatabase(ctx, databaseID, columnID, updateData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database or Column not found during update", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update column in database", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Column updated successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// DeleteColumnFromDatabase menghapus kolom tertentu dari database.
// @Summary Delete a column from database
// @Description Delete a specific column from an existing custom database.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param columnId path string true "Column ID"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Column deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to delete columns from this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database or Column not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/columns/{columnId} [delete]
func (h *databaseHandlerImpl) DeleteColumnFromDatabase(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	columnIDStr := c.Params("columnId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	columnID, err := primitive.ObjectIDFromHex(columnIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid column ID format", err.Error())
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for deleting column", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat menghapus kolom
	// TODO: Tambahkan otorisasi lebih kompleks
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to delete columns from this database", nil)
	}

	updatedDatabase, err := h.dbRepo.DeleteColumnFromDatabase(ctx, databaseID, columnID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database or Column not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete column from database", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Column deleted successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// AddSelectOptionToColumn menambahkan opsi pilihan baru ke kolom 'select' tertentu dalam database.
// @Summary Add select option to column
// @Description Add a new select option to a specific 'select' type column within a custom database.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param columnId path string true "Column ID"
// @Param option body dto.SelectOptionRequest true "New Select Option Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Select option added successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to modify this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database or Column not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/columns/{columnId}/options [post]
func (h *databaseHandlerImpl) AddSelectOptionToColumn(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	columnIDStr := c.Params("columnId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	columnID, err := primitive.ObjectIDFromHex(columnIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid column ID format", err.Error())
	}

	var req dto.SelectOptionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if req.Value == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "Option Value is required")
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi dan validasi kolom
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for adding select option", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat menambahkan opsi
	// TODO: Tambahkan otorisasi lebih kompleks
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to modify this database", nil)
	}

	// Periksa apakah kolom ada dan bertipe 'select'
	var targetColumn *model.DatabaseColumn
	for _, col := range existingDB.DatabaseData.Columns {
		if col.ID == columnID {
			targetColumn = &col
			break
		}
	}

	if targetColumn == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Column not found", nil)
	}

	if targetColumn.Type != "select" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Bad Request", "Column is not of 'select' type, cannot add options")
	}

	newOption := &model.SelectOption{
		Value: req.Value,
		Order: req.Order,
	}

	updatedDatabase, err := h.dbRepo.AddSelectOptionToColumn(ctx, databaseID, columnID, newOption)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database or Column not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to add select option to column", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Select option added successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// UpdateSelectOptionInColumn memperbarui opsi pilihan tertentu dalam kolom 'select' di database.
// @Summary Update select option in column
// @Description Update an existing select option within a specific 'select' type column in a custom database.
// @Tags Databases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param columnId path string true "Column ID"
// @Param optionId path string true "Option ID"
// @Param option body dto.SelectOptionUpdateRequest true "Updated Select Option Details"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Select option updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to modify this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database, Column, or Option not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/columns/{columnId}/options/{optionId} [put]
func (h *databaseHandlerImpl) UpdateSelectOptionInColumn(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	columnIDStr := c.Params("columnId")
	optionIDStr := c.Params("optionId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	columnID, err := primitive.ObjectIDFromHex(columnIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid column ID format", err.Error())
	}
	optionID, err := primitive.ObjectIDFromHex(optionIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid option ID format", err.Error())
	}

	var req dto.SelectOptionUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi manual
	if req.Value == "" && req.Order == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation error", "At least 'value' or 'order' must be provided for update")
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi dan validasi kolom/opsi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for updating select option", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat memperbarui opsi
	// TODO: Tambahkan otorisasi lebih kompleks
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to modify this database", nil)
	}

	// Periksa apakah kolom ada dan bertipe 'select'
	var targetColumn *model.DatabaseColumn
	for _, col := range existingDB.DatabaseData.Columns {
		if col.ID == columnID {
			targetColumn = &col
			break
		}
	}

	if targetColumn == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Column not found", nil)
	}

	if targetColumn.Type != "select" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Bad Request", "Column is not of 'select' type, cannot update options")
	}

	// Periksa apakah opsi ada di dalam kolom
	var targetOption *model.SelectOption
	for _, opt := range targetColumn.Options {
		if opt.ID == optionID {
			targetOption = &opt
			break
		}
	}

	if targetOption == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Select option not found in column", nil)
	}

	updateData := bson.M{
		"$set": bson.M{},
	}
	setMap := updateData["$set"].(bson.M)

	if req.Value != "" {
		setMap["value"] = req.Value
	}
	if req.Order != 0 {
		setMap["order"] = req.Order
	}

	updatedDatabase, err := h.dbRepo.UpdateSelectOptionInColumn(ctx, databaseID, columnID, optionID, updateData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database, Column, or Option not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update select option in column", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Select option updated successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

// DeleteSelectOptionFromColumn menghapus opsi pilihan tertentu dari kolom 'select' di database.
// @Summary Delete select option from column
// @Description Delete a specific select option from a 'select' type column within a custom database.
// @Tags Databases
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Database ID"
// @Param columnId path string true "Column ID"
// @Param optionId path string true "Option ID"
// @Success 200 {object} utils.APIResponse{data=dto.DatabaseResponse} "Select option deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID format"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Token not provided or invalid"
// @Failure 403 {object} utils.APIResponse "Forbidden - User not authorized to modify this database"
// @Failure 404 {object} utils.APIResponse "Not Found - Database, Column, or Option not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /databases/{id}/columns/{columnId}/options/{optionId} [delete]
func (h *databaseHandlerImpl) DeleteSelectOptionFromColumn(c *fiber.Ctx) error {
	databaseIDStr := c.Params("id")
	columnIDStr := c.Params("columnId")
	optionIDStr := c.Params("optionId")

	databaseID, err := primitive.ObjectIDFromHex(databaseIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid database ID format", err.Error())
	}
	columnID, err := primitive.ObjectIDFromHex(columnIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid column ID format", err.Error())
	}
	optionID, err := primitive.ObjectIDFromHex(optionIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid option ID format", err.Error())
	}

	// Ambil UserID dari Locals (dari JWT Middleware) untuk otorisasi
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID not found in token", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Dapatkan database yang ada untuk otorisasi dan validasi kolom/opsi
	existingDB, err := h.dbRepo.GetDatabaseByID(ctx, databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve database for deleting select option", err.Error())
	}
	if existingDB == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database not found", nil)
	}

	// Otorisasi: Hanya author atau user dengan izin yang dapat menghapus opsi
	// TODO: Tambahkan otorisasi lebih kompleks
	if existingDB.AuthorID.Hex() != userIDStr {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Forbidden: You are not authorized to modify this database", nil)
	}

	// Periksa apakah kolom ada dan bertipe 'select'
	var targetColumn *model.DatabaseColumn
	for _, col := range existingDB.DatabaseData.Columns {
		if col.ID == columnID {
			targetColumn = &col
			break
		}
	}

	if targetColumn == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Column not found", nil)
	}

	if targetColumn.Type != "select" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Bad Request", "Column is not of 'select' type, cannot delete options")
	}

	// Periksa apakah opsi ada di dalam kolom
	var targetOption *model.SelectOption
	for _, opt := range targetColumn.Options {
		if opt.ID == optionID {
			targetOption = &opt
			break
		}
	}

	if targetOption == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Select option not found in column", nil)
	}

	updatedDatabase, err := h.dbRepo.DeleteSelectOptionFromColumn(ctx, databaseID, columnID, optionID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Database, Column, or Option not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete select option from column", err.Error())
	}

	// Konversi ke DatabaseResponse
	respColumns := make([]dto.DatabaseColumnResponse, len(updatedDatabase.DatabaseData.Columns))
	for i, col := range updatedDatabase.DatabaseData.Columns {
		respOptions := make([]dto.SelectOptionResponse, len(col.Options))
		for j, opt := range col.Options {
			respOptions[j] = dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			}
		}
		respColumns[i] = dto.DatabaseColumnResponse{
			ID:      col.ID.Hex(),
			Name:    col.Name,
			Type:    col.Type,
			Options: respOptions,
			Order:   col.Order,
		}
	}

	respRows := make([]dto.DatabaseRowResponse, len(updatedDatabase.DatabaseData.Rows))
	for i, row := range updatedDatabase.DatabaseData.Rows {
		respRows[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Select option deleted successfully", dto.DatabaseResponse{
		ID:        updatedDatabase.ID.Hex(),
		ChannelID: updatedDatabase.ChannelID.Hex(),
		AuthorID:  updatedDatabase.AuthorID.Hex(),
		Title:     updatedDatabase.Title,
		Columns:   respColumns,
		Rows:      respRows,
		CreatedAt: updatedDatabase.CreatedAt,
		UpdatedAt: updatedDatabase.UpdatedAt,
	})
}

func (handler *databaseHandlerImpl) GetColumnInDatabase(c *fiber.Ctx) error {
	databaseID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID database tidak valid", err.Error())
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID kolom tidak valid", err.Error())
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Tidak terautentikasi", err.Error())
	}

	database, err := handler.dbRepo.GetDatabaseByID(c.Context(), databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil database", err.Error())
	}

	if database.AuthorID.Hex() != userID {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Anda tidak memiliki izin untuk mengakses database ini", nil)
	}

	column, err := handler.dbRepo.GetColumnInDatabase(c.Context(), databaseID, columnID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "tidak ditemukan") {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil kolom", err.Error())
	}

	columnResponse := dto.DatabaseColumnResponse{
		ID:    column.ID.Hex(),
		Name:  column.Name,
		Type:  column.Type,
		Order: column.Order,
	}
	if column.Type == "select" && len(column.Options) > 0 {
		for _, opt := range column.Options {
			columnResponse.Options = append(columnResponse.Options, dto.SelectOptionResponse{
				ID:        opt.ID.Hex(),
				Value:     opt.Value,
				Order:     opt.Order,
				CreatedAt: opt.CreatedAt,
			})
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Kolom berhasil diambil", columnResponse)
}

func (handler *databaseHandlerImpl) GetSelectOptionInColumn(c *fiber.Ctx) error {
	databaseID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID database tidak valid", err.Error())
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID kolom tidak valid", err.Error())
	}

	optionID, err := primitive.ObjectIDFromHex(c.Params("optionId"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID opsi pilihan tidak valid", err.Error())
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Tidak terautentikasi", err.Error())
	}

	database, err := handler.dbRepo.GetDatabaseByID(c.Context(), databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil database", err.Error())
	}

	if database.AuthorID.Hex() != userID {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Anda tidak memiliki izin untuk mengakses database ini", nil)
	}

	option, err := handler.dbRepo.GetSelectOptionInColumn(c.Context(), databaseID, columnID, optionID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "tidak ditemukan") {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil opsi pilihan", err.Error())
	}

	optionResponse := dto.SelectOptionResponse{
		ID:        option.ID.Hex(),
		Value:     option.Value,
		Order:     option.Order,
		CreatedAt: option.CreatedAt,
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Opsi pilihan berhasil diambil", optionResponse)
}

func (handler *databaseHandlerImpl) GetRowInDatabase(c *fiber.Ctx) error {
	databaseID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID database tidak valid", err.Error())
	}

	rowID, err := primitive.ObjectIDFromHex(c.Params("rowId"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID baris tidak valid", err.Error())
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Tidak terautentikasi", err.Error())
	}

	database, err := handler.dbRepo.GetDatabaseByID(c.Context(), databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil database", err.Error())
	}

	if database.AuthorID.Hex() != userID {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Anda tidak memiliki izin untuk mengakses database ini", nil)
	}

	row, err := handler.dbRepo.GetRowInDatabase(c.Context(), databaseID, rowID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "tidak ditemukan") {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil baris", err.Error())
	}

	rowResponse := dto.DatabaseRowResponse{
		ID:     row.ID.Hex(),
		Values: dto.DatabaseRowValueResponse(row.Values),
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Baris berhasil diambil", rowResponse)
}

func (handler *databaseHandlerImpl) GetRowsByDatabaseID(c *fiber.Ctx) error {
	databaseID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID database tidak valid", err.Error())
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Tidak terautentikasi", err.Error())
	}

	database, err := handler.dbRepo.GetDatabaseByID(c.Context(), databaseID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil database", err.Error())
	}

	if database.AuthorID.Hex() != userID {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Anda tidak memiliki izin untuk mengakses database ini", nil)
	}

	rows, err := handler.dbRepo.GetRowsByDatabaseID(c.Context(), databaseID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "tidak ditemukan") {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil baris", err.Error())
	}

	rowsResponse := make([]dto.DatabaseRowResponse, len(rows))
	for i, row := range rows {
		rowsResponse[i] = dto.DatabaseRowResponse{
			ID:     row.ID.Hex(),
			Values: dto.DatabaseRowValueResponse(row.Values),
		}
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Baris berhasil diambil", rowsResponse)
}
