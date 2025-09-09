package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Domain represents a domain entry from the CSV
type Domain struct {
	Name     string
	Category string
}

// Result represents the result of checking a domain
type Result struct {
	Domain     Domain
	StatusCode int
	Error      error
	Duration   time.Duration
}

// Worker function that processes domains from the jobs channel
func worker(id int, jobs <-chan Domain, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	
	
	
	for domain := range jobs {
		start := time.Now()
		result := Result{Domain: domain}
		
		// Add https:// prefix to the domain
		url := fmt.Sprintf("https://%s", domain.Name)
		
		// Create request with context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			result.Error = err
			result.Duration = time.Since(start)
			results <- result
			cancel()
			continue
		}
		
		// Set User-Agent to avoid being blocked
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		
		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			result.Error = err
		} else {
			result.StatusCode = resp.StatusCode
			resp.Body.Close()
		}
		
		result.Duration = time.Since(start)
		results <- result
		cancel()
		
		// Small delay to be respectful to servers
		time.Sleep(100 * time.Millisecond)
	}
}

// ReadDomainsFromCSV reads domains from a CSV file
func ReadDomainsFromCSV(filename string) ([]Domain, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Skip header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}
	
	var domains []Domain
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record: %v", err)
		}
		
		if len(record) >= 2 {
			domains = append(domains, Domain{
				Name:     record[0],
				Category: record[1],
			})
		}
	}
	
	return domains, nil
}

func main() {
	// Configuration
	const numWorkers = 10
	csvFile := "domains2.csv" // Change this to your CSV file path
	
	// Read domains from CSV
	domains, err := ReadDomainsFromCSV(csvFile)
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	domains = domains[:200]
	
	fmt.Printf("Loaded %d domains from CSV\n", len(domains))
	fmt.Printf("Starting %d workers...\n\n", numWorkers)
	
	// Create channels
	jobs := make(chan Domain, len(domains))
	results := make(chan Result, len(domains))
	
	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}
	
	// Send jobs to workers
	for _, domain := range domains {
		jobs <- domain
	}
	close(jobs)
	
	// Start a goroutine to collect results
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Statistics
	var successful, failed int
	statusCodes := make(map[int]int)
	categoryStats := make(map[string]struct {
		Total   int
		Success int
		Failed  int
	})
	
	// Process results
	fmt.Println("Processing domains...")
	fmt.Println("=====================================")
	
	for result := range results {
		category := result.Domain.Category
		stats := categoryStats[category]
		stats.Total++
		
		if result.Error != nil {
			failed++
			stats.Failed++
			fmt.Printf("❌ %s (%s) - Error: %v [%v]\n", 
				result.Domain.Name, 
				result.Domain.Category,
				result.Error, 
				result.Duration)
		} else {
			successful++
			stats.Success++
			statusCodes[result.StatusCode]++
			
			statusIcon := "✅"
			if result.StatusCode >= 400 {
				statusIcon = "⚠️"
			}
			
			fmt.Printf("%s %s (%s) - Status: %d [%v]\n", 
				statusIcon,
				result.Domain.Name, 
				result.Domain.Category,
				result.StatusCode, 
				result.Duration)
		}
		
		categoryStats[category] = stats
	}
	
	// Print summary
	fmt.Println("\n======================================")
	fmt.Println("SUMMARY")
	fmt.Println("=====================================")
	fmt.Printf("Total domains checked: %d\n", len(domains))
	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n\n", failed)
	
	// Status code distribution
	fmt.Println("Status Code Distribution:")
	for code, count := range statusCodes {
		fmt.Printf("  %d: %d domains\n", code, count)
	}
	
	// Category statistics
	fmt.Println("\nCategory Statistics:")
	for category, stats := range categoryStats {
		fmt.Printf("  %s: Total=%d, Success=%d, Failed=%d\n", 
			category, stats.Total, stats.Success, stats.Failed)
	}
}
