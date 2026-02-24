package updater

import (
	"andrew/sshman/internal/buildInfo"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	latestReleaseURL = "https://api.github.com/repos/andrewcomputsci2019/sshman/releases/latest"
)

type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
}

func CheckForUpdate() UpdateInfo {
	// function checks GitHub releases for new versions that exist
	panic("todo")
	return UpdateInfo{}
}

func UpdateApplication() error {
	panic("todo")
	return nil
	// this function will download the new binary
	// compare checksum, extract and replace the current binary and return nil if all works
}

func getCurrentBuildVersion() string {
	return fmt.Sprintf("v%v.%v.%v", buildInfo.BuildMajor, buildInfo.BuildMinor, buildInfo.BuildPatch)
}

func getLatestReleaseVersion() (string, error) {
	// query https://api.github.com/repos/andrewcomputsci2019/sshman/releases/latest
	ctx, cFun := context.WithTimeout(context.Background(), 10*time.Second)
	defer cFun()
	req, err := http.NewRequestWithContext(ctx, "GET", latestReleaseURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot get latest release version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot get latest release version, status code: %d", resp.StatusCode)
	}
	if resp.ContentLength > 1e+6 {
		return "", fmt.Errorf("content length too large, expected at most 1e+6, got %d", resp.ContentLength)
	}
	dec := json.NewDecoder(resp.Body)
	tok, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("cannot get latest release version: %w", err)
	}
	if tok != json.Delim('{') {
		return "", fmt.Errorf("cannot get latest release version, expected '{' opening character")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return "", fmt.Errorf("cannot get latest release version: %w", err)
		}
		key := keyTok.(string)
		switch key {
		case "tag_name", "name": // this holds the tag and or release name these will be the same
			var value string
			if err := dec.Decode(&value); err != nil {
				return "", fmt.Errorf("cannot get latest release version: %w", err)
			}
			return value, nil
		default:
			var discard json.RawMessage
			if err := dec.Decode(&discard); err != nil {
				return "", fmt.Errorf("cannot get latest release version: %w", err)
			}
		}
	}
	return "", fmt.Errorf("cannot get latest release version, as it was not found in the reponse JSON")
}
