package wordlist

import (
    "bufio"
    "os"
)

// Stream words one-by-one into a channel (memory efficient for huge lists)
func Stream(path string) (<-chan string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    ch := make(chan string, 100) // buffered channel

    go func() {
        defer f.Close()
        defer close(ch)

        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            line := scanner.Text()
            if line != "" {
                ch <- line
            }
        }
    }()

    return ch, nil
}
