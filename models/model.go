package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"msg/common/log"
	"msg/conf"
	"msg/constants"
	"os"
	"time"
)

var DB *gorm.DB

type Timestamp struct {
	CreatedAt time.Time      `sql:"index" gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time      `sql:"index" gorm:"comment:'更新时间'" json:"updated_at"`
	DeletedAt gorm.DeletedAt `sql:"index" gorm:"comment:'删除时间'" json:"deleted_at" swaggerignore:"true"`
}

type Model struct {
	ID string `gorm:"primaryKey;type:bigint AUTO_INCREMENT;comment:'ID'" json:"id" validate:"int64"`
}

type ExtCorpModel struct {
	// ID
	ID string `json:"id" gorm:"primaryKey;type:bigint;comment:'ID'" validate:"int64"`
	// ExtCorpID 外部企业ID
	ExtCorpID string `json:"ext_corp_id" gorm:"index;type:char(18);comment:外部企业ID" validate:"ext_corp_id"`
	// ExtCreatorID 创建者外部员工ID
	ExtCreatorID string `json:"ext_creator_id" gorm:"index;type:char(32);comment:创建者外部员工ID" validate:"word"`
}

// RefModel 关联表基本模型，ID仅用做唯一键，使用组合字段作为主键，方便去重，可实现Association replace保留原纪录
type RefModel struct {
	// ID
	ID string `json:"id" gorm:"unique;type:bigint;comment:'ID'" validate:"int64"`
	// ExtCorpID 外部企业ID
	ExtCorpID string `json:"ext_corp_id" gorm:"index;type:char(18);comment:外部企业ID" validate:"ext_corp_id"`
}

//InitDB 初始化数据库连接
func InitDB(c conf.DBConfig) (db *gorm.DB) {
	var err error

	gormLogLevel := logger.Warn
	if conf.Settings.App.Env == constants.DEV {
		gormLogLevel = logger.Info
	}

	db, err = gorm.Open(
		mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
			c.User,
			c.Password,
			c.Host,
			c.Name)),
		&gorm.Config{
			SkipDefaultTransaction: false,
			Logger:                 logger.Default.LogMode(gormLogLevel),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
			}},
	)
	if err != nil {
		log.Sugar.Error("models.Setup failed", "err", err, "conf", c)
		os.Exit(1)
	}
	if conf.Settings.App.AutoMigration {
		err = db.AutoMigrate(&ChatMsgContent{}, ChatMsg{})
		if err != nil {
			log.Sugar.Error("model auto migrate failed", "err", err, "conf", c)
			os.Exit(1)
		}
	}

	return db
}
