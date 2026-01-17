package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"security-service/internal/models"
)

// TestRoleModel tests the Role model
func TestRoleModel(t *testing.T) {
	t.Run("Create role with permissions", func(t *testing.T) {
		role := models.Role{
			Name:        "building_manager",
			Description: "Building manager with access to building data",
			Permissions: []models.Permission{
				{Resource: "buildings", Actions: []string{"read", "write"}},
				{Resource: "energy", Actions: []string{"read"}},
			},
			IsSystem:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.Equal(t, "building_manager", role.Name)
		assert.Len(t, role.Permissions, 2)
		assert.False(t, role.IsSystem)
	})

	t.Run("HasPermission check", func(t *testing.T) {
		role := models.Role{
			Permissions: []models.Permission{
				{Resource: "buildings", Actions: []string{"read", "write"}},
				{Resource: "energy", Actions: []string{"read"}},
			},
		}

		// Should have these permissions
		assert.True(t, role.HasPermission("buildings", "read"))
		assert.True(t, role.HasPermission("buildings", "write"))
		assert.True(t, role.HasPermission("energy", "read"))

		// Should not have these permissions
		assert.False(t, role.HasPermission("energy", "write"))
		assert.False(t, role.HasPermission("users", "read"))
	})

	t.Run("Wildcard permission", func(t *testing.T) {
		adminRole := models.Role{
			Permissions: []models.Permission{
				{Resource: "*", Actions: []string{"*"}},
			},
		}

		// Admin should have all permissions
		assert.True(t, adminRole.HasPermission("buildings", "read"))
		assert.True(t, adminRole.HasPermission("users", "delete"))
		assert.True(t, adminRole.HasPermission("any_resource", "any_action"))
	})

	t.Run("Wildcard actions only", func(t *testing.T) {
		role := models.Role{
			Permissions: []models.Permission{
				{Resource: "buildings", Actions: []string{"*"}},
			},
		}

		// Should have all actions on buildings
		assert.True(t, role.HasPermission("buildings", "read"))
		assert.True(t, role.HasPermission("buildings", "write"))
		assert.True(t, role.HasPermission("buildings", "delete"))

		// Should not have permissions on other resources
		assert.False(t, role.HasPermission("users", "read"))
	})
}

// TestRoleToResponse tests the ToResponse method
func TestRoleToResponse(t *testing.T) {
	role := models.Role{
		Name:        "test_role",
		Description: "Test role description",
		Permissions: []models.Permission{
			{Resource: "test", Actions: []string{"read"}},
		},
		IsSystem:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	role.ID = [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	resp := role.ToResponse()

	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, role.Name, resp.Name)
	assert.Equal(t, role.Description, resp.Description)
	assert.Equal(t, role.IsSystem, resp.IsSystem)
	assert.Equal(t, len(role.Permissions), len(resp.Permissions))
}

// TestRoleCreateRequest tests role creation request
func TestRoleCreateRequest(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		req := models.RoleCreateRequest{
			Name:        "new_role",
			Description: "A new custom role",
			Permissions: []models.Permission{
				{Resource: "reports", Actions: []string{"read", "write"}},
			},
		}

		assert.Equal(t, "new_role", req.Name)
		assert.Len(t, req.Permissions, 1)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := models.RoleCreateRequest{
			Name:        "json_role",
			Description: "Role for JSON test",
			Permissions: []models.Permission{
				{Resource: "data", Actions: []string{"read"}},
			},
		}

		jsonData, err := json.Marshal(req)
		require.NoError(t, err)

		var decoded models.RoleCreateRequest
		err = json.Unmarshal(jsonData, &decoded)
		require.NoError(t, err)

		assert.Equal(t, req.Name, decoded.Name)
		assert.Equal(t, req.Description, decoded.Description)
	})
}

// TestRoleUpdateRequest tests role update request
func TestRoleUpdateRequest(t *testing.T) {
	req := models.RoleUpdateRequest{
		Description: "Updated description",
		Permissions: []models.Permission{
			{Resource: "updated", Actions: []string{"read", "write", "delete"}},
		},
	}

	assert.Equal(t, "Updated description", req.Description)
	assert.Len(t, req.Permissions, 1)
	assert.Len(t, req.Permissions[0].Actions, 3)
}

// TestPermissionModel tests the Permission model
func TestPermissionModel(t *testing.T) {
	t.Run("Basic permission", func(t *testing.T) {
		perm := models.Permission{
			Resource: "buildings",
			Actions:  []string{"read", "write"},
		}

		assert.Equal(t, "buildings", perm.Resource)
		assert.Contains(t, perm.Actions, "read")
		assert.Contains(t, perm.Actions, "write")
	})

	t.Run("Multiple permissions", func(t *testing.T) {
		perms := []models.Permission{
			{Resource: "buildings", Actions: []string{"read"}},
			{Resource: "reports", Actions: []string{"read", "write"}},
			{Resource: "alerts", Actions: []string{"read", "write", "delete"}},
		}

		assert.Len(t, perms, 3)
		assert.Len(t, perms[0].Actions, 1)
		assert.Len(t, perms[1].Actions, 2)
		assert.Len(t, perms[2].Actions, 3)
	})
}

// TestDefaultRoles tests the default roles structure
func TestDefaultRoles(t *testing.T) {
	defaultRoles := []struct {
		name        string
		isSystem    bool
		description string
	}{
		{"admin", true, "Administrator with full access"},
		{"user", true, "Standard user with basic access"},
		{"building_manager", true, "Building manager with building and energy access"},
		{"energy_analyst", true, "Energy analyst with read access to energy data"},
	}

	for _, dr := range defaultRoles {
		t.Run(dr.name, func(t *testing.T) {
			assert.NotEmpty(t, dr.name)
			assert.True(t, dr.isSystem)
			assert.NotEmpty(t, dr.description)
		})
	}
}
