package api

import (
	"github.com/gin-gonic/gin"
	"github.com/keptn-sandbox/statistics-service/operations"
	"github.com/keptn/go-utils/pkg/api/models"
	"net/http"
)

// HandleEvent godoc
// @Summary Handle event
// @Description Handle incoming cloud event
// @Tags Events
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   event     body    models.Event     true        "Event type"
// @Success 200 "ok"
// @Failure 400 {object} models.Error "Invalid payload"
// @Failure 500 {object} models.Error "Internal error"
// @Router /event [post]
func HandleEvent(c *gin.Context) {
	event := operations.Event{}
	if err := c.ShouldBindJSON(event); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    400,
			Message: stringp("Invalid request format"),
		})
	}

	c.Status(http.StatusOK)
}
