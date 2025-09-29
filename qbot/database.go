package qbot

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var PsqlDB *gorm.DB = nil
var PsqlConnected bool = false

type Users struct {
	UserID     uint64 `gorm:"primaryKey;column:user_id"`
	Name       string `gorm:"not null;column:name"`
	Nickname   string `gorm:"column:nick_name"`
	Summary    string `gorm:"column:summary"`
	TokenUsage uint64 `gorm:"column:token_usage"`
}

type Messages struct {
	MsgID   uint64    `gorm:"primaryKey;column:msg_id"`
	UserID  uint64    `gorm:"not null;column:user_id"`
	GroupID uint64    `gorm:"not null;column:group_id"`
	Content string    `gorm:"not null;column:content"`
	Raw     string    `gorm:"not null;column:raw"`
	Deleted bool      `gorm:"column:deleted"`
	IsCmd   bool      `gorm:"column:is_cmd"`
	Time    time.Time `gorm:"not null;column:time"`
}

type UserEvents struct {
	UserID    uint64    `gorm:"primaryKey;column:user_id"`
	EventIdx  int       `gorm:"primaryKey;column:event_idx"`
	MsgRegex  string    `gorm:"not null;column:msg_regex"`
	ReplyText string    `gorm:"not null;column:reply_text"`
	RandProb  float32   `gorm:"not null;column:rand_prob;default:1.0"`
	CreatedAt time.Time `gorm:"not null;column:created_at;default:now()"`
}

type GroupRconConfigs struct {
	GroupID  uint64 `gorm:"primaryKey;column:group_id"`
	Address  string `gorm:"not null;column:address"`
	Password string `gorm:"not null;column:password"`
	Enabled  bool   `gorm:"not null;column:enabled;default:false"`
}

type LegacyGame struct {
	UserID  uint64 `gorm:"primaryKey;column:user_id"`
	Energy  int    `gorm:"not null;column:energy;default:0"`
	Balance int    `gorm:"not null;column:balance;default:0"`
}

func initPsqlDB(dsn string) error {
	var err error
	if PsqlDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return err
	}
	PsqlConnected = true
	return PsqlDB.AutoMigrate(&Users{}, &Messages{}, &UserEvents{}, &GroupRconConfigs{}, &LegacyGame{})
}

func SaveDatabase(msg *Message, isCmd bool) error {
	return PsqlDB.Transaction(func(tx *gorm.DB) error {
		user := Users{
			UserID:   msg.UserID,
			Name:     msg.Nickname,
			Nickname: msg.Card,
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
		newMessage := Messages{
			MsgID:   msg.MsgID,
			UserID:  msg.UserID,
			GroupID: msg.GroupID,
			Content: msg.Content,
			Raw:     msg.Raw,
			Deleted: false,
			IsCmd:   isCmd,
			Time:    time.Unix(int64(msg.Time), 0),
		}
		if err := tx.Create(&newMessage).Error; err != nil {
			return err
		}
		return nil
	})
}
