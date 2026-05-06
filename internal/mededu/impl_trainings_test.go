package mededu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/K1ko/mededu-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TrainingsSuite struct {
	suite.Suite
	dbServiceMock *DbServiceMock[Training]
}

func TestTrainingsSuite(t *testing.T) {
	suite.Run(t, new(TrainingsSuite))
}

type DbServiceMock[DocType interface{}] struct {
	mock.Mock
}

func (m *DbServiceMock[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *DbServiceMock[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	args := m.Called(ctx, id)
	if document := args.Get(0); document != nil {
		return document.(*DocType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *DbServiceMock[DocType]) ListDocuments(ctx context.Context) ([]DocType, error) {
	args := m.Called(ctx)
	if documents := args.Get(0); documents != nil {
		return documents.([]DocType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *DbServiceMock[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	args := m.Called(ctx, id, document)
	return args.Error(0)
}

func (m *DbServiceMock[DocType]) DeleteDocument(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *DbServiceMock[DocType]) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (suite *TrainingsSuite) SetupTest() {
	suite.dbServiceMock = &DbServiceMock[Training]{}

	var _ db_service.DbService[Training] = suite.dbServiceMock

	suite.dbServiceMock.
		On("FindDocument", mock.Anything, "test-training").
		Return(&Training{
			Id:         "test-training",
			Title:      "Povodne skolenie",
			Type:       MANDATORY,
			Department: "Urgent",
			StartAt:    time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC),
			Capacity:   20,
			Lecturer:   "Mgr. Jana Novakova",
			Location:   "Skoliaca miestnost A",
			Status:     PLANNED,
			Occupied:   2,
			Waitlisted: 0,
			Registrations: []Registration{
				testRegistration("reg-001", "EMP-001", "Peter Malina", time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)),
				testRegistration("reg-002", "EMP-002", "Lucia Krizova", time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)),
			},
		}, nil).
		Maybe()
}

func testRegistration(id string, employeeId string, employeeName string, registeredAt time.Time) Registration {
	return Registration{
		Id:           id,
		TrainingId:   "test-training",
		EmployeeId:   employeeId,
		EmployeeName: employeeName,
		Status:       REGISTERED,
		RegisteredAt: registeredAt,
	}
}

func (suite *TrainingsSuite) Test_UpdateTraining_DbServiceUpdateCalled() {
	// ARRANGE
	suite.dbServiceMock.
		On("UpdateDocument", mock.Anything, "test-training", mock.MatchedBy(func(training *Training) bool {
			return training.Id == "test-training" &&
				training.Title == "Aktualizovane skolenie" &&
				training.Type == DEPARTMENT &&
				training.Department == "JIS" &&
				training.Capacity == 24 &&
				training.Occupied == 2 &&
				training.Waitlisted == 0 &&
				len(training.Registrations) == 2
		})).
		Return(nil)

	json := `{
		"title": "Aktualizovane skolenie",
		"type": "department",
		"department": "JIS",
		"startAt": "2026-06-02T12:30:00Z",
		"capacity": 24,
		"lecturer": "MUDr. Eva Hruba",
		"location": "Skoliaca miestnost B",
		"onlineLink": "",
		"description": "Aktualizovany popis skolenia.",
		"requirements": "Zamestnanecky preukaz",
		"status": "planned"
	}`

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(dbServiceContextKey, suite.dbServiceMock)
	ctx.Params = []gin.Param{
		{Key: "trainingId", Value: "test-training"},
	}
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/trainings/test-training", strings.NewReader(json))
	ctx.Request.Header.Set("Content-Type", "application/json")

	sut := implTrainingsAPI{}

	// ACT
	sut.UpdateTraining(ctx)

	// ASSERT
	suite.Equal(http.StatusOK, recorder.Code)
	suite.dbServiceMock.AssertExpectations(suite.T())
}

func (suite *TrainingsSuite) Test_CreateRegistration_WaitlistsWhenCapacityIsFull() {
	// ARRANGE
	suite.dbServiceMock.
		On("FindDocument", mock.Anything, "full-training").
		Return(&Training{
			Id:         "full-training",
			Title:      "Plne skolenie",
			Type:       MANDATORY,
			Department: "Urgent",
			StartAt:    time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC),
			Capacity:   1,
			Lecturer:   "Mgr. Jana Novakova",
			Status:     PLANNED,
			Occupied:   1,
			Registrations: []Registration{
				testRegistration("reg-001", "EMP-001", "Peter Malina", time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)),
			},
		}, nil)

	suite.dbServiceMock.
		On("UpdateDocument", mock.Anything, "full-training", mock.MatchedBy(func(training *Training) bool {
			return training.Id == "full-training" &&
				training.Occupied == 1 &&
				training.Waitlisted == 1 &&
				len(training.Registrations) == 2 &&
				training.Registrations[1].EmployeeId == "EMP-002" &&
				training.Registrations[1].Status == WAITLISTED
		})).
		Return(nil)

	json := `{
		"employeeId": "EMP-002",
		"employeeName": "Lucia Krizova",
		"employeeEmail": "lucia.krizova@hospital.example",
		"department": "JIS",
		"note": "Nahradny termin vyhovuje"
	}`

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(dbServiceContextKey, suite.dbServiceMock)
	ctx.Params = []gin.Param{
		{Key: "trainingId", Value: "full-training"},
	}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/trainings/full-training/registrations", strings.NewReader(json))
	ctx.Request.Header.Set("Content-Type", "application/json")

	sut := implRegistrationsAPI{}

	// ACT
	sut.CreateRegistration(ctx)

	// ASSERT
	suite.Equal(http.StatusCreated, recorder.Code)
	suite.dbServiceMock.AssertExpectations(suite.T())
}

func (suite *TrainingsSuite) Test_DeleteRegistration_PromotesWaitlistedRegistration() {
	// ARRANGE
	waitlisted := testRegistration("reg-002", "EMP-002", "Lucia Krizova", time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC))
	waitlisted.Status = WAITLISTED

	suite.dbServiceMock.
		On("FindDocument", mock.Anything, "full-training").
		Return(&Training{
			Id:         "full-training",
			Title:      "Plne skolenie",
			Type:       MANDATORY,
			Department: "Urgent",
			StartAt:    time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC),
			Capacity:   1,
			Lecturer:   "Mgr. Jana Novakova",
			Status:     PLANNED,
			Occupied:   1,
			Waitlisted: 1,
			Registrations: []Registration{
				testRegistration("reg-001", "EMP-001", "Peter Malina", time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)),
				waitlisted,
			},
		}, nil)

	suite.dbServiceMock.
		On("UpdateDocument", mock.Anything, "full-training", mock.MatchedBy(func(training *Training) bool {
			return training.Occupied == 1 &&
				training.Waitlisted == 0 &&
				len(training.Registrations) == 1 &&
				training.Registrations[0].Id == "reg-002" &&
				training.Registrations[0].Status == REGISTERED
		})).
		Return(nil)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(dbServiceContextKey, suite.dbServiceMock)
	ctx.Params = []gin.Param{
		{Key: "trainingId", Value: "full-training"},
		{Key: "registrationId", Value: "reg-001"},
	}
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/trainings/full-training/registrations/reg-001", nil)

	sut := implRegistrationsAPI{}

	// ACT
	sut.DeleteRegistration(ctx)

	// ASSERT
	suite.Equal(http.StatusNoContent, recorder.Code)
	suite.dbServiceMock.AssertExpectations(suite.T())
}

