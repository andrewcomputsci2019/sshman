package flags

import "testing"

func TestNewStringFlag(t *testing.T) {
	test1 := NewStringSettableFlag("test", "empty", "flag is used for testing")
	valid, err := test1.conv("thing")
	if err != nil {
		t.Fatal(err)
	}
	if valid != "thing" {
		t.Fatalf("strconv did not return correct result value")
	}
	if test1.Value != "empty" {
		t.Fatalf("Defualt value should be the string \"empty\"")
	}
}

func TestNewUintSettableFlag(t *testing.T) {
	flag := NewUintSettableFlag("uint-test", 0, "uint flag is used for testing")
	valid, err := flag.conv("10")
	if err != nil {
		t.Fatalf("unexpected error converting uint flag: %v", err)
	}
	if valid != 10 {
		t.Fatalf("expected converted value 10, got %d", valid)
	}
	zero, err := flag.conv("")
	if err != nil {
		t.Fatalf("unexpected error converting empty string: %v", err)
	}
	if zero != 0 {
		t.Fatalf("expected zero value for empty string, got %d", zero)
	}
	if _, err := flag.conv("-1"); err == nil {
		t.Fatalf("expected error for negative input")
	}
}

func TestNewIntSettableFlag(t *testing.T) {
	flag := NewIntSettableFlag("int-test", 1, "int flag is used for testing")
	valid, err := flag.conv("15")
	if err != nil {
		t.Fatalf("unexpected error converting int flag: %v", err)
	}
	if valid != 15 {
		t.Fatalf("expected converted value 15, got %d", valid)
	}
	zero, err := flag.conv("")
	if err != nil {
		t.Fatalf("unexpected error converting empty string: %v", err)
	}
	if zero != 0 {
		t.Fatalf("expected zero value for empty string, got %d", zero)
	}
	neg, err := flag.conv("-20")
	if err != nil {
		t.Fatalf("unexpected error converting negative value: %v", err)
	}
	if neg != -20 {
		t.Fatalf("expected negative value -20, got %d", neg)
	}
}

func TestStringFlagSet(t *testing.T) {
	flag := NewStringSettableFlag("string-test-set", "empty", "test string")
	err := flag.Set("Test String Set")
	if err != nil {
		t.Fatal(err)
	}
	if !flag.SetByUser {
		t.Fatalf("Flag should report being set by user")
	}
}

func TestUintFlagSet(t *testing.T) {
	flag := NewUintSettableFlag("uint-test-set", 5, "test uint")
	if err := flag.Set("25"); err != nil {
		t.Fatalf("unexpected error setting uint flag: %v", err)
	}
	if !flag.SetByUser {
		t.Fatalf("Flag should report being set by user")
	}
	if flag.Value != 25 {
		t.Fatalf("expected flag value 25, got %d", flag.Value)
	}
	if err := flag.Set("invalid"); err == nil {
		t.Fatalf("expected error when setting invalid uint flag value")
	}
}

func TestIntFlagSet(t *testing.T) {
	flag := NewIntSettableFlag("int-test-set", -1, "test int")
	if err := flag.Set("15"); err != nil {
		t.Fatalf("unexpected error setting int flag: %v", err)
	}
	if !flag.SetByUser {
		t.Fatalf("Flag should report being set by user")
	}
	if flag.Value != 15 {
		t.Fatalf("expected flag value 15, got %d", flag.Value)
	}
	if err := flag.Set("-3"); err != nil {
		t.Fatalf("unexpected error setting negative int flag: %v", err)
	}
	if flag.Value != -3 {
		t.Fatalf("expected flag value -3, got %d", flag.Value)
	}
	if err := flag.Set("invalid"); err == nil {
		t.Fatalf("expected error when setting invalid int flag value")
	}
}
