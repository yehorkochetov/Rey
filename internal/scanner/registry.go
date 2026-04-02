package scanner

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type Registry struct {
	scanners []Scanner
}

func (r *Registry) Register(s Scanner) {
	r.scanners = append(r.scanners, s)
}

func (r *Registry) RunAll(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []DeadResource
		errs    []error
	)

	for _, s := range r.scanners {
		wg.Add(1)
		go func(s Scanner) {
			defer wg.Done()
			found, err := s.Scan(ctx, cfg)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results = append(results, found...)
		}(s)
	}

	wg.Wait()

	if len(errs) > 0 {
		return results, errs[0]
	}

	return results, nil
}
