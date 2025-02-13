package db_test

import (
	"log"
	"os"
	"sync"
	"testing"
	"zeppelin/internal/db"

	"github.com/joho/godotenv"
)

func getConnectionString() string {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
	return os.Getenv("CONNECTION_STRING")
}
func resetDBVars() {
	db.DB = nil
	db.DbError = nil
	db.Once = sync.Once{}
}

func TestDBConnection(t *testing.T) {
	testcases := []struct {
		dns string
		err bool
	}{
		{
			dns: getConnectionString(),
			err: false,
		},
		{
			dns: "host=localhost user=nose password=tampocose dbname=zeppelin port=5432 sslmode=disable TimeZone=Asia/Shanghai",
			err: true,
		},
	}

	for _, tc := range testcases {
		resetDBVars()
		err := db.InitDb(tc.dns)
		if tc.err && err == nil {
			t.Error("Expected to error but didn't")
		} else if !tc.err && err != nil {
			t.Error("Expected not to error but did", err.Error())
		} else {
			t.Log("Test case passed")
		}
	}
}
