package mededu

import "github.com/gin-gonic/gin"

type implDepartmentsAPI struct {
}

func NewDepartmentsAPI() DepartmentsAPI {
	return &implDepartmentsAPI{}
}

func (o implDepartmentsAPI) ListDepartments(c *gin.Context) {
	db, ok := getTrainingDb(c)
	if !ok {
		return
	}

	trainings, err := db.ListDocuments(c.Request.Context())
	if err != nil {
		writeError(c, 502, "Nepodarilo sa nacitat oddelenia z databazy", err)
		return
	}

	c.JSON(200, departmentsFromTrainings(trainings))
}
