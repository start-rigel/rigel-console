package console

import (
	"context"
	"sync"
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type usageState struct {
	WindowStarted time.Time `json:"window_started"`
	Used          int       `json:"used"`
	CooldownUntil time.Time `json:"cooldown_until"`
}

type storedRecommendation struct {
	Response  model.CatalogRecommendationResponse `json:"response"`
	ExpiresAt time.Time                           `json:"expires_at"`
}

type securityStore interface {
	LoadUsage(ctx context.Context, scope, key string) (usageState, bool, error)
	SaveUsage(ctx context.Context, scope, key string, usage usageState, ttl time.Duration) error
	LoadRecommendation(ctx context.Context, key string) (storedRecommendation, bool, error)
	SaveRecommendation(ctx context.Context, key string, value storedRecommendation, ttl time.Duration) error
	HasChallengePass(ctx context.Context, key string) (bool, error)
	SetChallengePass(ctx context.Context, key string, ttl time.Duration) error
}

type memorySecurityStore struct {
	mu              sync.Mutex
	usages          map[string]memoryValue[usageState]
	recommendations map[string]memoryValue[storedRecommendation]
	challengePasses map[string]time.Time
}

type memoryValue[T any] struct {
	Value     T
	ExpiresAt time.Time
}

func newMemorySecurityStore() securityStore {
	return &memorySecurityStore{
		usages:          make(map[string]memoryValue[usageState]),
		recommendations: make(map[string]memoryValue[storedRecommendation]),
		challengePasses: make(map[string]time.Time),
	}
}

func (s *memorySecurityStore) LoadUsage(_ context.Context, scope, key string) (usageState, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.usages[scope+":"+key]
	if !ok || expired(item.ExpiresAt) {
		delete(s.usages, scope+":"+key)
		return usageState{}, false, nil
	}
	return item.Value, true, nil
}

func (s *memorySecurityStore) SaveUsage(_ context.Context, scope, key string, usage usageState, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.usages[scope+":"+key] = memoryValue[usageState]{
		Value:     usage,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (s *memorySecurityStore) LoadRecommendation(_ context.Context, key string) (storedRecommendation, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.recommendations[key]
	if !ok || expired(item.ExpiresAt) {
		delete(s.recommendations, key)
		return storedRecommendation{}, false, nil
	}
	return item.Value, true, nil
}

func (s *memorySecurityStore) SaveRecommendation(_ context.Context, key string, value storedRecommendation, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recommendations[key] = memoryValue[storedRecommendation]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (s *memorySecurityStore) HasChallengePass(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	until, ok := s.challengePasses[key]
	if !ok || expired(until) {
		delete(s.challengePasses, key)
		return false, nil
	}
	return true, nil
}

func (s *memorySecurityStore) SetChallengePass(_ context.Context, key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.challengePasses[key] = time.Now().Add(ttl)
	return nil
}

func expired(deadline time.Time) bool {
	return !deadline.IsZero() && time.Now().After(deadline)
}
