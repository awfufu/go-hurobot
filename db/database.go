package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var PsqlDB *gorm.DB = nil
var PsqlConnected bool = false

type dbUsers struct {
	UserID uint64 `gorm:"primaryKey;column:user_id"`
	Name   string `gorm:"not null;column:name"`
}

func (dbUsers) TableName() string {
	return "users"
}

type dbMessages struct {
	MsgID   uint64    `gorm:"primaryKey;column:msg_id"`
	UserID  uint64    `gorm:"not null;column:user_id;index"`
	GroupID uint64    `gorm:"not null;column:group_id;index"`
	Raw     string    `gorm:"not null;column:raw"`
	Time    time.Time `gorm:"not null;column:time;index"`
}

func (dbMessages) TableName() string {
	return "messages"
}

func InitDB() {
	var err error
	// Ensure directory exists
	dbPath := config.Cfg.SQLite.Path
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("failed to create database directory: %v", err)
	}

	if PsqlDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{}); err != nil {
		log.Fatalln(err)
	}
	PsqlConnected = true
	PsqlDB.AutoMigrate(&dbUsers{}, &dbMessages{})
}

func SaveDatabase(msg *qbot.Message) error {
	return PsqlDB.Transaction(func(tx *gorm.DB) error {
		user := dbUsers{
			UserID: msg.UserID,
			Name:   msg.Name,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(
				map[string]any{
					"name": gorm.Expr("EXCLUDED.name"),
				},
			),
		}).Where("users.name <> EXCLUDED.name").Create(&user).Error; err != nil {
			return err
		}
		newMessage := dbMessages{
			MsgID:   uint64(msg.MsgID),
			UserID:  msg.UserID,
			GroupID: msg.GroupID,
			Raw:     msg.Raw,
			Time:    time.Unix(int64(msg.Time), 0),
		}
		if err := tx.Create(&newMessage).Error; err != nil {
			return err
		}
		return nil
	})
}
