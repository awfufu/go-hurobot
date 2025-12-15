package db

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var PsqlDB *gorm.DB = nil
var PsqlConnected bool = false

type dbUsers struct {
	UserID     uint64 `gorm:"primaryKey;column:user_id"`
	Name       string `gorm:"not null;column:name"`
	Nickname   string `gorm:"column:nick_name"`
	Summary    string `gorm:"column:summary"`
	TokenUsage uint64 `gorm:"column:token_usage"`
}

type dbMessages struct {
	MsgID   uint64    `gorm:"primaryKey;column:msg_id"`
	UserID  uint64    `gorm:"not null;column:user_id"`
	GroupID uint64    `gorm:"not null;column:group_id"`
	Content string    `gorm:"not null;column:content"`
	Raw     string    `gorm:"not null;column:raw"`
	Deleted bool      `gorm:"column:deleted"`
	Time    time.Time `gorm:"not null;column:time"`
}

type GroupRconConfigs struct {
	GroupID  uint64 `gorm:"primaryKey;column:group_id"`
	Address  string `gorm:"not null;column:address"`
	Password string `gorm:"not null;column:password"`
	Enabled  bool   `gorm:"not null;column:enabled;default:false"`
}

func InitDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Cfg.PostgreSQL.Host,
		strconv.Itoa(int(config.Cfg.PostgreSQL.Port)),
		config.Cfg.PostgreSQL.User,
		config.Cfg.PostgreSQL.Password,
		config.Cfg.PostgreSQL.DbName,
	)
	var err error
	if PsqlDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		log.Fatalln(err)
	}
	PsqlConnected = true
	PsqlDB.AutoMigrate(&dbUsers{}, &dbMessages{}, &GroupRconConfigs{})
}

func SaveDatabase(msg *qbot.Message) error {
	return PsqlDB.Transaction(func(tx *gorm.DB) error {
		user := dbUsers{
			UserID:   msg.UserID,
			Name:     msg.Name,
			Nickname: msg.GroupCard,
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
			Content: msg.Raw,
			Raw:     msg.Raw,
			Time:    time.Unix(int64(msg.Time), 0),
		}
		if err := tx.Create(&newMessage).Error; err != nil {
			return err
		}
		return nil
	})
}
