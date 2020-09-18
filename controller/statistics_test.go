package controller

import (
	"github.com/go-test/deep"
	"github.com/keptn-sandbox/statistics-service/db"
	"github.com/keptn-sandbox/statistics-service/operations"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"sync"
	"testing"
	"time"
)

// MockStatisticsRepo godoc
type MockStatisticsRepo struct {
	// GetStatisticsFunc godoc
	GetStatisticsFunc func(from, to time.Time) ([]operations.Statistics, error)
	// StoreStatisticsFunc godoc
	StoreStatisticsFunc func(statistics operations.Statistics) error
	// DeleteStatisticsFunc godoc
	DeleteStatisticsFunc func(from, to time.Time) error
}

// GetStatistics godoc
func (m *MockStatisticsRepo) GetStatistics(from, to time.Time) ([]operations.Statistics, error) {
	return m.GetStatisticsFunc(from, to)
}

// StoreStatistics godoc
func (m *MockStatisticsRepo) StoreStatistics(statistics operations.Statistics) error {
	return m.StoreStatisticsFunc(statistics)
}

// DeleteStatistics godoc
func (m *MockStatisticsRepo) DeleteStatistics(from, to time.Time) error {
	return m.DeleteStatisticsFunc(from, to)
}

func Test_statisticsBucket_createNewBucket(t *testing.T) {
	type fields struct {
		StatisticsRepo  db.StatisticsRepo
		Statistics      *operations.Statistics
		bucketTimer     *time.Ticker
		uniqueSequences map[string]bool
		logger          keptn.LoggerInterface
		lock            sync.Mutex
		cutoffTime      time.Time
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "create statistics bucket - initially nil",
			fields: fields{
				StatisticsRepo:  nil,
				Statistics:      nil,
				bucketTimer:     nil,
				uniqueSequences: nil,
				logger:          nil,
				lock:            sync.Mutex{},
				cutoffTime:      time.Time{},
			},
		},
		{
			name: "create statistics bucket - replace previous bucket",
			fields: fields{
				StatisticsRepo: nil,
				Statistics: &operations.Statistics{
					From: time.Time{},
					To:   time.Time{},
					Projects: map[string]*operations.Project{
						"my-project": &operations.Project{
							Name: "my-project",
						},
					},
				},
				bucketTimer: nil,
				uniqueSequences: map[string]bool{
					"test-context": true,
				},
				logger:     nil,
				lock:       sync.Mutex{},
				cutoffTime: time.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &statisticsBucket{
				StatisticsRepo:  tt.fields.StatisticsRepo,
				Statistics:      tt.fields.Statistics,
				bucketTimer:     tt.fields.bucketTimer,
				uniqueSequences: tt.fields.uniqueSequences,
				logger:          tt.fields.logger,
				lock:            tt.fields.lock,
				cutoffTime:      tt.fields.cutoffTime,
			}
			sb.createNewBucket()

			if sb.Statistics == nil {
				t.Error("createNewBucket() failed: Statistics == nil")
			}
			if len(sb.Statistics.Projects) > 0 {
				t.Errorf("Statistics have not been replaced properly. Got length = %d", len(sb.Statistics.Projects))
			}
			if len(sb.uniqueSequences) > 0 {
				t.Errorf("uniqueSequuences have not been replaced properly. Got length = %d", len(sb.uniqueSequences))
			}
		})
	}
}

func Test_statisticsBucket_storeCurrentBucket(t *testing.T) {
	type fields struct {
		StatisticsRepo  *MockStatisticsRepo
		Statistics      *operations.Statistics
		bucketTimer     *time.Ticker
		uniqueSequences map[string]bool
		logger          keptn.LoggerInterface
		lock            sync.Mutex
		cutoffTime      time.Time
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Store current bucket",
			fields: fields{
				StatisticsRepo: &MockStatisticsRepo{
					GetStatisticsFunc:    nil,
					StoreStatisticsFunc:  nil,
					DeleteStatisticsFunc: nil,
				},
				Statistics: &operations.Statistics{
					From: time.Time{},
					To:   time.Time{},
					Projects: map[string]*operations.Project{
						"my-project": &operations.Project{
							Name: "my-project",
						},
					},
				},
				bucketTimer:     nil,
				uniqueSequences: nil,
				logger:          keptn.NewLogger("", "", ""),
				lock:            sync.Mutex{},
				cutoffTime:      time.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sb := &statisticsBucket{
				StatisticsRepo:  tt.fields.StatisticsRepo,
				Statistics:      tt.fields.Statistics,
				bucketTimer:     tt.fields.bucketTimer,
				uniqueSequences: tt.fields.uniqueSequences,
				logger:          tt.fields.logger,
				lock:            tt.fields.lock,
				cutoffTime:      tt.fields.cutoffTime,
			}
			tt.fields.StatisticsRepo.StoreStatisticsFunc = func(statistics operations.Statistics) error {
				diff := deep.Equal(statistics, *sb.Statistics)
				if len(diff) > 0 {
					t.Error("StatisticsRepo did not receive expected value")
					for _, d := range diff {
						t.Log(d)
					}

				}
				return nil
			}
			sb.storeCurrentBucket()
		})
	}
}

