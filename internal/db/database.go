package db

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/awfufu/go-hurobot/internal/config"
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
	Perm   int    `gorm:"not null;column:perm;default:0"` // 0:guest, 1:admin, 2:master
}

func (dbUsers) TableName() string {
	return "users"
}

func GetUserPerm(userID uint64) int {
	var user dbUsers
	if err := PsqlDB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return 0 // default guest
	}
	return user.Perm
}

func UpdateUserPerm(userID uint64, perm int) error {
	// Updates perm for existing user, or creates user if not exists?
	// If user doesn't exist, we probably shouldn't create them just for perm unless we have a name.
	// But usually this is called for existing users.
	// Safe way: Update using Model with where clause.
	result := PsqlDB.Model(&dbUsers{}).Where("user_id = ?", userID).Update("perm", perm)
	return result.Error
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
	Command           string `gorm:"primaryKey;column:command"`
	UserAllow         int    `gorm:"not null;column:user_allow;default:2"` // 0:guest, 1:admin, 2:master
	SpecialUsers      string `gorm:"column:special_users"`                 // CSV string
	IsWhitelistUsers  int    `gorm:"column:is_users_whitelist;default:0"`  // 0:blacklist, 1:whitelist
	SpecialGroups     string `gorm:"column:special_groups"`                // CSV string. Note: db col "special_group"
	IsWhitelistGroups int    `gorm:"column:is_groups_whitelist;default:0"` // 0:blacklist, 1:whitelist
}

func (DbPermissions) TableName() string {
	return "permissions"
}

func (p *DbPermissions) ParseSpecialUsers() []uint64 {
	return ParseIDList(p.SpecialUsers)
}

func (p *DbPermissions) ParseSpecialGroups() []uint64 {
	return ParseIDList(p.SpecialGroups)
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
	// Prepend cmd_ prefix if not present (internal usage might pass raw name)
	key := cmd
	if !strings.HasPrefix(key, "cmd_") {
		key = "cmd_" + key
	}

	if err := PsqlDB.Where("command = ?", key).First(&perm).Error; err != nil {
		return nil
	}
	return &perm
}

func SaveCommandPermission(perm *DbPermissions) error {
	// Ensure Command field has cmd_ prefix
	if !strings.HasPrefix(perm.Command, "cmd_") {
		perm.Command = "cmd_" + perm.Command
	}
	return PsqlDB.Save(perm).Error
}
