package mededu

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implDepartmentsAPI struct {
}

func NewDepartmentsAPI() DepartmentsAPI {
	return &implDepartmentsAPI{}
}

func (o implDepartmentsAPI) ListDepartments(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