func Test_statisticsBucket_AddEvent(t *testing.T) {
	type fields struct {
		StatisticsRepo  db.StatisticsRepo
		Statistics      *operations.Statistics
		bucketTimer     *time.Ticker
		uniqueSequences map[string]bool
		logger          keptn.LoggerInterface
		lock            sync.Mutex
		cutoffTime      time.Time
	}
	type args struct {
		event operations.Event
	}
	tests := []struct {
		name                    string
		fields                  fields
		args                    args
		expectedStatistics      *operations.Statistics
		expectedUniqueSequences map[string]bool
	}{
		{
			name: "Add event to empty bucket",
			fields: fields{
				StatisticsRepo: nil,
				Statistics: &operations.Statistics{
					From:     time.Time{},
					To:       time.Time{},
					Projects: nil,
				},
				bucketTimer:     nil,
				uniqueSequences: map[string]bool{},
				logger:          keptn.NewLogger("", "", ""),
				lock:            sync.Mutex{},
				cutoffTime:      time.Time{},
			},
			args: args{
				event: operations.Event{
					Data: operations.KeptnBase{
						Project: "my-project",
						Service: "my-service",
					},
					Shkeptncontext: "my-context",
					Type:           "my-type",
				},
			},
			expectedStatistics: &operations.Statistics{
				From: time.Time{},
				To:   time.Time{},
				Projects: map[string]*operations.Project{
					"my-project": {
						Name: "my-project",
						Services: map[string]*operations.Service{
							"my-service": {
								Name:              "my-service",
								ExecutedSequences: 1,
								Events: map[string]int{
									"my-type": 1,
								},
							},
						},
					},
				},
			},
			expectedUniqueSequences: map[string]bool{
				"my-context": true,
			},
		},
		{
			name: "Add event to existing bucket",
			fields: fields{
				StatisticsRepo: nil,
				Statistics: &operations.Statistics{
					From: time.Time{},
					To:   time.Time{},
					Projects: map[string]*operations.Project{
						"my-project": {
							Name: "my-project",
							Services: map[string]*operations.Service{
								"my-service": {
									Name:              "my-service",
									ExecutedSequences: 1,
									Events: map[string]int{
										"my-type": 1,
									},
								},
							},
						},
					},
				},
				bucketTimer:     nil,
				uniqueSequences: map[string]bool{},
				logger:          keptn.NewLogger("", "", ""),
				lock:            sync.Mutex{},
				cutoffTime:      time.Time{},
			},
			args: args{
				event: operations.Event{
					Data: operations.KeptnBase{
						Project: "my-project",
						Service: "my-service",
					},
					Shkeptncontext: "my-context",
					Type:           "my-type",
				},
			},
			expectedStatistics: &operations.Statistics{
				From: time.Time{},
				To:   time.Time{},
				Projects: map[string]*operations.Project{
					"my-project": {
						Name: "my-project",
						Services: map[string]*operations.Service{
							"my-service": {
								Name:              "my-service",
								ExecutedSequences: 2,
								Events: map[string]int{
									"my-type": 2,
								},
							},
						},
					},
				},
			},
			expectedUniqueSequences: map[string]bool{
				"my-context": true,
			},
		},
		{
			name: "Add event to existing bucket for second event of same context",
			fields: fields{
				StatisticsRepo: nil,
				Statistics: &operations.Statistics{
					From: time.Time{},
					To:   time.Time{},
					Projects: map[string]*operations.Project{
						"my-project": {
							Name: "my-project",
							Services: map[string]*operations.Service{
								"my-service": {
									Name:              "my-service",
									ExecutedSequences: 1,
									Events: map[string]int{
										"my-type": 1,
									},
								},
							},
						},
					},
				},
				bucketTimer: nil,
				uniqueSequences: map[string]bool{
					"my-context": true,
				},
				logger:     keptn.NewLogger("", "", ""),
				lock:       sync.Mutex{},
				cutoffTime: time.Time{},
			},
			args: args{
				event: operations.Event{
					Data: operations.KeptnBase{
						Project: "my-project",
						Service: "my-service",
					},
					Shkeptncontext: "my-context",
					Type:           "my-type",
				},
			},
			expectedStatistics: &operations.Statistics{
				From: time.Time{},
				To:   time.Time{},
				Projects: map[string]*operations.Project{
					"my-project": {
						Name: "my-project",
						Services: map[string]*operations.Service{
							"my-service": {
								Name:              "my-service",
								ExecutedSequences: 1,
								Events: map[string]int{
									"my-type": 2,
								},
							},
						},
					},
				},
			},
			expectedUniqueSequences: map[string]bool{
				"my-context": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &statisticsBucket{
				StatisticsRepo:  tt.fields.StatisticsRepo,
				Statistics:      tt.fields.Statistics,
				bucketTimer:     tt.fields.bucketTimer,
				uniqueSequences: tt.fields.uniqueSequences,
				logger:          tt.fields.logger,
				lock:            tt.fields.lock,
				cutoffTime:      tt.fields.cutoffTime,
			}

			sb.AddEvent(tt.args.event)

			diffStatistics := deep.Equal(sb.Statistics, tt.expectedStatistics)
			if len(diffStatistics) > 0 {
				for _, diff := range diffStatistics {
					t.Error("AddEvent() failed: did not get expected Statistics")
					t.Log(diff)
				}
			}

			diffUniqueSequences := deep.Equal(sb.uniqueSequences, tt.expectedUniqueSequences)
			if len(diffUniqueSequences) > 0 {
				t.Error("AddEvent() failed: did not get expected uniqueSequences")
				for _, diff := range diffUniqueSequences {
					t.Log(diff)
				}
			}
		})
	}
}
