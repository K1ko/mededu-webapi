package mededu

import (
	"errors"
	"net/http"
	"strings"

	"github.com/K1ko/mededu-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implRegistrationsAPI struct {
}

func NewRegistrationsAPI() RegistrationsAPI {
	return &implRegistrationsAPI{}
}

func (o implRegistrationsAPI) CreateRegistration(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, ok := findTrainingForRegistration(c, db)
	if !ok {
		return
	}

	if training.Status != PLANNED {
		writeError(c, http.StatusConflict, "Na skolenie sa da prihlasit iba v stave planned", nil)
		return
	}

	var input RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "Neplatne vstupne data", err)
		return
	}

	if message := validateRegistrationInput(input); message != "" {
		writeError(c, http.StatusBadRequest, message, nil)
		return
	}

	if hasEmployeeRegistration(training, input.EmployeeId, "") {
		writeError(c, http.StatusConflict, "Zamestnanec je uz na skolenie prihlaseny", nil)
		return
	}

	registration := registrationFromInput(uuid.NewString(), training.Id, input, nil)
	training.Registrations = append(training.Registrations, registration)
	reconcileTrainingRegistrations(training)

	if err := db.UpdateDocument(c.Request.Context(), training.Id, training); err != nil {
		writeRegistrationDbError(c, "Nepodarilo sa vytvorit registraciu v databaze", err)
		return
	}

	if created, found := findRegistration(training, registration.Id); found {
		c.JSON(http.StatusCreated, created)
		return
	}
	writeError(c, http.StatusInternalServerError, "Registracia bola vytvorena, ale nepodarilo sa ju nacitat", nil)
}

func (o implRegistrationsAPI) DeleteRegistration(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, ok := findTrainingForRegistration(c, db)
	if !ok {
		return
	}

	index := findRegistrationIndex(training, c.Param("registrationId"))
	if index == -1 {
		writeError(c, http.StatusNotFound, "Registracia nebola najdena", nil)
		return
	}

	training.Registrations = append(training.Registrations[:index], training.Registrations[index+1:]...)
	reconcileTrainingRegistrations(training)

	if err := db.UpdateDocument(c.Request.Context(), training.Id, training); err != nil {
		writeRegistrationDbError(c, "Nepodarilo sa odstranit registraciu z databazy", err)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func (o implRegistrationsAPI) GetRegistration(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, ok := findTrainingForRegistration(c, db)
	if !ok {
		return
	}

	registration, found := findRegistration(training, c.Param("registrationId"))
	if !found {
		writeError(c, http.StatusNotFound, "Registracia nebola najdena", nil)
		return
	}

	c.JSON(http.StatusOK, registration)
}

func (o implRegistrationsAPI) ListRegistrations(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, ok := findTrainingForRegistration(c, db)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, sortedRegistrations(training.Registrations))
}

func (o implRegistrationsAPI) UpdateRegistration(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	training, ok := findTrainingForRegistration(c, db)
	if !ok {
		return
	}

	registrationId := c.Param("registrationId")
	index := findRegistrationIndex(training, registrationId)
	if index == -1 {
		writeError(c, http.StatusNotFound, "Registracia nebola najdena", nil)
		return
	}

	var input RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "Neplatne vstupne data", err)
		return
	}

	if message := validateRegistrationInput(input); message != "" {
		writeError(c, http.StatusBadRequest, message, nil)
		return
	}

	targetTrainingId := strings.TrimSpace(input.TargetTrainingId)
	if targetTrainingId != "" && targetTrainingId != training.Id {
		o.moveRegistration(c, db, training, index, input, targetTrainingId)
		return
	}

	if hasEmployeeRegistration(training, input.EmployeeId, registrationId) {
		writeError(c, http.StatusConflict, "Zamestnanec je uz na skolenie prihlaseny", nil)
		return
	}

	existing := training.Registrations[index]
	training.Registrations[index] = registrationFromInput(registrationId, training.Id, input, &existing)
	reconcileTrainingRegistrations(training)

	if err := db.UpdateDocument(c.Request.Context(), training.Id, training); err != nil {
		writeRegistrationDbError(c, "Nepodarilo sa upravit registraciu v databaze", err)
		return
	}

	registration, _ := findRegistration(training, registrationId)
	c.JSON(http.StatusOK, registration)
}

func (o implRegistrationsAPI) moveRegistration(
	c *gin.Context,
	db db_service.DbService[Training],
	sourceTraining *Training,
	sourceIndex int,
	input RegistrationInput,
	targetTrainingId string,
) {
	targetTraining, err := db.FindDocument(c.Request.Context(), targetTrainingId)
	switch {
	case err == nil:
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Cielove skolenie nebolo najdene", err)
		return
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa nacitat cielove skolenie z databazy", err)
		return
	}

	if targetTraining.Status != PLANNED {
		writeError(c, http.StatusConflict, "Na cielove skolenie sa da prihlasit iba v stave planned", nil)
		return
	}

	registrationId := sourceTraining.Registrations[sourceIndex].Id
	if hasEmployeeRegistration(targetTraining, input.EmployeeId, registrationId) {
		writeError(c, http.StatusConflict, "Zamestnanec je uz na cielove skolenie prihlaseny", nil)
		return
	}

	existing := sourceTraining.Registrations[sourceIndex]
	movedRegistration := registrationFromInput(registrationId, targetTraining.Id, input, &existing)
	sourceTraining.Registrations = append(sourceTraining.Registrations[:sourceIndex], sourceTraining.Registrations[sourceIndex+1:]...)
	targetTraining.Registrations = append(targetTraining.Registrations, movedRegistration)
	reconcileTrainingRegistrations(sourceTraining)
	reconcileTrainingRegistrations(targetTraining)

	if err := db.UpdateDocument(c.Request.Context(), sourceTraining.Id, sourceTraining); err != nil {
		writeRegistrationDbError(c, "Nepodarilo sa odstranit registraciu z povodneho skolenia", err)
		return
	}

	if err := db.UpdateDocument(c.Request.Context(), targetTraining.Id, targetTraining); err != nil {
		writeRegistrationDbError(c, "Nepodarilo sa presunut registraciu na cielove skolenie", err)
		return
	}

	registration, _ := findRegistration(targetTraining, registrationId)
	c.JSON(http.StatusOK, registration)
}

func findTrainingForRegistration(c *gin.Context, db db_service.DbService[Training]) (*Training, bool) {
	training, err := db.FindDocument(c.Request.Context(), c.Param("trainingId"))
	switch {
	case err == nil:
		return training, true
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie nebolo najdene", err)
	default:
		writeError(c, http.StatusBadGateway, "Nepodarilo sa nacitat skolenie z databazy", err)
	}
	return nil, false
}

func writeRegistrationDbError(c *gin.Context, message string, err error) {
	switch {
	case errors.Is(err, db_service.ErrNotFound):
		writeError(c, http.StatusNotFound, "Skolenie bolo odstranene pocas spracovania poziadavky", err)
	default:
		writeError(c, http.StatusBadGateway, message, err)
	}
}
