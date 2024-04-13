package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func ConnectDB() *sql.DB {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal("time load location error:", err.Error())
	}
	c := mysql.Config{
		DBName:    "app",
		User:      "app",
		Passwd:    "app",
		Addr:      "db:3306",
		Net:       "tcp",
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}
	if Db, err = sql.Open("mysql", c.FormatDSN()); err != nil {
		log.Fatal("Db open error:", err.Error())
	}

	return Db
}
