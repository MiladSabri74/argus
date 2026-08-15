package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	repoOwner      = "YARAHQ"
	repoName       = "yara-forge"
	githubAPIURL   = "https://api.github.com/repos/%s/%s/releases/latest"
	downloadFolder = "yara-rules"
)

// GitHubRelease represents a release from GitHub API
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name                string `json:"name"`
	BrowserDownloadURL  string `json:"browser_download_url"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get the latest release information
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	tagName := release.TagName
	dateTag := formatTagToDate(tagName)
	targetFolder := filepath.Join(downloadFolder, dateTag)

	// Check if folder already exists
	if _, err := os.Stat(targetFolder); err == nil {
		fmt.Printf("Folder %s already exists. Rules are up to date.\n", targetFolder)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking folder: %w", err)
	}

	fmt.Printf("Downloading YARA rules for release %s (%s)...\n", tagName, dateTag)

	// Find the full rules package
	var fullPackageURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "full.zip") {
			fullPackageURL = asset.BrowserDownloadURL
			break
		}
	}

	if fullPackageURL == "" {
		return fmt.Errorf("no full rules package found in release %s", tagName)
	}

	// Download the zip file
	tempZipPath := filepath.Join(os.TempDir(), fmt.Sprintf("yara-forge-%s.zip", tagName))
	defer os.Remove(tempZipPath) // Clean up temp file

	if err := downloadFile(fullPackageURL, tempZipPath); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	fmt.Println("Download complete. Extracting .yar files...")

	// Create target folder
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return fmt.Errorf("failed to create target folder: %w", err)
	}

	// Extract only .yar files
	yarCount, err := extractYarFiles(tempZipPath, targetFolder)
	if err != nil {
		return fmt.Errorf("failed to extract files: %w", err)
	}

	fmt.Printf("Successfully extracted %d .yar files to %s\n", yarCount, targetFolder)
	return nil
}

func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, repoOwner, repoName)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func formatTagToDate(tag string) string {
	// Tag format is YYYYMMDD, convert to YYYY-MM-DD
	if len(tag) >= 8 {
		return fmt.Sprintf("%s-%s-%s", tag[:4], tag[4:6], tag[6:8])
	}
	return tag
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractYarFiles(zipPath, destFolder string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	yarCount := 0

	for _, f := range r.File {
		// Skip directories and non-.yar files
		if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".yar") {
			continue
		}

		// Create destination path
		destPath := filepath.Join(destFolder, filepath.Base(f.Name))

		// Open the file in the zip
		rc, err := f.Open()
		if err != nil {
			return yarCount, err
		}

		// Create the destination file
		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return yarCount, err
		}

		// Copy contents
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return yarCount, err
		}

		yarCount++
	}

	return yarCount, nil
}
