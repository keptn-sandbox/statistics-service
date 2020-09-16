package operations

import (
	"github.com/go-openapi/strfmt"
)

// GetStatisticsParams godoc
type GetStatisticsParams struct {
	From strfmt.DateTime `form:"from" json:"from"`
	To   strfmt.DateTime `form:"to" json:"to"`
}

// GetStatisticsResponse godoc
type GetStatisticsResponse struct {
	From     strfmt.DateTime `json:"from"`
	To       strfmt.DateTime `json:"to"`
	Projects []Project       `json:"projects"`
}

// Project godoc
type Project struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

// Service godoc
type Service struct {
	Name              string         `json:"name"`
	ExecutedSequences int            `json:"executedSequences"`
	Events            map[string]int `json:"events"`
}
