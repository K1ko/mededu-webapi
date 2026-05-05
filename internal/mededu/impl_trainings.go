package mededu

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implTrainingsAPI struct {
}

func NewTrainingsAPI() TrainingsAPI {
	return &implTrainingsAPI{}
}

func (o implTrainingsAPI) CreateTraining(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implTrainingsAPI) DeleteTraining(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implTrainingsAPI) GetTraining(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implTrainingsAPI) ListTrainings(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o implTrainingsAPI) UpdateTraining(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
