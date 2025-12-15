package db

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type DbPermissions struct {
	Command      string `gorm:"primaryKey;column:command"`
	UserDefault  string `gorm:"not null;column:user_default;default:master"`   // guest, admin, master
	GroupDefault string `gorm:"not null;column:group_default;default:disable"` // enable, disable
	AllowUsers   string `gorm:"column:allow_users"`                            // CSV string: "123,456"
	RejectUsers  string `gorm:"column:reject_users"`                           // CSV string: "123,456"
	AllowGroups  string `gorm:"column:allow_groups"`                           // CSV string: "123,456"
	RejectGroups string `gorm:"column:reject_groups"`                          // CSV string: "123,456"
}

func (DbPermissions) TableName() string {
	return "permissions"
}

func (p *DbPermissions) ParseAllowUsers() []uint64 {
	return ParseIDList(p.AllowUsers)
}

func (p *DbPermissions) ParseRejectUsers() []uint64 {
	return ParseIDList(p.RejectUsers)
}

func (p *DbPermissions) ParseAllowGroups() []uint64 {
	return ParseIDList(p.AllowGroups)
}

func (p *DbPermissions) ParseRejectGroups() []uint64 {
	return ParseIDList(p.RejectGroups)
}

func ParseIDList(s string) []uint64 {
	if s == "" {
		return nil
	}
	// Simple split by comma
	// Assuming valid string like "123,456"
	// To be safe against empty parts like "123,,456", we can filter
	// But user said "only digits and commas".
	parts := strings.Split(s, ",")
	res := make([]uint64, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		if val, err := strconv.ParseUint(part, 10, 64); err == nil {
			res = append(res, val)
		}
	}
	return res
}

func JoinIDList(ids []uint64) string {
	if len(ids) == 0 {
		return ""
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatUint(id, 10)
	}
	return strings.Join(strs, ",")
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
	PsqlDB.AutoMigrate(&dbUsers{}, &dbMessages{}, &DbPermissions{})
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

func GetCommandPermission(cmd string) *DbPermissions {
	var perm DbPermissions
	if err := PsqlDB.Where("command = ?", cmd).First(&perm).Error; err != nil {
		return nil
	}
	return &perm
}

func SaveCommandPermission(perm *DbPermissions) error {
	return PsqlDB.Save(perm).Error
}
