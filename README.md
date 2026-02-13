# Zhyper All-in-One Checker V4.8.7

A powerful multi-CMS credential checker with GUI, supporting 12 different CMS platforms with advanced detection and validation capabilities.

## Features

### Supported CMS Platforms (12)
1. **WordPress** - Admin detection, Filemanager, Appearance Editor, Install Plugin capability
2. **Joomla** - CSRF token extraction, protection detection
3. **Drupal** - Login validation
4. **OpenCart** - Admin panel access
5. **Moodle** - Education platform
6. **OJS (Open Journal Systems)** - v2/v3 support, multi-journal scanning, admin role detection
7. **cPanel** - Domain extraction (7 different methods)
8. **WHM** - Web Host Manager
9. **Plesk** - Hosting control panel
10. **DirectAdmin** - Server control panel
11. **phpMyAdmin** - Database management
12. **Adminer** - Lightweight database management

### Key Features
- 🎨 Modern GUI with dark theme using Fyne framework
- 🚀 Multi-threaded checking (1-500 configurable threads)
- 📊 Real-time statistics tracking
- 🎯 Auto-detection of CMS from filename
- 🔍 Deep scanning mode for thorough checks
- 🐛 Debug mode with detailed logging
- 📁 Organized output files by CMS type
- 🎨 Color-coded terminal output
- 🔒 SSL verification bypass for testing
- 🍪 Session cookie management
- 🌐 22 rotating user agents
- ⚡ Concurrent processing with goroutines

## Requirements

- Go 1.21 or higher
- Internet connection
- Linux/Windows/macOS

## Installation

### Clone the Repository
```bash
git clone https://github.com/wppluginsn/golang.git
cd golang
```

### Install Dependencies
```bash
go mod download
```

## Building

### Linux
```bash
go build -o zhyper-checker main.go
chmod +x zhyper-checker
./zhyper-checker
```

### Windows
```bash
go build -o zhyper-checker.exe main.go
zhyper-checker.exe
```

### macOS
```bash
go build -o zhyper-checker main.go
chmod +x zhyper-checker
./zhyper-checker
```

### Cross-Compilation

**For Windows (from Linux/Mac):**
```bash
GOOS=windows GOARCH=amd64 go build -o zhyper-checker.exe main.go
```

**For Linux (from Windows/Mac):**
```bash
GOOS=linux GOARCH=amd64 go build -o zhyper-checker main.go
```

**For macOS (from Linux/Windows):**
```bash
GOOS=darwin GOARCH=amd64 go build -o zhyper-checker main.go
```

## Usage

### Input Format
Create a text file with credentials in the format:
```
url|username|password
```

Example (`combo.txt`):
```
https://example.com/wp-admin|admin|password123
https://site.com/administrator|user|pass456
https://demo.com/cpanel|root|secret789
```

### Running the Application

1. **Launch the GUI:**
   ```bash
   ./zhyper-checker
   ```

2. **Select Your File:**
   - Click the "Browse" button
   - Select your combo file

3. **Configure Settings:**
   - **Threads:** Set number of concurrent checks (1-500, default: 10)
   - **Debug:** Enable for detailed logging
   - **DEEP:** Enable for thorough scanning

4. **Start Checking:**
   - Click "START" button
   - Monitor real-time statistics
   - View colored logs in terminal
   - Click "STOP" to pause

### Output Files

Results are saved in the `result/` directory:

**General:**
- `Good_wordpress.txt` - Valid WordPress credentials
- `Good_joomla.txt` - Valid Joomla credentials
- `Good_drupal.txt` - Valid Drupal credentials
- `Good_opencart.txt` - Valid OpenCart credentials
- `Good_moodle.txt` - Valid Moodle credentials
- `Good_ojs.txt` - Valid OJS credentials
- `Good_cpanel.txt` - Valid cPanel credentials (with extracted domains)
- `Good_whm.txt` - Valid WHM credentials
- `Good_plesk.txt` - Valid Plesk credentials
- `Good_directadmin.txt` - Valid DirectAdmin credentials
- `Good_phpmyadmin.txt` - Valid phpMyAdmin credentials
- `Good_adminer.txt` - Valid Adminer credentials

