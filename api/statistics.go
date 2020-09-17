package api

import (
	"github.com/gin-gonic/gin"
	"github.com/keptn-sandbox/statistics-service/controller"
	"github.com/keptn-sandbox/statistics-service/db"
	"github.com/keptn-sandbox/statistics-service/operations"
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
// @Success 200 {object} operations.Statistics	"ok"
// @Failure 400 {object} operations.Error "Invalid payload"
// @Failure 500 {object} operations.Error "Internal error"
// @Router /statistics [get]
func GetStatistics(c *gin.Context) {
	params := &operations.GetStatisticsParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, operations.Error{
			ErrorCode: 400,
			Message:   "Invalid request format",
		})
		return
	}

	if params.To.Before(params.From) {
		c.JSON(http.StatusBadRequest, operations.Error{
			ErrorCode: 400,
			Message:   "Invalid time frame: 'from' timestamp must not be greater than 'to' timestamp",
		})
		return
	}

	var payload = operations.Statistics{}

	sb := controller.GetStatisticsBucketInstance()
	cutoffTime := sb.GetCutoffTime()

	// check time
	if params.From.After(cutoffTime) {
		// case 1: time frame within "in-memory" interval (e.g. last 30 minutes)
		// -> return in-memory object
		payload = sb.Statistics

	} else {
		var statistics []operations.Statistics
		var err error
		if params.From.Before(cutoffTime) && params.To.Before(cutoffTime) {
			// case 2: time frame outside of "in-memory" interval
			// -> return results from database
			statistics, err = sb.StatisticsRepo.GetStatistics(params.From, params.To)
			if err != nil && err == db.NoStatisticsFoundError {
				c.JSON(http.StatusNotFound, operations.Error{
					Message:   "no statistics found for selected time frame",
					ErrorCode: 404,
				})
				return
			} else if err != nil {
				c.JSON(http.StatusNotFound, operations.Error{
					Message:   "",
					ErrorCode: 500,
				})
				return
			}
		} else if params.From.Before(cutoffTime) && params.To.After(cutoffTime) {
			// case 3: time frame includes "in-memory" interval
			// -> get results from database and from in-memory and merge them
			statistics, err := sb.StatisticsRepo.GetStatistics(params.From, params.To)
			if err != nil {
				c.JSON(http.StatusNotFound, operations.Error{
					Message:   "",
					ErrorCode: 500,
				})
				return
			}
			statistics = append(statistics, sb.Statistics)
		}

		payload = operations.Statistics{
			From: params.From,
			To:   params.To,
		}
		payload = mergeStatistics(payload, statistics)
	}

	c.JSON(http.StatusOK, payload)
}

func mergeStatistics(target operations.Statistics, statistics []operations.Statistics) operations.Statistics {
	for _, stats := range statistics {
		for projectName, project := range stats.Projects {
			for serviceName, service := range project.Services {
				for eventType, count := range service.Events {
					target.IncreaseEventTypeCount(projectName, serviceName, eventType, count)
				}
				if service.ExecutedSequences > 0 {
					target.IncreaseExecutedSequencesCount(projectName, serviceName, service.ExecutedSequences)
				}
			}
		}
	}
	return target
}

func stringp(s string) *string {
	return &s
}
