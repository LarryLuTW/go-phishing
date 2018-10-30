package db

import (
	"log"

	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var db *ledis.DB

// Connect create a db connection
func Connect() {
	cfg := config.NewConfigDefault()
	cfg.DataDir = "./db_data"

	l, _ := ledis.Open(cfg)
	_db, err := l.Select(0)

	if err != nil {
		panic(err)
	}

	db = _db
	log.Println("Connect to db successfully")
}
