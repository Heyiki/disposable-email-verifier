package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	disposableDomains map[string]struct{}
	mu                sync.RWMutex
)

const domainsURL = "https://disposable.github.io/disposable-email-domains/domains_mx.json"

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

	mu.Lock()
	defer mu.Unlock()
	disposableDomains = make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		disposableDomains[domain] = struct{}{}
	}
	return nil
}

func isDisposableEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]

	mu.RLock()
	defer mu.RUnlock()
	_, found := disposableDomains[domain]
	return found
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	isDisposable := isDisposableEmail(email)
	response := map[string]bool{"disposable": isDisposable}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	if err := loadDisposableDomains(); err != nil {
		log.Fatalf("Failed to load disposable email domains: %v", err)
	}

	// 定时刷新域名列表，例如每小时刷新一次
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := loadDisposableDomains(); err != nil {
				log.Printf("Failed to refresh disposable email domains: %v", err)
			}
		}
	}()

	http.HandleFunc("/verify", verifyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
