# YARA Downloader - INI Configuration Example

This directory contains the configuration files for yara-downloader.

## Structure

- `yara-downloader.ini` - Main configuration file
- `conf.d/` - Directory for override configuration files

## Main Configuration File

The main configuration file (`yara-downloader.ini`) contains:

### [github] Section
- `repo_owner` - GitHub repository owner (default: YARAHQ)
- `repo_name` - GitHub repository name (default: yara-forge)
- `api_url` - GitHub API URL pattern with %s placeholders

### [download] Section
- `folder` - Directory to store downloaded rules (default: yara-rules)
- `package_pattern` - Pattern to match full rules package (default: full.zip)

## Override Files (conf.d)

Files in the `conf.d/` directory are loaded in alphabetical order and override
settings from the main configuration file. This allows you to:

1. Keep default settings in the main file
2. Create environment-specific overrides in conf.d
3. Manage different configurations without modifying the main file

Example override file `01-custom.ini`:
```ini
[github]
repo_owner = MyOrg
repo_name = my-yara-rules

[download]
folder = /opt/yara-rules
```

## Command Line Usage

Use the `-c` flag to specify a custom config directory:

```bash
./yara-downloader -c /etc/yara-downloader
./yara-downloader -c ./config
```

If no `-c` flag is provided, the application uses `./config` as the default.
