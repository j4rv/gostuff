package sqlite

import (
	// driver for sqlite3

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

func InitDB(dbFileName string) {
	var err error
	db, err = sqlx.Open("sqlite3", dbFileName)
	panicIfErr(err)
	err = db.Ping()
	panicIfErr(err)
}

func CreateTables() {
	createTableWhiteCard()
	createTableBlackCard()
}
