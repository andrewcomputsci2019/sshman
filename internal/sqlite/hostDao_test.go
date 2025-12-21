package sqlite

import (
	"log"
	"slices"
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

	db := NewHostDao(conn)
	err := db.Delete(Host{Host: "Test"})
	if err != nil {
		t.Fatalf("Failed to delete Host: Test from db")
	}
	_, err = db.Get("Test")
	if err == nil {
		t.Fatalf("Host should not be fetchable from db")
	}

}

func TestInsertMany(t *testing.T) {
	host1 := Host{
		Host:      "Host_TEST1",
		Tags:      []string{"TEST1"},
		CreatedAt: time.Now(),
		Notes:     "HOST 1 test",
		Options:   []HostOptions{{Key: "User", Value: "MyUser"}},
	}
	host2 := Host{
		Host:      "Host_TEST2",
		Tags:      []string{"TEST2"},
		CreatedAt: time.Now(),
		Notes:     "HOST 2 Test",
		Options:   []HostOptions{{Key: "User", Value: "MyUser"}},
	}
	db := NewHostDao(conn)
	err := db.InsertMany(host1, host2)
	if err != nil {
		t.Fatalf("Failed to insert multiple host into database. Error: %v", err)
	}

}

func TestUpdateMany(t *testing.T) {

}

func TestUpsertHost(t *testing.T) {

	host1 := Host{
		Host:      "Host_TEST1",
		Tags:      []string{"TEST1", "UPSERT"},
		CreatedAt: time.Now(),
		Notes:     "HOST 1 upsert test",
		Options:   []HostOptions{{Key: "User", Value: "UpsertHost"}},
	}

	db := NewHostDao(conn)
	err := db.InsertOrUpdate(host1)
	if err != nil {
		t.Fatalf("Failed to upsert host into table. err: %v", err)
	}
	updatedHost, err := db.Get("Host_TEST1")
	if err != nil {
		t.Fatalf("failed to grab updated host from db")
	}
	if updatedHost.Tags[1] != "UPSERT" || updatedHost.Options[0].Value != "UpsertHost" || updatedHost.Notes != "HOST 1 upsert test" {
		t.Fatalf("Host was not properly updated. Host: %v", updatedHost)
	}
	log.Printf("Updated host %v", updatedHost)
}

func TestInsertIgnoreConflict(t *testing.T) {
	host1 := Host{
		Host:      "Host_TEST1",
		Tags:      []string{"TEST1"},
		CreatedAt: time.Now(),
		Notes:     "Should not be inserted into db",
		Options:   []HostOptions{{Key: "User", Value: "MyUser"}},
	}
	newHost := Host{
		Host:      "New_Host",
		Tags:      []string{"Test"},
		CreatedAt: time.Now(),
		Notes:     "Should be inserted",
	}
	db := NewHostDao(conn)
	err := db.InsertManyIgnoreConflict(host1, newHost)
	if err != nil {
		t.Fatalf("Failed to insert many host with conflict. Error %v", err)
	}
	validate, err := db.Get("Host_TEST1")
	if err != nil {
		t.Fatalf("Failed to get Host from db")
	}
	if validate.Notes == "Should not be inserted into db" {
		t.Fatalf("Test failed, host should have not been inserted into the table")
	}
}

func TestUpsertMany(t *testing.T) {
	db := NewHostDao(conn)
	newHost, err := db.Get("New_Host")
	if err != nil {
		t.Fatal(err)
	}
	host1, err := db.Get("Host_TEST1")
	if err != nil {
		t.Fatal(err)
	}
	timeStamp := time.Now().Add(time.Second * 2) // give a buffer so clock doesn't return the same time twice
	host1.UpdatedAt = &timeStamp
	newHost.UpdatedAt = &timeStamp

	host1.Tags = append(host1.Tags, "Many")
	newHost.Tags = append(newHost.Tags, "Many")

	err = db.InsertOrUpdateMany(newHost, host1)
	if err != nil {
		t.Fatalf("Failed to upsert multiple host into db")
	}
	valid1, _ := db.Get("New_Host")
	valid2, _ := db.Get("Host_TEST1")

	if slices.Compare(valid1.Tags, newHost.Tags) != 0 || slices.Compare(valid2.Tags, host1.Tags) != 0 {
		t.Fatalf("Host should be updated yet they are not")
	}
	log.Printf("Valid 1 %v", valid1)
	log.Printf("Valid 2 %v", valid2)
	if !valid1.UpdatedAt.After(valid1.CreatedAt) || !valid2.UpdatedAt.After(valid2.CreatedAt) {
		t.Fatalf("Updated at should be newer than created timestamp")
	}

}

func TestGetAll(t *testing.T) {
	hosts, err := NewHostDao(conn).GetAll()
	if err != nil {
		t.Fatalf("Failed to get all hosts. error %v", err)
	}
	t.Logf("Hosts returned by database: %v", hosts)
	if len(hosts) == 0 {
		t.Fatalf("Should be a filled slice")
	}
}

func TestCountAll(t *testing.T) {
	hosts, err := NewHostDao(conn).GetAll()
	if err != nil {
		t.Fatalf("Failed to get all hosts. error %v", err)
	}
	count, err := NewHostDao(conn).Count()
	if err != nil {
		t.Fatalf("Failed to get count of database. Error: %v", err)
	}
	if len(hosts) != int(count) {
		t.Fatalf("db mismatch between count and return hosts. Expected %v, got %v hosts", count, len(hosts))
	}

}
