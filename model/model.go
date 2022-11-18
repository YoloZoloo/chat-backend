package model

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID             uint   `gorm:"primarykey"`
	Username       string `gorm:"unique"`
	Firstname      string
	Lastname       string
	Icon           string
	AdminAuthority int32
	Password       string
}

type RoomChat struct {
	RoomID uint   `gorm:"primarykey"`
	UserID []uint `gorm:"not null"`
}

type RoomChat_t struct {
	MessageID uint `gorm:"primarykey; autoIncrement"`
	Message   string
	SenderID  uint
	RoomID    uint
	Datetime  time.Time
}

type PrivateChat struct {
	RoomID uint   `gorm:"primarykey"`
	UserID []uint `gorm:"not null"`
}
type PrivateChat_t struct {
	MessageID     uint `gorm:"primarykey; autoIncrement"`
	Message       string
	SenderID      uint
	PrivateRoomID uint
	Datetime      time.Time
}

func Migrate() {
	db := DbInit()
	db.AutoMigrate(&User{})
	db.AutoMigrate(&RoomChat{})
	db.AutoMigrate(&RoomChat_t{})
	db.AutoMigrate(&PrivateChat{})
	db.AutoMigrate(&PrivateChat_t{})
}

func DbInit() *gorm.DB {
	dsn := os.Getenv("GO_CHAT_DB_USERNAME") + ":" + os.Getenv("GO_CHAT_DB_PASSWORD") +
		"@tcp(" +
		os.Getenv("GO_CHAT_DB_HOST") + ":" + os.Getenv("GO_CHAT_DB_PORT") +
		")/" +
		os.Getenv("GO_CHAT_DATABASE")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	return db
}
