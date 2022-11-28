package model

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID             uint   `gorm:"primarykey; autoIncrement"`
	Username       string `gorm:"unique"`
	Firstname      string
	Lastname       string
	Icon           string
	AdminAuthority int
	Password       string
}

type RoomChat struct {
	RoomID int  `gorm:"not null"`
	UserID uint `gorm:"not null"`
}

type RoomChat_t struct {
	MessageID uint `gorm:"primarykey; autoIncrement"`
	Message   string
	SenderID  uint
	RoomID    int
	Datetime  time.Time
}

type PrivateChat struct {
	RoomID uint `gorm:"primarykey; autoIncrement"`
	PeerID uint `gorm:"not null"`
	UserId uint `gorm:"not null"`
}
type PrivateChat_t struct {
	MessageID uint `gorm:"primarykey; autoIncrement"`
	Message   string
	SenderID  uint
	RoomID    uint
	Datetime  time.Time
}

func Migrate() {
	db, err := DbInit()
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})
	db.AutoMigrate(&RoomChat{})
	db.AutoMigrate(&RoomChat_t{})
	db.AutoMigrate(&PrivateChat{})
	db.AutoMigrate(&PrivateChat_t{})
}

func DbInit() (*gorm.DB, error) {
	dsn := os.Getenv("GO_CHAT_DB_USERNAME") + ":" + os.Getenv("GO_CHAT_DB_PASSWORD") +
		"@tcp(" +
		os.Getenv("GO_CHAT_DB_HOST") + ":" + os.Getenv("GO_CHAT_DB_PORT") +
		")/" +
		os.Getenv("GO_CHAT_DATABASE")
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{})
	return db, err
}
