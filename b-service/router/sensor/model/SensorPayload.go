package model

// Domain models
// swagger:model SensorPayload
type SensorPayload struct {
	ID          string   `json:"id"          example:"abc-123"`
	SensorType  string   `json:"sensorType"  example:"temperature"`
	Value       float64  `json:"value"       example:"23.5"`
	TimestampMs int64    `json:"timestampMs" example:"1724550000000"`
	Tags        []string `json:"tags,omitempty"`
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
