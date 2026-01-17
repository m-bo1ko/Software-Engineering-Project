package tests

import (
	"context"
	"testing"
	"time"

	"forecast-service/internal/config"
	"forecast-service/internal/integrations"
	"forecast-service/internal/models"
	"forecast-service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestOptimizationScenarioCreation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizationRepo := repository.NewOptimizationRepository(db.Collection("optimization_scenarios"))

	ctx := context.Background()
	scenario := &models.OptimizationScenario{
		BuildingID:     "test-building-1",
		Name:           "Test Optimization Scenario",
		Description:    "Test description",
		Type:           models.OptimizationTypeCostReduction,
		Status:         models.OptimizationStatusDraft,
		ScheduledStart: time.Now().Add(time.Hour),
		ScheduledEnd:   time.Now().Add(8 * time.Hour),
		Actions: []models.OptimizationAction{
			{
				ID:             "action-1",
				DeviceID:       "device-1",
				DeviceName:     "Test Device",
				DeviceType:     "HVAC",
				ActionType:     "REDUCE_POWER",
				CurrentValue:   "25.0 kW",
				TargetValue:    "20.0 kW",
				ScheduledTime:  time.Now().Add(time.Hour),
				Duration:       60,
				Status:         "PENDING",
				ExpectedImpact: 5.0,
			},
		},
		ExpectedSavings: models.Savings{
			EnergyKWh:       10.0,
			CostAmount:      1.50,
			Currency:        "USD",
			CO2ReductionKg:  4.0,
			PercentReduction: 12.5,
		},
		Priority: 5,
		CreatedBy: "test-user",
	}

	created, err := optimizationRepo.Create(ctx, scenario)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "test-building-1", created.BuildingID)
	assert.Len(t, created.Actions, 1)
	assert.Equal(t, models.OptimizationStatusDraft, created.Status)
}

func TestOptimizationScenarioApproval(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizationRepo := repository.NewOptimizationRepository(db.Collection("optimization_scenarios"))

	ctx := context.Background()
	scenario := &models.OptimizationScenario{
		BuildingID:     "test-building-1",
		Name:           "Test Scenario",
		Type:           models.OptimizationTypeCostReduction,
		Status:         models.OptimizationStatusDraft,
		ScheduledStart: time.Now().Add(time.Hour),
		ScheduledEnd:   time.Now().Add(8 * time.Hour),
		CreatedBy:      "test-user",
	}

	created, err := optimizationRepo.Create(ctx, scenario)
	require.NoError(t, err)

	// Approve scenario
	err = optimizationRepo.ApproveScenario(ctx, created.ID.Hex(), "approver-user")
	require.NoError(t, err)

	// Retrieve and verify
	approved, err := optimizationRepo.FindByID(ctx, created.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, models.OptimizationStatusApproved, approved.Status)
	assert.Equal(t, "approver-user", approved.ApprovedBy)
	assert.NotNil(t, approved.ApprovedAt)
}

func TestRecommendationCreation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	recommendationRepo := repository.NewRecommendationRepository(db.Collection("recommendations"))

	ctx := context.Background()
	rec := &models.Recommendation{
		BuildingID:  "test-building-1",
		Type:        models.RecommendationTypeImmediate,
		Priority:    models.RecommendationPriorityHigh,
		Title:       "Test Recommendation",
		Description: "Test description",
		ActionRequired: "Test action",
		ExpectedSavings: models.Savings{
			EnergyKWh:       100.0,
			CostAmount:      15.0,
			Currency:        "USD",
			CO2ReductionKg:  40.0,
		},
		ImplementationSteps: []string{"Step 1", "Step 2"},
		AutomationAvailable: true,
		Category:            "HVAC",
		ValidFrom:           time.Now(),
	}

	created, err := recommendationRepo.Create(ctx, rec)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "test-building-1", created.BuildingID)
	assert.Equal(t, "NEW", created.Status)
}

func TestIntegrationWithSecurityService(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityServiceConfig{
			URL:     "http://localhost:8080",
			Timeout: 5 * time.Second,
		},
	}

	securityClient := integrations.NewSecurityClient(cfg)
	assert.NotNil(t, securityClient)

	// Test that client can be created (won't actually validate without running service)
	ctx := context.Background()
	_, err := securityClient.ValidateToken(ctx, "test-token")
	// We expect this to fail without the actual service running
	// but it tests that the client is properly initialized
	assert.Error(t, err) // Expected since service won't be running
}

