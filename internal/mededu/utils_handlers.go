package mededu

import (
	"net/http"
	"slices"
	"strings"

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
		training.Occupied = existing.Occupied
		training.Waitlisted = existing.Waitlisted
	}

	return training
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
