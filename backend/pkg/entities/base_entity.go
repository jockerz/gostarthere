package entities

import "time"

// Database base model
type BaseModel struct {
	CreatedBy uint
	UpdatedBy uint
	CreatedAt time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime" json:"updated_at"`
}
