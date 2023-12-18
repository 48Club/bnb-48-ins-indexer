package database

import (
	"strings"
	"time"

	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

type gormLogger struct{}

func (*gormLogger) Printf(format string, v ...interface{}) {
	format = strings.Replace(format, "\n", " ", 1)
	log.Sugar.Infof(format, v...)
}

// NewMysql connect to mysql
func NewMysql() {
	var err error
	_mysql := config.GetConfig().Mysql

	databaseURL := _mysql.Url
	newLogger := logger.New(
		&gormLogger{},
		logger.Config{
			SlowThreshold:             time.Second * time.Duration(_mysql.SlowThreshold),
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	)

	mysqlConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   _mysql.Prefix, // table name prefix, table for `User` would be `t_users`
			SingularTable: true,          // use singular table name, table for `User` would be `user` with this option enabled
		},
		Logger: newLogger,
	}
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       databaseURL, // data source name
		DefaultStringSize:         255,         // default size for string fields
		DisableDatetimePrecision:  true,        // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,        // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,        // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,       // auto configure based on currently MySQL version
	}), &mysqlConfig)
	if err != nil {
		panic(err)
	}
	mysqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	mysqlDB.SetMaxOpenConns(_mysql.MaxOpenConns)
	mysqlDB.SetMaxIdleConns(_mysql.MaxIdleConns)
}

// Mysql get a connection for mysql
func Mysql() *gorm.DB {
	return db
}

// DisconnectMysql disconnect mysql
func DisconnectMysql() {
	mysqlDB, _ := db.DB()
	if mysqlDB != nil {
		_ = mysqlDB.Close()
	}
}
