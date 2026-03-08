// package wordlist

// import (
//     "bufio"
//     "os"
// )

// // Stream words one-by-one into a channel (memory efficient for huge lists)
// func Stream(path string) (<-chan string, error) {
//     f, err := os.Open(path)
//     if err != nil {
//         return nil, err
//     }

//     ch := make(chan string, 100) // buffered channel

//     go func() {
//         defer f.Close()
//         defer close(ch)

//         scanner := bufio.NewScanner(f)
//         for scanner.Scan() {
//             line := scanner.Text()
//             if line != "" {
//                 ch <- line
//             }
//         }
//     }()

//     return ch, nil
// }

package wordlist

import (
	"bufio"
	"os"
	"strings"
)

// Stream opens a wordlist file and pushes each valid word into a channel.
// Skips blank lines, comment lines (#), and strips inline comments.
// Memory efficient — never loads the full file into RAM.
func Stream(path string) (<-chan string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 100) // buffered to keep reader ahead of workers

	go func() {
		defer f.Close()
		defer close(ch)

		scanner := bufio.NewScanner(f)

		// Increase buffer size for very long lines
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Skip blank lines
			if line == "" {
				continue
			}

			// Skip full-line comments
			if strings.HasPrefix(line, "#") {
				continue
			}

			// Strip inline comments e.g. "administrator (Joomla)" → "administrator"
			if idx := strings.Index(line, " "); idx != -1 {
				line = line[:idx]
			}

			ch <- line
		}
	}()

	return ch, nil
}
