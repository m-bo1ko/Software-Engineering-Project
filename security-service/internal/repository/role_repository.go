package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"security-service/internal/models"
)

// RoleRepository handles role database operations
type RoleRepository struct {
	collection *mongo.Collection
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(collection *mongo.Collection) *RoleRepository {
	return &RoleRepository{collection: collection}
}

// Create inserts a new role into the database
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) (*models.Role, error) {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	if role.Permissions == nil {
		role.Permissions = []models.Permission{}
	}

	result, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("role with this name already exists")
		}
		return nil, err
	}

	role.ID = result.InsertedID.(primitive.ObjectID)
	return role, nil
}

// FindByID retrieves a role by its ID
func (r *RoleRepository) FindByID(ctx context.Context, id string) (*models.Role, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	var role models.Role
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// FindByName retrieves a role by its name
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// FindAll retrieves all roles
func (r *RoleRepository) FindAll(ctx context.Context) ([]*models.Role, error) {
	findOptions := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// FindByNames retrieves multiple roles by their names
func (r *RoleRepository) FindByNames(ctx context.Context, names []string) ([]*models.Role, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"name": bson.M{"$in": names}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// Update updates an existing role by name
func (r *RoleRepository) Update(ctx context.Context, name string, updates bson.M) (*models.Role, error) {
	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"name": name},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var role models.Role
	if err := result.Decode(&role); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// Delete removes a role from the database by name
func (r *RoleRepository) Delete(ctx context.Context, name string) error {
	// First check if it's a system role
	var role models.Role
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("role not found")
		}
		return err
	}

	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("role not found")
	}

	return nil
}

// ExistsByName checks if a role exists with the given name
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": name})
	return count > 0, err
}

// InitializeDefaultRoles creates default system roles if they don't exist
func (r *RoleRepository) InitializeDefaultRoles(ctx context.Context) error {
	defaultRoles := []models.Role{
		{
			Name:        "admin",
			Description: "Administrator with full access",
			IsSystem:    true,
			Permissions: []models.Permission{
				{Resource: "*", Actions: []string{"*"}},
			},
		},
		{
			Name:        "user",
			Description: "Standard user with basic access",
			IsSystem:    true,
			Permissions: []models.Permission{
				{Resource: "profile", Actions: []string{"read", "write"}},
				{Resource: "notifications", Actions: []string{"read", "write"}},
			},
		},
		{
			Name:        "building_manager",
			Description: "Building manager with building and energy access",
			IsSystem:    true,
			Permissions: []models.Permission{
				{Resource: "buildings", Actions: []string{"read", "write"}},
				{Resource: "energy", Actions: []string{"read"}},
				{Resource: "reports", Actions: []string{"read", "write"}},
				{Resource: "alerts", Actions: []string{"read", "write"}},
			},
		},
		{
			Name:        "energy_analyst",
			Description: "Energy analyst with read access to energy data",
			IsSystem:    true,
			Permissions: []models.Permission{
				{Resource: "energy", Actions: []string{"read"}},
				{Resource: "reports", Actions: []string{"read", "write"}},
				{Resource: "buildings", Actions: []string{"read"}},
			},
		},
	}

	for _, role := range defaultRoles {
		exists, err := r.ExistsByName(ctx, role.Name)
		if err != nil {
			return err
		}

		if !exists {
			role.CreatedAt = time.Now()
			role.UpdatedAt = time.Now()
			if _, err := r.collection.InsertOne(ctx, role); err != nil {
				return err
			}
		}
	}

	return nil
}
