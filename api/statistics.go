package api

import (
	"github.com/gin-gonic/gin"
	"github.com/keptn-sandbox/statistics-service/operations"
	"github.com/keptn/go-utils/pkg/api/models"
	"net/http"
)

// GetStatistics godoc
// @Summary Get statistics
// @Description get statistics about Keptn installation
// @Tags Statistics
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   from     query    string     false        "From"
// @Param   to     query    string     false        "To"
// @Success 200 {object} operations.GetStatisticsResponse	"ok"
// @Failure 400 {object} operations.Error "Invalid payload"
// @Failure 500 {object} operations.Error "Internal error"
// @Router /statistics [get]
func GetStatistics(c *gin.Context) {
	params := &operations.GetStatisticsParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    400,
			Message: stringp("Invalid request format"),
		})
	}

	var payload = &operations.GetStatisticsResponse{}

	c.JSON(http.StatusOK, payload)
}

func stringp(s string) *string {
	return &s
}
