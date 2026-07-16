package entities

import "time"

type User struct {
	ID       uint    `gorm:"primaryKey" json:"-"`
	Email    string  `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username string  `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Password *string `gorm:"default:null" json:"-"`
	Name     string  `gorm:"size:127" json:"name"`
	Avatar   string  `gorm:"size:256" json:"avatar"`
	Active   bool    `json:"active"`

	CreatedBy uint
	UpdatedBy uint
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user_user"
}

func (u *User) HasPassword() bool {
	return u.Password != nil
}
