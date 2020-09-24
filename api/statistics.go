package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/keptn-sandbox/statistics-service/controller"
	"github.com/keptn-sandbox/statistics-service/db"
	"github.com/keptn-sandbox/statistics-service/operations"
	keptn "github.com/keptn/go-utils/pkg/lib"
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
	logger := keptn.NewLogger("", "", "statistics-service")
	params := &operations.GetStatisticsParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, operations.Error{
			ErrorCode: 400,
			Message:   "Invalid request format",
		})
		return
	}

	if !validateQueryTimestamps(params) {
		c.JSON(http.StatusBadRequest, operations.Error{
			ErrorCode: 400,
			Message:   "Invalid time frame: 'from' timestamp must not be greater than 'to' timestamp",
		})
		return
	}

	sb := controller.GetStatisticsBucketInstance()

	payload, err := getStatistics(params, sb)

	if err != nil && err == db.NoStatisticsFoundError {
		c.JSON(http.StatusNotFound, operations.Error{
			Message:   "no statistics found for selected time frame",
			ErrorCode: 404,
		})
		return
	} else if err != nil {
		logger.Error("could not retrieve statistics: " + err.Error())
		c.JSON(http.StatusInternalServerError, operations.Error{
			Message:   "Internal server error",
			ErrorCode: 500,
		})
		return
	}

	c.JSON(http.StatusOK, payload)
}

func getStatistics(params *operations.GetStatisticsParams, sb controller.StatisticsInterface) (operations.GetStatisticsResponse, error) {
	var mergedStatistics = operations.Statistics{}

	cutoffTime := sb.GetCutoffTime()

	// check time
	if params.From.After(cutoffTime) {
		// case 1: time frame within "in-memory" interval (e.g. last 30 minutes)
		// -> return in-memory object
		mergedStatistics = *sb.GetStatistics()

	} else {
		var statistics []operations.Statistics
		var err error
		if params.From.Before(cutoffTime) && params.To.Before(cutoffTime) {
			// case 2: time frame outside of "in-memory" interval
			// -> return results from database
			statistics, err = sb.GetRepo().GetStatistics(params.From, params.To)
			if err != nil && err == db.NoStatisticsFoundError {
				return operations.GetStatisticsResponse{}, err
			}
		} else if params.From.Before(cutoffTime) && params.To.After(cutoffTime) {
			// case 3: time frame includes "in-memory" interval
			// -> get results from database and from in-memory and merge them
			statistics, err = sb.GetRepo().GetStatistics(params.From, params.To)
			if statistics == nil {
				statistics = []operations.Statistics{}
			}
			statistics = append(statistics, *sb.GetStatistics())
		}

		mergedStatistics = operations.Statistics{
			From: params.From,
			To:   params.To,
		}
		mergedStatistics = operations.MergeStatistics(mergedStatistics, statistics)
	}
	return convertToGetStatisticsResponse(mergedStatistics)
}

func convertToGetStatisticsResponse(mergedStatistics operations.Statistics) (operations.GetStatisticsResponse, error) {
	marshal, _ := json.Marshal(mergedStatistics)
	var result operations.GetStatisticsResponse

	err := json.Unmarshal(marshal, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func validateQueryTimestamps(params *operations.GetStatisticsParams) bool {
	if params.To.Before(params.From) {
		return false
	}
	return true
}
