package updater

import (
	"andrew/sshman/internal/buildInfo"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-extract"
)

const (
	latestReleaseURL = "https://api.github.com/repos/andrewcomputsci2019/sshman/releases/latest"
	DownloadURL      = "https://github.com/andrewcomputsci2019/sshman/releases/latest/download/%s"
)

type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
}

func CheckForUpdate() UpdateInfo {
	// function checks GitHub releases for new versions that exist
	currentBuildVersion := getCurrentBuildVersion()
	return checkForUpdate(currentBuildVersion)
}

func checkForUpdate(currentBuildVersion string) UpdateInfo {
	updateInfo := UpdateInfo{}
	latestBuild, err := getLatestReleaseVersion()
	if err != nil {
		slog.Error("Failed to get latest release", "error", err)
		return updateInfo
	}

	updateInfo.CurrentVersion = currentBuildVersion
	updateInfo.LatestVersion = latestBuild

	currentBuildVersion = strings.TrimPrefix(currentBuildVersion, "v")
	latestVersion := strings.TrimPrefix(latestBuild, "v")

	currentVersions := strings.Split(currentBuildVersion, ".")
	if len(currentVersions) != 3 {
		slog.Error("Failed to extract versions from current build", "error", fmt.Sprintf("Expected 3 versions but got %d", len(currentVersions)))
		return UpdateInfo{}
	}
	releaseVersions := strings.Split(latestVersion, ".")
	if len(releaseVersions) != 3 {
		slog.Error("Failed to extract versions from release build", "error", fmt.Sprintf("Expected 3 versions but got %d", len(releaseVersions)))
	}

	order := slices.Compare(currentVersions, releaseVersions)
	if order < 0 {
		updateInfo.UpdateAvailable = true
	}
	return updateInfo
}

func UpdateApplication(dryRun bool) error {
	// this function will download the new binary
	// compare checksum, extract and replace the current binary and return nil if all works
	if buildInfo.BUILD_OS == "windows" {
		return errors.New("windows does not support update function")
	}
	checksumFileName, archiveFileName := getDownloadStrings()
	currentBinaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	currentBinaryPath, err = filepath.EvalSymlinks(currentBinaryPath)
	if err != nil {
		return err
	}
	binDir := filepath.Dir(currentBinaryPath)
	tempDir, err := os.MkdirTemp(binDir, ".sshman-update-*")
	if err != nil {
		return err
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			slog.Warn("Failed to remove temporary directory", "path", path, "error", err)
		}
	}(tempDir)

	checkFile, err := os.Create(filepath.Join(tempDir, checksumFileName))
	if err != nil {
		return err
	}
	defer checkFile.Close()

	archiveFile, err := os.Create(filepath.Join(tempDir, archiveFileName))
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	if err := downloadFile(fmt.Sprintf(DownloadURL, checksumFileName), checkFile); err != nil {
		return err
	}
	if err := downloadFile(fmt.Sprintf(DownloadURL, archiveFileName), archiveFile); err != nil {
		return err
	}

	if err := rewindFile(archiveFile); err != nil {
		return err
	}

	if err := rewindFile(checkFile); err != nil {
		return err
	}

	checker := sha256.New()

	if _, err := io.Copy(checker, archiveFile); err != nil {
		return err
	}

	sumString := hex.EncodeToString(checker.Sum(nil))

	if err := checkChecksumFile(checkFile, sumString, archiveFileName); err != nil {
		return err
	}

	if err := rewindFile(archiveFile); err != nil {
		return err
	}

	if err := extract.Unpack(context.Background(), tempDir, archiveFile, extract.NewConfig()); err != nil {
		return err
	}

	if err := checkFile.Close(); err != nil {
		slog.Warn("Failed to close file", "path", filepath.Join(tempDir, checksumFileName), "error", err)
	}
	if err := archiveFile.Close(); err != nil {
		slog.Warn("Failed to close file", "path", filepath.Join(tempDir, archiveFileName), "error", err)
	}

	if dryRun {
		return os.RemoveAll(tempDir)
	}

	if err := replaceBinaryFile(filepath.Join(tempDir, getExecutableName()), currentBinaryPath); err != nil {
		return err
	}

	if err := os.RemoveAll(tempDir); err != nil {
		slog.Warn("Failed to remove temporary directory", "path", tempDir, "error", err)
	}
	// kinda trippy method but we are going to evoke the new binary which will cause it to restart, this does have the counter
	return syscall.Exec(currentBinaryPath, os.Args, os.Environ())
}

func checkChecksumFile(checkFile *os.File, expectedHash, expectedFile string) error {
	scanner := bufio.NewScanner(checkFile)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}

		if fields[0] == expectedHash && fields[1] == expectedFile {
			return nil
		}
	}

	return fmt.Errorf("checksum mismatch for %s", expectedFile)
}

func downloadFile(url string, destFile *os.File) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status code: %d", resp.StatusCode)
	}
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getDownloadStrings() (string, string) {
	checksumFileName := "checksums.txt"
	archiveFileName := fmt.Sprintf("ssh-man-%s-%s.tar.gz", buildInfo.BUILD_OS, buildInfo.BUILD_ARC)
	return checksumFileName, archiveFileName
}

func getExecutableName() string {
	return "ssh-man"
}

func rewindFile(file *os.File) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}

func replaceBinaryFile(newBinaryPath string, oldBinaryPath string) error {
	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		return err
	}

	if err := os.Rename(newBinaryPath, oldBinaryPath); err != nil {
		return err
	}
	return nil
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
