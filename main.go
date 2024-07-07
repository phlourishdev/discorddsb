package main

import (
	"discorddsb/dbmanager"
	"discorddsb/htmlparser"
	"fmt"
	"github.com/phlourishdev/DSBgo"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func retrieveToken() string {
	username := os.Getenv("DSB_USER")
	if username == "" {
		log.Fatalf("DSB_USER environment variable is not set")
	}

	password := os.Getenv("DSB_PASSWORD")
	if password == "" {
		log.Fatalf("DSB_PASSWORD environment variable is not set")
	}

	token, err := DSBgo.Authenticate(username, password)
	if err != nil {
		log.Fatalf("DSBgo.Authenticate failed: %v", err)
	}
	log.Printf("Retrieved token: %s", token)

	return token
}

func retrievePlans(token string) ([]DSBgo.ProcessedPlan, error) {
	plans, err := DSBgo.GetPlans(token)
	if err != nil {
		return nil, fmt.Errorf("DSBgo.GetPlans failed: %v", err)
	}

	return plans, nil
}

func main() {
	_, err := dbmanager.InitDB()
	if err != nil {
		log.Fatalf("DB initialization sequence failed: %v", err)
	}

	// Testing setup
	token := retrieveToken()
	plans, err := retrievePlans(token)
	ticker := time.NewTicker(15 * time.Second) // TODO: subject to change
	done := make(chan bool)
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-ticker.C:
				plans, err = retrievePlans(token)
				for i := range plans {
					wg.Add(1)
					go func(url string) {
						defer wg.Done()
						htmlparser.ParseHTML(url) // TODO: get vars here
					}(plans[i].URL)
				}
				wg.Wait()

			case <-signals:
				fmt.Println("Got shutdown signal")
				ticker.Stop()
				done <- true
				return

			case <-done:
				return
			}
		}
	}()

	<-done
	fmt.Println("Shutting down gracefully")
}
