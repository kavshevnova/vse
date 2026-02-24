package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ТЗ: Сервис должен периодически проверять доступность внешних API
// и хранить результаты последних проверок в памяти.
// Найти и исправить проблемы

type HealthCheck struct {
	URL       string
	Status    string
	CheckedAt time.Time
	Response  *http.Response
	BodyBytes []byte
}

var healthChecks = []*HealthCheck{}

func main() {
	urls := []string{
		"https://api.example.com/health",
		"https://api2.example.com/status",
		"https://api3.example.com/ping",
	}

	// Проверяем каждую минуту
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		for _, url := range urls {
			go checkHealth(url)
		}
	}
}

func checkHealth(url string) {
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		healthChecks = append(healthChecks, &HealthCheck{
			URL:       url,
			Status:    "ERROR",
			CheckedAt: time.Now(),
		})
		return
	}

	body, _ := io.ReadAll(resp.Body)

	healthChecks = append(healthChecks, &HealthCheck{
		URL:       url,
		Status:    "OK",
		CheckedAt: time.Now(),
		Response:  resp,
		BodyBytes: body,
	})

	fmt.Printf("Checked %s: %s\n", url, resp.Status)
}

func GetRecentChecks() []*HealthCheck {
	return healthChecks
}
