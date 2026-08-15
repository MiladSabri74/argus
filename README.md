# YARA Downloader

A Go-based CLI application that downloads the latest YARA Forge rules from GitHub, creates a date-tagged folder, extracts only .yar files, and manages updates by checking existing downloads.

## Features

- Downloads the latest YARA Forge rules package from GitHub
- Creates folders with date-based tags (YYYY-MM-DD format)
- Extracts only `.yar` files from the downloaded package
- Skips download if the rules are already up-to-date
- INI-based configuration with conf.d override support
- Command-line argument for custom config directory
- Debian package support for easy installation

## Project Structure

```
yara-downloader/
├── main.go                 # Main entry point
├── pkg/
│   ├── config/            # Configuration management (INI-based)
│   │   └── config.go
│   └── downloader/        # GitHub downloader interface & implementation
│       └── github_downloader.go
├── config/                # Configuration directory
│   ├── yara-downloader.ini    # Main configuration file
│   └── conf.d/            # Override configuration files
├── debian/                # Debian packaging files
│   ├── rules
│   ├── control
│   ├── changelog
│   ├── compat
│   └── install
└── yara-rules/           # Downloaded YARA rules storage
```

## Installation

### Build from Source

```bash
go build -o yara-downloader .
```

### Install Debian Package

```bash
# Build the Debian package
dpkg-buildpackage -us -uc

# Install the generated package
sudo dpkg -i ../yara-downloader_1.0.0_amd64.deb
```

## Usage

### Default Usage

Run the downloader with default settings:

```bash
./yara-downloader
```

This will:
1. Load configuration from `./config/yara-downloader.ini` (or use defaults)
2. Check for the latest release from YARAHQ/yara-forge
3. Create a folder with the release date (e.g., `yara-rules/2024-08-15`)
4. Download and extract only `.yar` files to that folder
5. Skip download if the folder already exists (rules are up-to-date)

### Custom Config Directory

Use the `-c` flag to specify a custom configuration directory:

```bash
./yara-downloader -c /etc/yara-downloader
./yara-downloader -c ./config
```

### Configuration

The application uses INI-based configuration with support for override files.

#### Main Configuration File

Create a `config/yara-downloader.ini` file:

```ini
[github]
repo_owner = YARAHQ
repo_name = yara-forge
api_url = https://api.github.com/repos/%s/%s/releases/latest

[download]
folder = yara-rules
package_pattern = full.zip
```

#### Configuration Options

| Section | Field | Description | Default |
|---------|-------|-------------|---------|
| `[github]` | `repo_owner` | GitHub repository owner | `YARAHQ` |
| `[github]` | `repo_name` | GitHub repository name | `yara-forge` |
| `[github]` | `api_url` | GitHub API URL template | `https://api.github.com/repos/%s/%s/releases/latest` |
| `[download]` | `folder` | Local folder to store rules | `yara-rules` |
| `[download]` | `package_pattern` | Pattern to match in release assets | `full.zip` |

#### Override Files (conf.d)

Files in the `conf.d/` directory are loaded in alphabetical order and override settings from the main configuration file. This allows you to:

1. Keep default settings in the main file
2. Create environment-specific overrides
3. Manage different configurations without modifying the main file

Example override file `conf.d/01-custom.ini`:

```ini
[github]
repo_owner = MyOrg
repo_name = my-yara-rules

[download]
folder = /opt/yara-rules
```

## Architecture

The project follows a clean architecture with separation of concerns:

- **pkg/downloader**: Contains the `Downloader` interface and `GitHubDownloader` implementation
- **pkg/config**: Handles INI-based configuration loading with conf.d support
- **main.go**: Application entry point and orchestration
- **debian/**: Debian packaging files for creating .deb packages

## Building Debian Package

To build a Debian package for distribution:

```bash
# Update debian/changelog with your version
# Edit debian/control with your maintainer information

# Build the package
dpkg-buildpackage -us -uc

# The package will be created in the parent directory
```

## License

See LICENSE file for details.