package updater

import (
	"andrew/sshman/internal/buildInfo"
	"fmt"
	"testing"
)

func TestGetLatestReleaseVersion(t *testing.T) {
	expected := fmt.Sprintf("v%v.%v.%v", buildInfo.BuildMajor, buildInfo.BuildMinor, buildInfo.BuildPatch)
	release, err := getLatestReleaseVersion()
	if err != nil {
		t.Error(err)
	}
	if release == "" {
		t.Error("got an empty string")
	}
	if release != expected {
		t.Errorf("got %q, expected %q", release, expected)
	}
}

func TestGetCurrentBuildVersion(t *testing.T) {
	expected := fmt.Sprintf("v%v.%v.%v", buildInfo.BuildMajor, buildInfo.BuildMinor, buildInfo.BuildPatch)
	if getCurrentBuildVersion() != expected {
		t.Errorf("Expected string to be %v but got %v", expected, getCurrentBuildVersion())
	}
}
