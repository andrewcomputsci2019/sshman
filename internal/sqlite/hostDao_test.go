package sqlite

import (
	"log"
	"strconv"
	"testing"
	"time"
)

var conn *Connection

func TestMain(m *testing.M) {
	// set up db connection
	connection, err := CreateAndLoadDB(":memory:")
	if err != nil {
		log.Fatal("failed to create db: " + err.Error())
	}
	conn = connection
	defer conn.Close()
	code := m.Run()
	if code != 0 {
		log.Fatal("Failed test, Status code: " + strconv.Itoa(code))
	}
}

func TestInsert(t *testing.T) {
	t.Logf("Testing Host DAO Insert")
	// todo finish testing code here
	db := NewHostDao(conn)
	opts := []HostOptions{{
		Key:   "Hostname",
		Value: "test.local",
	}, {
		Key:   "User",
		Value: "testUser",
	}}
	testingHost := Host{
		Host:      "Test",
		CreatedAt: time.Now(),
		Notes:     "Test host",
		Options:   opts,
		Tags:      []string{"test"},
	}
	err := db.Insert(testingHost)
	if err != nil {
		t.Fatalf("Failed to inset host, host: %s. Error: %s", testingHost.String(), err)
	}
}

func TestGetHost(t *testing.T) {
	db := NewHostDao(conn)
	host, err := db.Get("Test")
	if err != nil {
		t.Fatalf("Failed to get host from db, Host: %s. err: %s", "Test", err)
	}
	log.Printf("Received this host from db: %s", host.String())
	if host.Host != "Test" {
		log.Fatalf("Failed to get correct host, Expected Test but got %s", host.Host)
	}
}

func TestUpdateHost(t *testing.T) {

	db := NewHostDao(conn)
	host, err := db.Get("Test")
	if err != nil {
		t.Fatalf("Failed to get Host from db")
	}
	opts := []HostOptions{{
		Key:   "Hostname",
		Value: "test.local",
	}, {
		Key:   "User",
		Value: "NewUser",
	}}
	host.Options = opts
	updTime := new(time.Time)
	*updTime = time.Now()
	host.UpdatedAt = updTime
	host.Notes += " Updated"
	err = db.Update(host)
	if err != nil {
		t.Fatalf("Failed to update host into table. Host: %s  Error %s", host.String(), err.Error())
	}
	reCheck, err := db.Get("Test")
	if err != nil {
		t.Fatalf("Failed to get host after update. Err %s", err)
	}
	t.Logf("Host received after update %s", reCheck.String())
}

func TestDeleteHost(t *testing.T) {

}

func TestGetAll(t *testing.T) {

}

func TestCountAll(t *testing.T) {

}
