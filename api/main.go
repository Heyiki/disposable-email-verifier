package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	disposableDomains sync.Map
	requestCount      int
	resetTime         time.Time
)

const (
	domainsURL         = "https://disposable.github.io/disposable-email-domains/domains_mx.json"
	defaultRequestLimit = 20
	resetInterval      = 24 * time.Hour // The counter is reset every 24 hours
	requestLimitKey    = "REQUEST_LIMIT"
)

func loadDisposableDomains() error {
	resp, err := http.Get(domainsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var domains []string
	if err := json.Unmarshal(body, &domains); err != nil {
		return err
	}

	for _, domain := range domains {
		disposableDomains.Store(domain, struct{}{})
	}
	return nil
}

func isDisposableEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	_, found := disposableDomains.Load(parts[1])
	return found
}

func getRequestLimit() int {
	limitStr := os.Getenv(requestLimitKey)
	if limitStr == "" {
		return defaultRequestLimit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("Invalid request limit value: %v", err)
		return defaultRequestLimit
	}
	return limit
}

func checkRequestLimit() bool {
	now := time.Now()
	if now.After(resetTime) {
		// 重置计数器
		requestCount = 0
		resetTime = now.Add(resetInterval)
	}

	if requestCount >= getRequestLimit() {
		return false
	}

	requestCount++
	return true
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if !checkRequestLimit() {
		// http.Error(w, "Daily request limit exceeded", http.StatusTooManyRequests)
		response := map[string]interface{}{
			"code": http.StatusTooManyRequests,
			"msg": "Daily request limit exceeded",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		// http.Error(w, "Email is required", http.StatusBadRequest)
		response := map[string]interface{}{
			"code": http.StatusBadRequest,
			"msg": "Email is required",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	isDisposable := isDisposableEmail(email)
	response := map[string]interface{}{
		"code": http.StatusOK,
		"msg": "success",
		"data": map[string]interface{}{
			"disposable": isDisposable,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Vercel expects a function named "Handler" to be exported
func Handler(w http.ResponseWriter, r *http.Request) {
	verifyHandler(w, r)
}

func init() {
	if err := loadDisposableDomains(); err != nil {
		log.Fatalf("Failed to load disposable email domains: %v", err)
	}

	// Refresh the list of domain names on a regular basis
	go func() {
		ticker := time.NewTicker(resetInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := loadDisposableDomains(); err != nil {
				log.Printf("Failed to refresh disposable email domains: %v", err)
			}
		}
	}()
}
