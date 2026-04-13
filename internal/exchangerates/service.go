package exchangerates

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"
)

//go:embed currency_fallback.json
var embeddedFallback []byte

type CurrencyInfo struct {
	Rate float64 `json:"rate"`
}

type Response struct {
	Base       string                  `json:"base"`
	Timestamp  int64                   `json:"timestamp"`
	Stale      bool                    `json:"stale"`
	Currencies map[string]CurrencyInfo `json:"currencies"`
}

type rawRatesPayload struct {
	Base      string             `json:"base"`
	Timestamp int64              `json:"timestamp"`
	Rates     map[string]float64 `json:"rates"`
}

type serviceCache struct {
	response        Response
	lastAttemptAt   time.Time
	lastSuccessAt   time.Time
	lastFailureAt   time.Time
	lastFailureText string
	source          string
}

type Service struct {
	logger          *slog.Logger
	client          *http.Client
	refreshInterval time.Duration
	now             func() time.Time
	appID           string

	mu    sync.RWMutex
	cache *serviceCache
}

var supportedCurrencies = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"GBP": {},
	"RUB": {},
	"CAD": {},
	"AUD": {},
	"JPY": {},
	"CNY": {},
	"INR": {},
	"BRL": {},
}

var supportedCurrencyCodes = func() []string {
	codes := make([]string, 0, len(supportedCurrencies))
	for code := range supportedCurrencies {
		codes = append(codes, code)
	}
	slices.Sort(codes)
	return codes
}()

func NewService(logger *slog.Logger, appID string) *Service {
	s := &Service{
		logger: logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		refreshInterval: 24 * time.Hour,
		now:             time.Now,
		appID:           appID,
	}

	if err := s.loadInitialFallback(); err != nil {
		s.logger.Error("exchange rate fallback bootstrap failed", "error", err)
	}

	return s
}

func (s *Service) Get(ctx context.Context) (Response, error) {
	if err := s.ensureFresh(ctx); err != nil {
		s.logger.Error("exchange rate refresh failed", "error", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cache == nil {
		return Response{}, errors.New("exchange rates unavailable")
	}

	return s.cache.response, nil
}

func (s *Service) ensureFresh(ctx context.Context) error {
	s.mu.RLock()
	cache := s.cache
	shouldRefresh := cache == nil || cache.lastAttemptAt.IsZero() || s.now().Sub(cache.lastAttemptAt) >= s.refreshInterval
	s.mu.RUnlock()

	if !shouldRefresh {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cache = s.cache
	if cache != nil && !cache.lastAttemptAt.IsZero() && s.now().Sub(cache.lastAttemptAt) < s.refreshInterval {
		return nil
	}

	now := s.now()
	if cache == nil {
		cache = &serviceCache{}
	}
	cache.lastAttemptAt = now

	resp, err := s.fetchFromUpstream(ctx)
	if err != nil {
		cache.lastFailureAt = now
		cache.lastFailureText = err.Error()
		if s.cache == nil {
			if fallbackResp, fallbackErr := s.loadFallbackResponse(); fallbackErr == nil {
				cache.response = fallbackResp
				cache.source = "fallback"
			} else {
				return errors.Join(err, fallbackErr)
			}
		}
		cache.response.Stale = true
		s.cache = cache
		return err
	}

	cache.response = resp
	cache.response.Stale = false
	cache.lastSuccessAt = now
	cache.lastFailureAt = time.Time{}
	cache.lastFailureText = ""
	cache.source = "upstream"
	s.cache = cache
	return nil
}

func (s *Service) loadInitialFallback() error {
	resp, err := s.loadFallbackResponse()
	if err != nil {
		return err
	}

	now := s.now()
	s.mu.Lock()
	s.cache = &serviceCache{
		response: Response{
			Base:       resp.Base,
			Timestamp:  resp.Timestamp,
			Stale:      true,
			Currencies: resp.Currencies,
		},
		source:        "fallback",
		lastAttemptAt: time.Time{},
		lastSuccessAt: time.Time{},
		lastFailureAt: now,
	}
	s.mu.Unlock()

	return nil
}

func (s *Service) loadFallbackResponse() (Response, error) {
	paths := []string{
		"/mnt/data/currency.json",
		"internal/exchangerates/currency_fallback.json",
	}

	for _, path := range paths {
		payload, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		resp, err := normalizePayload(payload)
		if err == nil {
			resp.Stale = true
			return resp, nil
		}
	}

	resp, err := normalizePayload(embeddedFallback)
	if err != nil {
		return Response{}, err
	}
	resp.Stale = true
	return resp, nil
}

func (s *Service) fetchFromUpstream(ctx context.Context) (Response, error) {
	if s.appID == "" {
		return Response{}, errors.New("exchange rates app id is not configured")
	}

	upstreamURL := url.URL{
		Scheme: "https",
		Host:   "openexchangerates.org",
		Path:   "/api/latest.json",
	}
	query := upstreamURL.Query()
	query.Set("app_id", s.appID)
	upstreamURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL.String(), nil)
	if err != nil {
		return Response{}, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("failed to close exchange rates response body", "error", closeErr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("upstream returned status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}

	return normalizePayload(body)
}

func normalizePayload(payload []byte) (Response, error) {
	var raw rawRatesPayload
	if err := json.Unmarshal(payload, &raw); err != nil {
		return Response{}, err
	}

	if raw.Base == "" {
		raw.Base = "USD"
	}
	if raw.Base != "USD" {
		return Response{}, fmt.Errorf("unsupported base currency %q", raw.Base)
	}
	if raw.Rates == nil {
		return Response{}, errors.New("missing rates payload")
	}

	currencies := make(map[string]CurrencyInfo, len(supportedCurrencies))
	for _, code := range supportedCurrencyCodes {
		rate := 1.0
		if code != "USD" {
			value, ok := raw.Rates[code]
			if !ok {
				return Response{}, fmt.Errorf("missing rate for %s", code)
			}
			rate = value
		}

		currencies[code] = CurrencyInfo{
			Rate: rate,
		}
	}

	return Response{
		Base:       "USD",
		Timestamp:  raw.Timestamp,
		Stale:      false,
		Currencies: currencies,
	}, nil
}
