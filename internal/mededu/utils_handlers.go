package mededu

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/K1ko/mededu-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

const dbServiceContextKey = "db_service"

var defaultDepartments = []string{
	"Urgent",
	"JIS",
	"Chirurgia",
	"Interné",
	"Pediatria",
	"Radiológia",
	"Anestéziológia",
}

func getTrainingDb(ctx *gin.Context) (db_service.DbService[Training], bool) {
	value, exists := ctx.Get(dbServiceContextKey)
	if !exists {
		writeError(ctx, http.StatusInternalServerError, "db_service not found", nil)
		return nil, false
	}

	db, ok := value.(db_service.DbService[Training])
	if !ok {
		writeError(ctx, http.StatusInternalServerError, "db_service context is not of required type", nil)
		return nil, false
	}

	return db, true
}

func writeError(ctx *gin.Context, status int, message string, err error) {
	if err != nil {
		message = message + ": " + err.Error()
	}
	ctx.JSON(status, ErrorResponse{Message: message})
}

func trainingFromInput(id string, input TrainingInput, existing *Training) Training {
	training := Training{
		Id:           id,
		Title:        input.Title,
		Type:         input.Type,
		Department:   input.Department,
		StartAt:      input.StartAt,
		Capacity:     input.Capacity,
		Lecturer:     input.Lecturer,
		Location:     input.Location,
		OnlineLink:   input.OnlineLink,
		Description:  input.Description,
		Requirements: input.Requirements,
		Status:       input.Status,
		Occupied:     0,
		Waitlisted:   0,
	}

	if existing != nil {
		training.Registrations = existing.Registrations
	}

	reconcileTrainingRegistrations(&training)
	return training
}

func registrationFromInput(id string, trainingId string, input RegistrationInput, existing *Registration) Registration {
	registration := Registration{
		Id:            id,
		TrainingId:    trainingId,
		EmployeeId:    strings.TrimSpace(input.EmployeeId),
		EmployeeName:  strings.TrimSpace(input.EmployeeName),
		EmployeeEmail: strings.TrimSpace(input.EmployeeEmail),
		Department:    strings.TrimSpace(input.Department),
		Note:          strings.TrimSpace(input.Note),
		Status:        REGISTERED,
		RegisteredAt:  time.Now().UTC(),
	}

	if existing != nil {
		registration.Status = existing.Status
		registration.RegisteredAt = existing.RegisteredAt
	}

	return registration
}

func validateTrainingInput(input TrainingInput) string {
	switch {
	case strings.TrimSpace(input.Title) == "":
		return "Nazov skolenia je povinny."
	case len([]rune(strings.TrimSpace(input.Title))) < 3:
		return "Nazov skolenia musi mat aspon 3 znaky."
	case strings.TrimSpace(string(input.Type)) == "":
		return "Typ skolenia je povinny."
	case strings.TrimSpace(input.Department) == "":
		return "Oddelenie je povinne."
	case input.StartAt.IsZero():
		return "Datum a cas skolenia je povinny."
	case input.Capacity < 1:
		return "Kapacita musi byt aspon 1."
	case strings.TrimSpace(input.Lecturer) == "":
		return "Lektor je povinny."
	case strings.TrimSpace(string(input.Status)) == "":
		return "Stav skolenia je povinny."
	}
	return ""
}

func validateRegistrationInput(input RegistrationInput) string {
	switch {
	case strings.TrimSpace(input.EmployeeId) == "":
		return "Identifikator zamestnanca je povinny."
	case strings.TrimSpace(input.EmployeeName) == "":
		return "Meno zamestnanca je povinne."
	}
	return ""
}

func sortedTrainings(trainings []Training) []Training {
	slices.SortFunc(trainings, func(left, right Training) int {
		if left.StartAt.Before(right.StartAt) {
			return -1
		}
		if left.StartAt.After(right.StartAt) {
			return 1
		}
		return strings.Compare(left.Title, right.Title)
	})
	return trainings
}

func sortedRegistrations(registrations []Registration) []Registration {
	if registrations == nil {
		return []Registration{}
	}

	sorted := append([]Registration(nil), registrations...)
	slices.SortStableFunc(sorted, func(left, right Registration) int {
		if left.Status != right.Status {
			if left.Status == REGISTERED {
				return -1
			}
			if right.Status == REGISTERED {
				return 1
			}
		}
		if left.RegisteredAt.Before(right.RegisteredAt) {
			return -1
		}
		if left.RegisteredAt.After(right.RegisteredAt) {
			return 1
		}
		return strings.Compare(left.EmployeeName, right.EmployeeName)
	})
	return sorted
}

func reconcileTrainingRegistrations(training *Training) {
	if training.Registrations == nil {
		training.Registrations = []Registration{}
	}

	slices.SortStableFunc(training.Registrations, func(left, right Registration) int {
		if left.RegisteredAt.Before(right.RegisteredAt) {
			return -1
		}
		if left.RegisteredAt.After(right.RegisteredAt) {
			return 1
		}
		return strings.Compare(left.Id, right.Id)
	})

	training.Occupied = 0
	training.Waitlisted = 0
	for index := range training.Registrations {
		training.Registrations[index].TrainingId = training.Id
		if training.Occupied < training.Capacity {
			training.Registrations[index].Status = REGISTERED
			training.Occupied++
			continue
		}
		training.Registrations[index].Status = WAITLISTED
		training.Waitlisted++
	}
}

func findRegistration(training *Training, registrationId string) (Registration, bool) {
	index := findRegistrationIndex(training, registrationId)
	if index == -1 {
		return Registration{}, false
	}
	return training.Registrations[index], true
}

func findRegistrationIndex(training *Training, registrationId string) int {
	for index, registration := range training.Registrations {
		if registration.Id == registrationId {
			return index
		}
	}
	return -1
}

func hasEmployeeRegistration(training *Training, employeeId string, excludedRegistrationId string) bool {
	employeeId = strings.TrimSpace(employeeId)
	for _, registration := range training.Registrations {
		if registration.Id == excludedRegistrationId {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(registration.EmployeeId), employeeId) {
			return true
		}
	}
	return false
}

func departmentsFromTrainings(trainings []Training) []string {
	seen := map[string]bool{}
	departments := make([]string, 0, len(defaultDepartments))
	for _, department := range defaultDepartments {
		seen[department] = true
		departments = append(departments, department)
	}

	for _, training := range trainings {
		department := strings.TrimSpace(training.Department)
		if department == "" || seen[department] {
			continue
		}
		seen[department] = true
		departments = append(departments, department)
	}

	slices.Sort(departments)
	return departments
}
