package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8081/health", "URL to test")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	duration := flag.Duration("d", 10*time.Second, "Duration of the test")
	flag.Parse()

	fmt.Printf("Starting load test: %s with %d workers for %s\n", *url, *concurrency, *duration)

	var wg sync.WaitGroup
	var requests int64
	var errors int64

	stop := time.After(*duration)
	start := time.Now()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					resp, err := client.Get(*url)
					if err != nil {
						errors++
					} else {
						requests++
						resp.Body.Close()
					}
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\n--- Results ---\n")
	fmt.Printf("Duration: %s\n", elapsed)
	fmt.Printf("Total Requests: %d\n", requests)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Requests/sec: %.2f\n", float64(requests)/elapsed.Seconds())
}
