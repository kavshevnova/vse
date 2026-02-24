// ЗАДАЧА 7: GetFirstResult и GetResults
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type result struct {
	msg string
	err error
}

type searh func() *result
type replicas []searh

func fakeSearch(kind string) searh {
	return func() *result {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		return &result{
			msg: fmt.Sprintf("%q result", kind),
		}
	}
}

// getFirstResult запускает все реплики параллельно и возвращает первый результат.
func getFirstResult(ctx context.Context, replicas replicas) *result {
	ch := make(chan *result, len(replicas))
	for _, r := range replicas {
		go func(s searh) {
			ch <- s()
		}(r)
	}
	select {
	case <-ctx.Done():
		return &result{err: ctx.Err()}
	case r := <-ch:
		return r
	}
}

// getResults запускает поиск для каждого набора реплик параллельно.
func getResults(ctx context.Context, replicaKinds []replicas) []*result {
	results := make([]*result, len(replicaKinds))
	var wg = make(chan struct{}, len(replicaKinds))
	for i, rk := range replicaKinds {
		go func(idx int, rs replicas) {
			results[idx] = getFirstResult(ctx, rs)
			wg <- struct{}{}
		}(i, rk)
	}
	for range replicaKinds {
		<-wg
	}
	return results
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	replicaKinds := []replicas{
		{fakeSearch("web1"), fakeSearch("web2")},
		{fakeSearch("image1"), fakeSearch("image2")},
		{fakeSearch("video1"), fakeSearch("video2")},
	}

	for _, res := range getResults(ctx, replicaKinds) {
		fmt.Println(res.msg, res.err)
	}
}