func (suite *TrainingsSuite) Test_SortedRegistrations_ReturnsNonNilSliceForEmptyInput() {
	suite.NotNil(sortedRegistrations(nil))
	suite.NotNil(sortedRegistrations([]Registration{}))
	suite.Empty(sortedRegistrations(nil))
	suite.Empty(sortedRegistrations([]Registration{}))
}

func (suite *TrainingsSuite) Test_TrainingForResponse_RecalculatesCountsFromRegistrations() {
	waitlisted := testRegistration("reg-003", "EMP-003", "Tomas Hruska", time.Date(2026, 5, 3, 8, 0, 0, 0, time.UTC))
	waitlisted.Status = WAITLISTED

	training := trainingForResponse(Training{
		Id:         "stale-training",
		Title:      "Skolenie so starymi poctami",
		Type:       DEPARTMENT,
		Department: "JIS",
		StartAt:    time.Date(2026, 6, 2, 12, 30, 0, 0, time.UTC),
		Capacity:   2,
		Lecturer:   "MUDr. Eva Hruba",
		Status:     PLANNED,
		Occupied:   0,
		Waitlisted: 0,
		Registrations: []Registration{
			testRegistration("reg-001", "EMP-001", "Peter Malina", time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)),
			testRegistration("reg-002", "EMP-002", "Lucia Krizova", time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)),
			waitlisted,
		},
	})

	suite.Equal(int32(2), training.Occupied)
	suite.Equal(int32(1), training.Waitlisted)
	suite.Len(training.Registrations, 3)
	suite.Equal(REGISTERED, training.Registrations[0].Status)
	suite.Equal(REGISTERED, training.Registrations[1].Status)
	suite.Equal(WAITLISTED, training.Registrations[2].Status)
}
