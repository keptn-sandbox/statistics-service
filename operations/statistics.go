package operations

import (
	"time"
)

// GetStatisticsParams godoc
type GetStatisticsParams struct {
	From time.Time `form:"from" json:"from" time_format:"unix"`
	To   time.Time `form:"to" json:"to" time_format:"unix"`
}

// Statistics godoc
type Statistics struct {
	From     time.Time           `json:"from" bson:"from"`
	To       time.Time           `json:"to" bson:"to"`
	Projects map[string]*Project `json:"projects" bson:"projects"`
}

// Project godoc
type Project struct {
	Name     string              `json:"name" bson:"name"`
	Services map[string]*Service `json:"services" bson:"services"`
}

// Service godoc
type Service struct {
	Name                   string         `json:"name" bson:"name"`
	ExecutedSequences      int            `json:"executedSequences" bson:"executedSequences"`
	Events                 map[string]int `json:"events" bson:"events"`
	KeptnServiceExecutions map[string]int `json:"keptnServiceExecutions" bson:"keptnServiceExecutions"`
}

func (s *Statistics) ensureProjectAndServiceExist(projectName string, serviceName string) {
	s.ensureProjectExists(projectName)
	if s.Projects[projectName].Services[serviceName] == nil {
		s.Projects[projectName].Services[serviceName] = &Service{
			Name:                   serviceName,
			ExecutedSequences:      0,
			Events:                 map[string]int{},
			KeptnServiceExecutions: map[string]int{},
		}
	}
}

func (s *Statistics) ensureProjectExists(projectName string) {
	if s.Projects == nil {
		s.Projects = map[string]*Project{}
	}
	if s.Projects[projectName] == nil {
		s.Projects[projectName] = &Project{
			Name:     projectName,
			Services: map[string]*Service{},
		}
	}
}

// IncreaseEventTypeCount godoc
func (s *Statistics) IncreaseEventTypeCount(projectName, serviceName, eventType string, increment int) {
	s.ensureProjectAndServiceExist(projectName, serviceName)
	service := s.Projects[projectName].Services[serviceName]
	service.Events[eventType] = service.Events[eventType] + increment
}

// IncreaseExecutedSequencesCount godoc
func (s *Statistics) IncreaseExecutedSequencesCount(projectName, serviceName string, increment int) {
	s.ensureProjectAndServiceExist(projectName, serviceName)
	service := s.Projects[projectName].Services[serviceName]
	service.ExecutedSequences = service.ExecutedSequences + increment
}

// IncreaseKeptnServiceExecutionCount godoc
func (s *Statistics) IncreaseKeptnServiceExecutionCount(projectName, serviceName, keptnServiceName string, increment int) {
	s.ensureProjectAndServiceExist(projectName, serviceName)
	service := s.Projects[projectName].Services[serviceName]
	service.KeptnServiceExecutions[keptnServiceName] = service.KeptnServiceExecutions[keptnServiceName] + increment
}

// MergeStatistics godoc
func MergeStatistics(target Statistics, statistics []Statistics) Statistics {
	for _, stats := range statistics {
		for projectName, project := range stats.Projects {
			target.ensureProjectExists(projectName)
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
