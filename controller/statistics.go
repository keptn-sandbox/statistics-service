package controller

import (
	"fmt"
	"github.com/keptn-sandbox/statistics-service/config"
	"github.com/keptn-sandbox/statistics-service/db"
	"github.com/keptn-sandbox/statistics-service/operations"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"sync"
	"time"
)

var statisticsBucketInstance *statisticsBucket

type statisticsBucket struct {
	StatisticsRepo  db.StatisticsRepo
	Statistics      operations.Statistics
	bucketTimer     *time.Ticker
	uniqueSequences map[string]bool
	logger          keptn.LoggerInterface
	lock            sync.Mutex
	cutoffTime      time.Time
}

func GetStatisticsBucketInstance() *statisticsBucket {
	if statisticsBucketInstance == nil {
		env := config.GetConfig()
		statisticsBucketInstance = &statisticsBucket{
			StatisticsRepo: db.StatisticsMongoDBRepo{},
			logger:         keptn.NewLogger("", "", "statistics service"),
		}

		statisticsBucketInstance.bucketTimer = time.NewTicker(time.Duration(env.AggregationIntervalMinutes) * time.Minute)
		statisticsBucketInstance.createNewBucket()
		go func() {
			for {
				<-statisticsBucketInstance.bucketTimer.C
				statisticsBucketInstance.logger.Info(fmt.Sprintf("%n minutes have passed. Creating a new statistics bucket\n", env.AggregationIntervalMinutes))
				statisticsBucketInstance.storeCurrentBucket()
				statisticsBucketInstance.createNewBucket()
			}
		}()
	}
	return statisticsBucketInstance
}

func (sb *statisticsBucket) GetCutoffTime() time.Time {
	return sb.cutoffTime
}

func (sb *statisticsBucket) AddEvent(event operations.Event) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	sb.logger.Info("updating statistics for service " + event.Data.Service + " in project " + event.Data.Project)
	increaseExecutedSequences := sb.uniqueSequences[event.Shkeptncontext]
	sb.uniqueSequences[event.Shkeptncontext] = true

	sb.Statistics.IncreaseEventTypeCount(event.Data.Project, event.Data.Service, event.Type, 1)
	if increaseExecutedSequences {
		sb.Statistics.IncreaseExecutedSequencesCount(event.Data.Project, event.Data.Service, 1)
	}
}

func (sb *statisticsBucket) storeCurrentBucket() {
	sb.lock.Lock()
	defer sb.lock.Unlock()
	sb.logger.Info(fmt.Sprintf("Storing statistics for time frame %s - %s\n\n", sb.Statistics.From.String(), sb.Statistics.To.String()))
	if err := sb.StatisticsRepo.StoreStatistics(sb.Statistics); err != nil {
		sb.logger.Error(fmt.Sprintf("Could not store statistics: " + err.Error()))
	}
	sb.logger.Info(fmt.Sprintf("Statistics stored successfully"))
}

func (sb *statisticsBucket) createNewBucket() {
	sb.lock.Lock()
	defer sb.lock.Unlock()
	sb.cutoffTime = time.Now()
	sb.uniqueSequences = map[string]bool{}
	sb.Statistics = operations.Statistics{
		From: time.Now(),
	}
}
