// ЗАДАЧА 2: Параллельное скачивание с обработкой ошибок
package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const timeoutLimit = 90

type Result struct {
	msg string
	err error
}

func fakeDownload(url string) Result {
	r := rand.Intn(100)
	time.Sleep(time.Duration(r) * time.Millisecond)
	if r > timeoutLimit {
		return Result{
			err: fmt.Errorf("failed to download data from %s: timeout", url),
		}
	}
	return Result{
		msg: fmt.Sprintf("downloaded data from %s\n", url),
	}
}

// download конкурентно скачивает данные и возвращает все ошибки через errors.Join.
func download(urls []string) ([]string, error) {
	results := make(chan Result, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			results <- fakeDownload(u)
		}(url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var msgs []string
	var errs []error
	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			msgs = append(msgs, r.msg)
		}
	}
	return msgs, errors.Join(errs...)
}

func main() {
	msgs, err := download([]string{
		"https://example.com/e25e26d3-6aa3-4d79-9ab4-fc9b71103a8c.xml",
		"https://example.com/a601590e-31c1-424a-8ccc-decf5b35c0f6.xml",
		"https://example.com/1cf0dd69-a3e5-4682-84e3-dfe22ca771f4.xml",
		"https://example.com/ceb566f2-a234-4cb8-9466-4a26f1363aa8.xml",
		"https://example.com/b6ed16d7-cb3d-4cba-b81a-01a789d3a914.xml",
	})
	if err != nil {
		fmt.Println("Errors:", err)
	}
	fmt.Println(msgs)
}
