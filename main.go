package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"yara-downloader/pkg/config"
	"yara-downloader/pkg/downloader"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line arguments
	configPath := flag.String("c", "", "Path to config directory (default: ./config)")
	flag.Parse()

	// Determine config directory and file path
	var configDir string
	if *configPath == "" {
		// Use default config directory
		configDir = "./config"
	} else {
		configDir = *configPath
	}

	configFile := filepath.Join(configDir, "yara-downloader.ini")
	confDir := filepath.Join(configDir, "conf.d")

	// Load configuration
	configManager := config.NewManager(configFile, confDir)

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create downloader with config
	dl := downloader.NewGitHubDownloader(cfg)

	// Get the latest release information
	release, err := dl.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	tagName := release.TagName
	dateTag := formatTagToDate(tagName)
	targetFolder := filepath.Join(cfg.DownloadFolder, dateTag)

	// Check if folder already exists
	if _, err := os.Stat(targetFolder); err == nil {
		fmt.Printf("Folder %s already exists. Rules are up to date.\n", targetFolder)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking folder: %w", err)
	}

	// Download and extract YARA rules
	yarCount, err := dl.DownloadPackage(release, targetFolder)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	fmt.Printf("Successfully extracted %d .yar files to %s\n", yarCount, targetFolder)
	return nil
}

func formatTagToDate(tag string) string {
	// Tag format is YYYYMMDD, convert to YYYY-MM-DD
	if len(tag) >= 8 {
		return fmt.Sprintf("%s-%s-%s", tag[:4], tag[4:6], tag[6:8])
	}
	return tag
}
