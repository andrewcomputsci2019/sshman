package sshParser

import (
	"os"
	path "path/filepath"
	"testing"
)

// todo create a test that creats a temp file with a simpile config and make sure DumpCheckSum does work
// user internal version that does not use sys paths
func TestDumpCheckSum(t *testing.T) {
	dir := t.TempDir()
	filename := "testing"
	dumploc := path.Join(dir, "dump", filename)
	os.MkdirAll(path.Dir(dumploc), 0755)
	filename = path.Join(dir, filename)
	os.WriteFile(filename, []byte("Host example.com\n\tUser example\n\tPort 2022\n"), 0644)
	err := dumpCheckSumInternal(filename, dumploc)
	if err != nil {
		t.Fatalf("Checksum dumper failed to create checksum of file. Error %v", err)
	}
}

// todo create a test that creates a file hashes it using DumpChecksum internal version, then uses IsSame internal version
// to see if the checksum works. Next make a change in the temp file and run isSame again. Also check giving a file that hasnt been hashed before
func TestIsSame(t *testing.T) {
	dir := t.TempDir()
	filename := "example_config"
	dumploc := path.Join(dir, "dump", filename)
	os.MkdirAll(path.Dir(dumploc), 0755)
	filename = path.Join(dir, filename)
	os.WriteFile(filename, []byte("Host example.com\n\tUser example\n\tPort 2022\n"), 0644)
	err := dumpCheckSumInternal(filename, dumploc)
	if err != nil {
		t.Fatalf("Failed to create checksum. Error %v", err)
	}
	isSame, err := isSameInternal(filename, dumploc)
	if err != nil {
		t.Fatalf("Failed to verify creation bug")
	}
	if !isSame {
		t.Fatalf("Checksum should be the same")
	}
	//changed the port
	os.WriteFile(filename, []byte("Host example.com\n\tUser example\n\tPort 2021\n"), 0644)
	isSame, err = isSameInternal(filename, dumploc)
	if err != nil {
		t.Fatalf("Failed to recheck checksum after change. Error %v", err)
	}
	if isSame {
		t.Fatalf("Checksum should report not the same but reported the files are the same after change")
	}
}
