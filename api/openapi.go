package api

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed mededu.openapi.yaml
var openapiSpec []byte

func HandleOpenAPI(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/yaml", openapiSpec)
}