**WordPress Specific:**
- `Good_WP_Filemanager.txt` - WordPress with filemanager access
- `Good_WP_Appearance.txt` - WordPress with appearance editor
- `Good_WP_InstallPlugin.txt` - WordPress with plugin install capability
- `Good_WP_Access.txt` - WordPress administrator accounts

**OJS Specific:**
- `Good_OJS_Admin.txt` - OJS administrator accounts

### CMS Auto-Detection

The tool automatically detects CMS type from filenames:
- `wordpress_combo.txt` → WordPress
- `wp_sites.txt` → WordPress
- `joomla_list.txt` → Joomla
- `cpanel_hosts.txt` → cPanel
- `ojs_journals.txt` → OJS
- etc.

### GUI Overview

```
┌─────────────────────────────────────────────────┐
│         ZHYPER ALL-IN-ONE V4.8.7                │
├─────────────────────────────────────────────────┤
│  Statistics:                                    │
│  Total: 100 | Done: 50 | Valid: 25             │
│  Fail: 20 | Err: 5                              │
├─────────────────────────────────────────────────┤
│  Terminal Output (Color-coded):                 │
│  ✓ Valid credentials (Green)                    │
│  ✗ Failed attempts (Red)                        │
│  ⚠ Errors (Yellow)                              │
│  ℹ Info messages (Cyan)                         │
├─────────────────────────────────────────────────┤
│  Control Panel:                                 │
│  File: [Browse]                                 │
│  Threads: [10]  ☐ Debug  ☐ DEEP               │
│  [START] [STOP]                                 │
└─────────────────────────────────────────────────┘
```

## Features in Detail

### WordPress Detection
- Admin role verification with scoring system
- Filemanager detection from dashboard HTML
- Appearance editor capability check
- Plugin installation permission check
- Enhanced validation with multiple indicators

### OJS (Open Journal Systems)
- Pattern-based detection with caching
- v2 and v3 support
- Multi-journal scanning
- Admin role detection across journals
- Strict validation with response analysis

### Joomla
- CSRF token extraction (32/64 char patterns)
- FireShield, Cloudflare, reCAPTCHA detection
- Scoring-based validation
- Protection bypass attempts

### cPanel Domain Extraction
7 different extraction methods:
1. Domains page parsing
2. listaccts API
3. domaininfo API
4. Park domain API
5. Addon domain API
6. HTML scraping
7. Domain cleanup and deduplication

### Multi-Threading
- Goroutines for parallel processing
- Thread pool management
- WaitGroup for synchronization
- Mutex for thread-safe operations
- Configurable concurrency (1-500 threads)

### HTTP Client Features
- SSL verification bypass for testing
- Random user agent rotation (22 agents)
- Cookie jar for session persistence
- Configurable timeouts (8-20 seconds)
- Automatic redirect following

## Troubleshooting

### GUI doesn't open
- **Linux:** Install X11 development libraries:
  ```bash
  sudo apt-get install libgl1-mesa-dev xorg-dev
  ```
- **Windows:** Ensure you're not running in headless mode
- **macOS:** Grant accessibility permissions if prompted

### Build errors
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o zhyper-checker main.go
```

### Dependencies issues
```bash
# Update dependencies
go get -u fyne.io/fyne/v2@v2.4.5
go mod tidy
```

## Performance Tips

1. **Optimal Thread Count:**
   - Fast internet: 50-100 threads
   - Medium internet: 20-50 threads
   - Slow internet: 10-20 threads

2. **Debug Mode:**
   - Disable for production (faster)
   - Enable only when troubleshooting

3. **DEEP Mode:**
   - Enable for thorough checks
   - Disable for quick scans

## Security Notice

⚠️ **This tool is for educational and authorized testing purposes only.**
- Only test systems you own or have explicit permission to test
- Unauthorized access to computer systems is illegal
- The developers assume no liability for misuse

## Support & Contact

- **Telegram:** https://t.me/zhyperflow
- **GitHub Issues:** https://github.com/wppluginsn/golang/issues

## License

This project is provided as-is for educational purposes.

## Credits

Developed by Zhyper Team
Version 4.8.7 - Golang Edition

---

**Star ⭐ this repository if you find it useful!**