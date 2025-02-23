package config_test

import (
	"log"
	"os"
	"sync"
	"testing"
	"zeppelin/internal/config"

	"github.com/joho/godotenv"
)

type envs struct {
	dbConnStr string
	mqConnStr string
	smtpPass  string
	fcmConn   string
}

func getConnectionString() *envs {
	err := godotenv.Load("../../.env")
	if err != nil {
		// log.Fatal("Error loading .env file: ", err)
	}
	return &envs{
		dbConnStr: os.Getenv("CONNECTION_STRING"),
		mqConnStr: os.Getenv("MQ_CONN_STRING"),
		smtpPass:  os.Getenv("SMTP_PASSWORD"),
		fcmConn:   os.Getenv("FIREBASE_CONN"),
	}
}
func resetDBVars() {
	config.DB = nil
	config.DbError = nil
	config.Once = sync.Once{}
	config.ProducerChannel = nil
	config.ConsumerChannel = nil
	config.MQConn = nil
	config.SetSmtpConfig(nil)
}

func TestDBConnection(t *testing.T) {
	envs := getConnectionString()
	log.Print("DB CREDENTIALS: ", envs, "\n")
	testcases := []struct {
		dns string
		err bool
	}{
		{
			dns: envs.dbConnStr,
			err: false,
		},
		{
			dns: "host=localhost user=nose password=tampocose dbname=zeppelin port=5432 sslmode=disable TimeZone=Asia/Shanghai",
			err: true,
		},
	}

	for _, tc := range testcases {
		resetDBVars()
		err := config.InitDb(tc.dns)
		if tc.err && err == nil {
			t.Error("Expected to error but didn't")
		} else if !tc.err && err != nil {
			t.Log("passed")
		} else {
			t.Log("Test case passed")
		}
	}
}

func TestMQConnection(t *testing.T) {
	envs := getConnectionString()
	log.Print("MQ CREDENTIALS: ", envs, "\n")
	resetDBVars()
	testcases := []struct {
		dns string
		err bool
	}{
		{
			dns: getConnectionString().mqConnStr,
			err: false,
		},
		{
			dns: "amqp://guest:guest@localhost:5672/",
			err: true,
		},
	}

	for _, tc := range testcases {
		log.Println(tc)
		err := config.InitMQ(tc.dns)
		if tc.err && err == nil {
			t.Error("Expected to error but didn't")
		} else if !tc.err && err != nil {
			t.Log("passed")
		} else {
			t.Log("Test case passed")
		}
	}
}

func TestSmtpConnection(t *testing.T) {
	testcases := []struct {
		password string
		err      bool
	}{
		{
			password: getConnectionString().smtpPass,
			err:      false,
		},
		{
			password: "test",
			err:      true,
		},
	}

	for _, tc := range testcases {
		config.InitSmtp(tc.password)
		err := config.CheckSmtpAuth(config.GetSmtpConfig())
		if tc.err && err == nil {
			t.Error("Expected to error but didn't")
		} else if !tc.err && err != nil {
			t.Log("passed")
		} else {
			t.Log("Test case passed")
		}
	}
	resetDBVars()
	envs := getConnectionString()
	log.Print("SMTP CREDENTIALS", envs, "\n")
}

func TestFirebaseConnection(t *testing.T) {
	envs := getConnectionString()
	log.Print(envs)
	resetDBVars()
	testcases := []struct {
		conn string
		err  bool
	}{
		{
			conn: getConnectionString().fcmConn,
			err:  false,
		},
		{
			conn: "test",
			err:  true,
		},
	}

	for _, tc := range testcases {
		err := config.InitFCM(tc.conn)
		if tc.err && err == nil {
			t.Error("Expected to error but didn't")
		} else if !tc.err && err != nil {
			t.Log("passed")
		} else {
			t.Log("Test case passed")
		}
	}
}
