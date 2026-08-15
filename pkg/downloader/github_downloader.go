package downloader

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

// GitHubRelease represents a release from GitHub API
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Config holds the configuration for the downloader
type Config struct {
	RepoOwner      string
	RepoName       string
	GitHubAPIURL   string
	DownloadFolder string
	PackagePattern string // e.g., "full.zip"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		RepoOwner:      "YARAHQ",
		RepoName:       "yara-forge",
		GitHubAPIURL:   "https://api.github.com/repos/%s/%s/releases/latest",
		DownloadFolder: "yara-rules",
		PackagePattern: "full.zip",
	}
}

// Downloader defines the interface for downloading YARA rules
type Downloader interface {
	GetLatestRelease() (*GitHubRelease, error)
	DownloadPackage(release *GitHubRelease, targetFolder string) (int, error)
}

// GitHubDownloader implements the Downloader interface
type GitHubDownloader struct {
	config *Config
	client *http.Client
}

// NewGitHubDownloader creates a new GitHubDownloader with the given config
func NewGitHubDownloader(cfg *Config) *GitHubDownloader {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &GitHubDownloader{
		config: cfg,
		client: &http.Client{},
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (d *GitHubDownloader) GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf(d.config.GitHubAPIURL, d.config.RepoOwner, d.config.RepoName)

	resp, err := d.client.Get(url)
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

// DownloadPackage downloads and extracts .yar files from the release
func (d *GitHubDownloader) DownloadPackage(release *GitHubRelease, targetFolder string) (int, error) {
	// Find the full rules package
	var fullPackageURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, d.config.PackagePattern) {
			fullPackageURL = asset.BrowserDownloadURL
			break
		}
	}

	if fullPackageURL == "" {
		return 0, fmt.Errorf("no full rules package found in release %s", release.TagName)
	}

	fmt.Printf("Downloading YARA rules for release %s...\n", release.TagName)

	// Download the zip file to temp location
	tempZipPath := filepath.Join(os.TempDir(), fmt.Sprintf("yara-forge-%s.zip", release.TagName))
	defer os.Remove(tempZipPath) // Clean up temp file

	if err := d.downloadFile(fullPackageURL, tempZipPath); err != nil {
		return 0, fmt.Errorf("failed to download package: %w", err)
	}

	fmt.Println("Download complete. Extracting .yar files...")

	// Create target folder
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return 0, fmt.Errorf("failed to create target folder: %w", err)
	}

	// Extract only .yar files
	yarCount, err := d.extractYarFiles(tempZipPath, targetFolder)
	if err != nil {
		return 0, fmt.Errorf("failed to extract files: %w", err)
	}

	return yarCount, nil
}

// downloadFile downloads a file from URL to destPath
func (d *GitHubDownloader) downloadFile(url, destPath string) error {
	resp, err := d.client.Get(url)
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

// extractYarFiles extracts only .yar files from a zip archive
func (d *GitHubDownloader) extractYarFiles(zipPath, destFolder string) (int, error) {
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
