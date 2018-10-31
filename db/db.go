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

// Insert save a string into db
func Insert(s string) {
	fishes := []byte("fishes")
	db.RPush(fishes, []byte(s))
}

// SelectAll get all strings in db
func SelectAll() []string {
	fishes := []byte("fishes")
	nFish, _ := db.LLen(fishes)
	datas, _ := db.LRange(fishes, 0, int32(nFish))

	strs := []string{}
	for _, data := range datas {
		strs = append(strs, string(data))
	}

	return strs
}
