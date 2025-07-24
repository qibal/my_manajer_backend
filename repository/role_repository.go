package repository

import (
	"context"

	"backend_my_manajer/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoleRepository struct {
	rolesCollection *mongo.Collection
}

func NewRoleRepository(dbClient *mongo.Client) *RoleRepository {
	return &RoleRepository{
		rolesCollection: dbClient.Database("mydatabase").Collection("roles"),
	}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role *model.Role) error {
	role.ID = primitive.NewObjectID() // Generate a new ObjectID
	_, err := r.rolesCollection.InsertOne(ctx, role)
	return err
}

func (r *RoleRepository) GetRoleByID(ctx context.Context, roleID primitive.ObjectID) (*model.Role, error) {
	var role model.Role
	filter := bson.M{"_id": roleID}
	err := r.rolesCollection.FindOne(ctx, filter).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	cursor, err := r.rolesCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) UpdateRole(ctx context.Context, roleID primitive.ObjectID, update *model.Role) error {
	filter := bson.M{"_id": roleID}
	updateDoc := bson.M{"$set": update}
	result, err := r.rolesCollection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments // Role not found to update
	}
	return nil
}

func (r *RoleRepository) DeleteRole(ctx context.Context, roleID primitive.ObjectID) error {
	filter := bson.M{"_id": roleID}
	result, err := r.rolesCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments // Role not found to delete
	}
	return nil
}
