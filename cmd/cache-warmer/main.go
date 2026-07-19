package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/logger"
)

func main() {
	concurrency := flag.Int("concurrency", 5, "Max concurrent requests to Cloudflare")
	dryRun := flag.Bool("dry-run", false, "List URLs that would be warmed, without making HTTP requests")
	flag.Parse()

	// Bootstrap configuration and database
	config.Load()
	logger.InitLogger(config.AppConfig.Env)
	database.Connect()

	baseURL := config.AppConfig.CacheWarmerBaseURL
	if baseURL == "" {
		fmt.Println("Error: CACHE_WARMER_BASE_URL environment variable is required")
		os.Exit(1)
	}

	svc := service.NewCacheWarmerService(database.DB, baseURL, logger.Log)

	if *dryRun {
		urls, err := svc.CollectActiveURLs()
		if err != nil {
			log.Fatalf("Failed to collect active URLs: %v", err)
		}

		if len(urls) == 0 {
			fmt.Println("No active URLs found to warm.")
			return
		}

		fmt.Println("--- DRY RUN: URLs to be warmed ---")
		for _, u := range urls {
			fmt.Println(baseURL + u)
		}
		fmt.Printf("\nTotal: %d URLs\n", len(urls))
		return
	}

	fmt.Printf("Warming active invitations... (Concurrency: %d)\n", *concurrency)
	total, ok, fail, err := svc.WarmActiveInvitations(*concurrency)
	if err != nil {
		log.Fatalf("Cache warmer failed: %v", err)
	}

	fmt.Printf("\nDone!\nTotal: %d\nSuccess: %d\nFailed: %d\n", total, ok, fail)
	if fail > 0 {
		os.Exit(1) // Exit non-zero if there were any failures, useful for scripting
	}
}
