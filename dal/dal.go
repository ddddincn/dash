package dal

import (
	"context"
	"dash/config"
	"dash/consts"
	dashLog "dash/log"
	"dash/model/entity"

	"dash/utils/xerr"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 单例模式
var (
	DB     *gorm.DB
	DBType consts.DBType
)

func NewGormDB(conf *config.Config, gormLogger logger.Interface) *gorm.DB {
	var err error
	// 优先使用sqlite
	if conf.SQLite3 != nil && conf.SQLite3.Enable {
		DB, err = initSQLite(conf, gormLogger)
		if err != nil {
			dashLog.Fatal("open SQLite3 error", zap.Error(err))
		}
		DBType = consts.DBTypeSQLite
	} else if conf.MySQL != nil {
		DB, err = initMySQL(conf, gormLogger)
		if err != nil {
			dashLog.Fatal("connect to MySQL error", zap.Error(err))
		}
		DBType = consts.DBTypeMySQL
	} else {
		dashLog.Fatal("no database available")
	}

	if DB == nil {
		dashLog.Fatal("no available database")
	}
	dashLog.Info("connect database success")
	sqlDB, err := DB.DB()
	if err != nil {
		dashLog.Fatal("get database connection error", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(200)
	sqlDB.SetMaxOpenConns(300)
	sqlDB.SetConnMaxIdleTime(time.Hour)
	SetDefault(DB)
	autoMigrate()
	return DB
}

func initSQLite(conf *config.Config, gormLogger logger.Interface) (*gorm.DB, error) {
	sqliteConfig := conf.SQLite3
	if sqliteConfig == nil {
		return nil, xerr.WithMsg(nil, "nil SQLite config")
	}
	dashLog.Info("try to open SQLite3 db", zap.String("path", sqliteConfig.File))
	db, err := gorm.Open(sqlite.Open(sqliteConfig.File), &gorm.Config{
		Logger:                   gormLogger,
		PrepareStmt:              true,
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
	})
	return db, err
}

func initMySQL(conf *config.Config, gormLogger logger.Interface) (*gorm.DB, error) {
	mysqlConfig := conf.MySQL
	if mysqlConfig == nil {
		return nil, xerr.WithMsg(nil, "nil MySQL config")
	}
	dsn := mysqlConfig.Dsn
	dashLog.Info("try connect to MySQL", zap.Any("dsn", mysqlConfig))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                   gormLogger,
		PrepareStmt:              true,
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
	})
	return db, err
}

func autoMigrate() {
	db := DB.Session(&gorm.Session{
		Logger: DB.Logger.LogMode(logger.Warn),
	})
	err := db.AutoMigrate(&entity.Category{}, &entity.Menu{}, &entity.Option{}, &entity.Post{}, &entity.PostCategory{}, &entity.PostTag{}, &entity.Tag{}, &entity.ThemeSetting{}, &entity.User{})
	if err != nil {
		dashLog.Fatal("failed auto migrate db", zap.Error(err))
	}
}

type ctxTransaction struct{}

// GetQueryByCtx 从上下文获取数据库查询实例
func GetQueryByCtx(ctx context.Context) *Query {
	dbI := ctx.Value(ctxTransaction{})

	if dbI != nil {
		db, ok := dbI.(*Query)
		if !ok {
			panic("unexpected context query value type")
		}
		if db != nil {
			return db
		}
	}
	return Q
}

func SetCtxQuery(ctx context.Context, q *Query) context.Context {
	return context.WithValue(ctx, ctxTransaction{}, q)
}

func Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	q := GetQueryByCtx(ctx)
	return q.Transaction(func(tx *Query) error {
		txCtx := SetCtxQuery(ctx, tx)
		return fn(txCtx)
	})
}
