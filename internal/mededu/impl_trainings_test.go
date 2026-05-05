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
			Occupied:   8,
			Waitlisted: 2,
		}, nil)
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
				training.Occupied == 8 &&
				training.Waitlisted == 2
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
