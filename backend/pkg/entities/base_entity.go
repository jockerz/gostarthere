package entities

import "time"

// Database base model
type BaseModel struct {
	CreatedBy uint
	UpdatedBy uint
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
