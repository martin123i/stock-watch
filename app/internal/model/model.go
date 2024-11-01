package model

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique"`
	Password  string
	Portfolio []Portfolio `gorm:"foreignKey:UserID"`
}

// Portfolio model
type Portfolio struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"index"`
	Symbol string `gorm:"index"`
}
