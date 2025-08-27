package model

import "time"

// Domain models
// swagger:model SensorPayload
type SensorPayload struct {
	ID1         string  `json:"id1"          example:"abc-123"`
	ID2         int     `json:"id2"          example:"123"`
	SensorType  string  `json:"sensorType"  example:"temperature"`
	Value       float64 `json:"value"       example:"23.5"`
	TimestampMs int64   `json:"timestampMs" example:"1724550000000"`
}

// swagger:model SensorPage
type SensorPage struct {
	Items    []SensorPayload `json:"items"`
	Page     int             `json:"page"     example:"1"`
	PageSize int             `json:"pageSize" example:"50"`
	Total    int64           `json:"total"    example:"1234"`
}

// swagger:model UpdateSensorByIDPayload
type UpdateSensorByIDPayload struct {
	ID          string   `json:"id"          example:"abc-123"`
	Value       *float64 `json:"value,omitempty"       example:"25.1"`
	SensorType  *string  `json:"sensorType,omitempty"  example:"humidity"`
	TimestampMs *int64   `json:"timestampMs,omitempty" example:"1724550000000"`
}

// swagger:model IDCombination
type IDCombination struct {
	ID1 string `json:"id1" validate:"required"`
	ID2 int    `json:"id2" validate:"required"`
}

// swagger:model DeleteByIDsRequest
type DeleteByIDsRequest struct {
	IDCombinations []IDCombination `json:"id_combinations" validate:"required,min=1"`
}

// swagger:model DeleteResponse
type DeleteResponse struct {
	DeletedCount int64  `json:"deleted_count"`
	Message      string `json:"message"`
}

// swagger:model DeleteByTimeRequest
type DeleteByTimeRequest struct {
	FromTime time.Time `json:"from_time" validate:"required" example:"2025-08-25T18:00:24.947000+07:00"`
	ToTime   time.Time `json:"to_time" validate:"required" example:"2025-08-25T19:00:24.947000+07:00"`
}

// swagger:model DeleteByIDAndTimesRequest
type DeleteByIDAndTimesRequest struct {
	IDCombinations []IDCombination `json:"id_combinations" validate:"required,min=1"`
	FromTime       time.Time       `json:"from_time" validate:"required" example:"2025-08-25T18:00:24.947000+07:00"`
	ToTime         time.Time       `json:"to_time" validate:"required" example:"2025-08-25T19:00:24.947000+07:00"`
}
