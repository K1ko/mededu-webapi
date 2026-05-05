package mededu

import (
	"errors"
	"net/http"
	"strings"

	"github.com/K1ko/mededu-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implTrainingsAPI struct {
}

func NewTrainingsAPI() TrainingsAPI {
	return &implTrainingsAPI{}
}

func (o implTrainingsAPI) CreateTraining(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	var input TrainingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "Neplatne vstupne data", err)
		return
	}

	if message := validateTrainingInput(input); message != "" {
		writeError(c, http.StatusBadRequest, message, nil)
		return
	}

	training := trainingFromInput(uuid.NewString(), input, nil)
	err := db.CreateDocument(c.Request.Context(), training.Id, &training)
	switch {
	case err == nil:
		c.JSON(http.StatusCreated, training)
	case errors.Is(err, db_service.ErrConflict):
		writeError(c, http.StatusConflict, "Skolenie s rovnakym identifikatorom uz existuje", err)
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa vytvorit skolenie v databaze", err)
	}
}

func (o implTrainingsAPI) DeleteTraining(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	err := db.DeleteDocument(c.Request.Context(), c.Param("trainingId"))
	switch {
	case err == nil:
		c.AbortWithStatus(http.StatusNoContent)
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie nebolo najdene", err)
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa odstranit skolenie z databazy", err)
	}
}

func (o implTrainingsAPI) GetTraining(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, err := db.FindDocument(c.Request.Context(), c.Param("trainingId"))
	switch {
	case err == nil:
		c.JSON(http.StatusOK, training)
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie nebolo najdene", err)
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa nacitat skolenie z databazy", err)
	}
}

func (o implTrainingsAPI) ListTrainings(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	trainings, err := db.ListDocuments(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusBadGateway, "Nepodarilo sa nacitat skolenia z databazy", err)
		return
	}

	department := strings.TrimSpace(c.Query("department"))
	if department != "" {
		filtered := make([]Training, 0, len(trainings))
		for _, training := range trainings {
			if strings.EqualFold(training.Department, department) {
				filtered = append(filtered, training)
			}
		}
		trainings = filtered
	}

	c.JSON(http.StatusOK, sortedTrainings(trainings))
}

func (o implTrainingsAPI) UpdateTraining(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	trainingId := c.Param("trainingId")
	existing, err := db.FindDocument(c.Request.Context(), trainingId)
	switch {
	case err == nil:
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie nebolo najdene", err)
		return
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa nacitat skolenie z databazy", err)
		return
	}

	var input TrainingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "Neplatne vstupne data", err)
		return
	}

	if message := validateTrainingInput(input); message != "" {
		writeError(c, http.StatusBadRequest, message, nil)
		return
	}

	training := trainingFromInput(trainingId, input, existing)
	err = db.UpdateDocument(c.Request.Context(), trainingId, &training)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, training)
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie bolo odstranene pocas spracovania poziadavky", err)
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa upravit skolenie v databaze", err)
	}
}
