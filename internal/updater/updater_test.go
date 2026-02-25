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

func TestCheckForUpdateNormal(t *testing.T) {
	updateInfo := CheckForUpdate()
	if updateInfo.UpdateAvailable {
		t.Error("UpdateAvailable should be false")
	}
	if updateInfo.LatestVersion != updateInfo.CurrentVersion || updateInfo.CurrentVersion != getCurrentBuildVersion() {
		t.Errorf("updateInfo holds incorrect versioning either latest doest not much current or currentVersion is not correct. updateInfo: %v", updateInfo)
	}

}

func TestCheckForUpdateSpoofVersion(t *testing.T) {
	// fictitious version less than current release
	updateInfo := checkForUpdate("v1.0.0")
	if !updateInfo.UpdateAvailable {
		t.Error("UpdateAvailable should be true")
	}
}

func TestUpdateApplication(t *testing.T) {
	if err := UpdateApplication(true); err != nil {
		t.Error(err)
	}
}
