package sqlite

import (
	"log"
	"strconv"
	"testing"
)

var conn *Connection

func TestMain(m *testing.M) {
	// set up db connection
	connection, err := CreateAndLoadDB(":memory:")
	if err != nil {
		log.Fatal("failed to create db: " + err.Error())
	}
	conn = connection
	code := m.Run()
	if code != 0 {
		log.Fatal("Failed test, Status code: " + strconv.Itoa(code))
	}
}

func TestInsert(t *testing.T) {
	t.Logf("Testing Host DAO Insert")
	// todo finish testing code here
	_ = NewHostDao(conn)
}
