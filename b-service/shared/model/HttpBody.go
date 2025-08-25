package model

type Body[T any] struct {
	Data    T      `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// ========= Swagger-visible envelope =========
// Use this in @Success/@Failure with field overrides to show the real T.
// See: {object} Envelope{data=YourType}
// swagger:model Envelope
type Envelope struct {
	Data    interface{} `json:"data"`
	Error   bool        `json:"error" example:"false"`
	Message string      `json:"message" example:"ok"`
}

// Use when data is intentionally empty (e.g., 204/DELETE) .
// swagger:model Empty
type Empty struct{}
