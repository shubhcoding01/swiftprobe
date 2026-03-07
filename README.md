# 🔍 SwiftProbe

**SwiftProbe** is a lightning-fast, concurrent directory and file fuzzer written in Go. Designed for ethical hackers, penetration testers, and bug bounty hunters to rapidly discover hidden paths, files, and endpoints on web servers.

By leveraging Go's powerful Goroutines, SwiftProbe can send hundreds of HTTP requests per second — significantly faster than traditional Python-based fuzzers like dirb or dirbuster.

---

## ✨ Features

- ⚡ **Insanely Fast** — Built with Go concurrency (Goroutines + semaphore pool) for maximum throughput
- 🪶 **Lightweight & Portable** — Single standalone binary. Works on Kali Linux, Windows, and macOS
- 📂 **Extension Fuzzing** — Probe each word with multiple extensions (`-x php,bak,html,json`)
- 🎯 **Smart Filtering** — Match or exclude specific HTTP status codes (`-mc 200,301` / `-fc 404`)
- 🏷️ **Auto-Tagging** — Findings are automatically tagged (`admin-panel`, `git-exposure`, `backup-file`)
- 🎨 **Color-Coded Output** — Instantly distinguish 200s, 301s, 403s, and 500s at a glance
- 💾 **Memory Efficient** — Streams wordlists line-by-line, handles multi-GB lists with ~5MB RAM
- 🔐 **Auth Support** — Send custom headers like `Authorization: Bearer <token>`
- 📄 **Save Results** — Export findings to a file with `-o results.txt`

---

## 🛠️ Installation

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)
- Git

### Clone and Build
```bash
git clone https://github.com/shubhcoding01/swiftprobe.git
cd swiftprobe
go mod tidy
go build -o swiftprobe ./cmd/swiftprobe
```

### Windows
```powershell
git clone https://github.com/shubhcoding01/swiftprobe.git
cd swiftprobe
go mod tidy
go build -o swiftprobe.exe ./cmd/swiftprobe
```

### Verify Installation
```bash
./swiftprobe --help
```

---

## 🚀 Usage
```
./swiftprobe -u <URL> -w <WORDLIST> [OPTIONS]
```

### Flags

| Flag       | Default                        | Description                          |
|------------|--------------------------------|--------------------------------------|
| `-u`       | *(required)*                   | Target URL (e.g. `http://example.com`) |
| `-w`       | *(required)*                   | Path to wordlist file                |
| `-t`       | `50`                           | Number of concurrent threads         |
| `-timeout` | `5`                            | HTTP request timeout in seconds      |
| `-mc`      | `200,201,301,302,401,403,500`  | Match these HTTP status codes        |
| `-fc`      | `404`                          | Exclude these HTTP status codes      |
| `-x`       | *(none)*                       | Extensions to append (e.g. `php,bak,html`) |
| `-ua`      | `Mozilla/5.0 (SwiftProbe/1.0)` | Custom User-Agent string             |
| `-follow`  | `false`                        | Follow HTTP redirects                |
| `-v`       | `false`                        | Verbose mode (show errors + progress)|
| `-o`       | *(none)*                       | Save results to output file          |

---

## 📖 Examples

**Basic scan**
```bash
./swiftprobe -u http://testphp.vulnweb.com -w wordlists/common.txt
```

**Extension fuzzing**
```bash
./swiftprobe -u http://example.com -w wordlists/common.txt -x php,bak,html,json
```

**Only show 200 OK results**
```bash
./swiftprobe -u http://example.com -w wordlists/common.txt -mc 200
```

**Authenticated scan**
```bash
./swiftprobe -u http://api.example.com \
             -w wordlists/common.txt \
             -H "Authorization: Bearer eyJhbGci..." \
             -mc 200,201,403
```

**Aggressive scan with 100 threads**
```bash
./swiftprobe -u http://example.com \
             -w wordlists/common.txt \
             -t 100 \
             -timeout 3 \
             -x php,html \
             -v
```

**Save results to file**
```bash
./swiftprobe -u http://example.com -w wordlists/common.txt -o results.txt
```

---

## 🎨 Output Example
```
 ____          _  __ _   ____            _
/ ___|_      _(_)/ _| |_|  _ \ _ __ ___ | |__   ___
\___ \ \ /\ / / | |_| __| |_) | '__/ _ \| '_ \ / _ \
 ___) \ V  V /| |  _| |_|  __/| | | (_) | |_) |  __/
|____/ \_/\_/ |_|_|  \__|_|   |_|  \___/|_.__/ \___|

  High-Speed Directory Fuzzer

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Target   :  http://testphp.vulnweb.com
  Wordlist :  wordlists/common.txt
  Threads  :  50
  Timeout  :  5s
  Codes    :  200,301,302,403
  Started  :  2026-03-06 20:54:28
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [200]  /login.php                  [size: 5.2KB]   [accessible, login-page]
  [301]  /admin                      [size: 169B]    [redirect → /admin/]
  [301]  /images                     [size: 169B]    [redirect]
  [200]  /.git                       [size: 231B]    [accessible, git-exposure]
  [403]  /backup                     [size: 0B]      [forbidden, backup-file]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Finished :  2026-03-06 20:54:33
  Duration :  4.865s
  Total    :  214
  Found    :  5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Color Reference

| Color    | Status Codes     | Meaning                        |
|----------|------------------|--------------------------------|
| 🟢 Green  | 200, 201, 204    | Accessible — investigate first |
| 🟡 Yellow | 301, 302, 307    | Redirect — note destination    |
| 🟣 Magenta| 401              | Auth required — try credentials|
| 🔴 Red    | 403, 500         | Forbidden / Server error       |

---

## 📁 Project Structure
```
swiftprobe/
├── cmd/swiftprobe/main.go        # Entry point, CLI flags
├── internal/
│   ├── fuzzer/
│   │   ├── fuzzer.go             # Core engine, goroutine pool
│   │   └── result.go             # Result struct, severity, tags
│   ├── requester/http.go         # HTTP client, Probe(), BuildURL()
│   ├── wordlist/reader.go        # Buffered wordlist streaming
│   └── output/printer.go         # Thread-safe terminal output
├── wordlists/common.txt          # Built-in starter wordlist
├── docs/SwiftProbe.txt           # Full documentation
├── go.mod
└── README.md
```

---

## ⚠️ Legal Disclaimer

SwiftProbe is intended **only** for use on systems you own or have **explicit written permission** to test. Unauthorized scanning is illegal under the Computer Fraud and Abuse Act (CFAA), the Computer Misuse Act (CMA), and equivalent laws worldwide.

**Safe practice targets (no permission needed):**
- `http://testphp.vulnweb.com` — Acunetix intentionally vulnerable site
- `http://demo.testfire.net` — IBM intentionally vulnerable site
- `https://juice-shop.herokuapp.com` — OWASP Juice Shop
- Your own local Docker containers

The authors accept no liability for misuse of this tool.

---

## 📜 License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">Built with ❤️ in Go by <a href="https://github.com/shubhcoding01">shubhcoding01</a></p>