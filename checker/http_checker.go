package checker

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Name     string
	URL      string
	Success  bool
	Status   int
	Duration time.Duration
	Err      error
}

func CheckHTTP(name, url string) Result {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		duration := time.Since(start)
		return Result{
			Name:     name,
			URL:      url,
			Success:  false,
			Duration: duration,
			Err:      err,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		return Result{
			Name:     name,
			URL:      url,
			Success:  false,
			Duration: duration,
			Err:      err,
		}
	}

	computing_duration := time.Since(start)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return Result{
			Name:     name,
			URL:      url,
			Success:  true,
			Status:   resp.StatusCode,
			Duration: computing_duration,
			Err:      nil,
		}
	} else {
		return Result{
			Name:     name,
			URL:      url,
			Success:  false,
			Status:   resp.StatusCode,
			Duration: computing_duration,
			Err:      nil,
		}
	}
}

type Target struct {
	Name string
	URL  string
}

func RunAll(targets []Target, workers int) []Result {

	jobs := make(chan Target, len(targets))
	results := make(chan Result, len(targets))

	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				results <- CheckHTTP(t.Name, t.URL)
			}
		}()
	}

	for _, target := range targets {
		jobs <- target
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	allResults := make([]Result, 0, len(targets))
	for result := range results {
		allResults = append(allResults, result)
	}
	return allResults
}
