package checker

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
		return Result{
			Name:    name,
			URL:     url,
			Success: false,
			Err:     err,
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
			Err:      fmt.Errorf("status code: %d", resp.StatusCode),
		}
	}
}
