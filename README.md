# 🚀 Swiftprobe

**Swiftprobe** is a lightning-fast, concurrent directory and file fuzzer written in Go. It is designed for ethical hackers, penetration testers, and bug bounty hunters to rapidly discover hidden paths on web servers.

By leveraging Go's powerful Goroutines, Swiftprobe can send hundreds of HTTP requests per second, making it significantly faster than traditional Python-based fuzzers.

## ✨ Features
* **Insanely Fast:** Built with Go concurrency (Goroutines) for maximum performance.
* **Lightweight & Portable:** Compiles down to a single, standalone executable binary. Drop it into Kali Linux, Windows, or macOS and run it instantly.
* **Custom Wordlists:** Supports massive wordlists without consuming excessive RAM.
* **Clean Output:** Colorized terminal output to easily distinguish between found directories (200 OK) and redirects (301/302).

## 🛠️ Installation

Ensure you have [Go](https://go.dev/) installed on your system.

**1. Clone the repository:**
```bash
git clone [https://github.com/shubhcoding01/swiftprobe.git](https://github.com/shubhcoding01/swiftprobe.git)
cd swiftprobe