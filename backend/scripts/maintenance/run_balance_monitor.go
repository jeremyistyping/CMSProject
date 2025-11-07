package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	// Command line flags
	intervalPtr := flag.Int("interval", 30, "Check interval in minutes")
	autoFixPtr := flag.Bool("autofix", false, "Automatically fix balance discrepancies")
	oncePtr := flag.Bool("once", false, "Run check once and exit (no periodic monitoring)")
	healthPtr := flag.Bool("health", false, "Show balance health metrics and exit")
	helpPtr := flag.Bool("help", false, "Show help information")

	flag.Parse()

	if *helpPtr {
		showHelp()
		return
	}

	// Connect to database
	db := database.ConnectDB()
	
	// Create monitoring service
	monitoringService := services.NewBalanceMonitoringService(db)

	if *healthPtr {
		showHealthMetrics(monitoringService)
		return
	}

	if *oncePtr {
		runSingleCheck(monitoringService, *autoFixPtr)
		return
	}

	// Run periodic monitoring
	runPeriodicMonitoring(monitoringService, *intervalPtr, *autoFixPtr)
}

func showHelp() {
	fmt.Println("Balance Monitoring Utility")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("This utility monitors balance synchronization between cash/bank accounts and GL accounts.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/run_balance_monitor.go [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -interval int     Check interval in minutes (default: 30)")
	fmt.Println("  -autofix         Automatically fix balance discrepancies (default: false)")
	fmt.Println("  -once            Run check once and exit (default: false)")
	fmt.Println("  -health          Show balance health metrics and exit (default: false)")
	fmt.Println("  -help            Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run periodic monitoring every 15 minutes without auto-fix")
	fmt.Println("  go run cmd/run_balance_monitor.go -interval 15")
	fmt.Println()
	fmt.Println("  # Run periodic monitoring with auto-fix enabled")
	fmt.Println("  go run cmd/run_balance_monitor.go -interval 30 -autofix")
	fmt.Println()
	fmt.Println("  # Run single check with auto-fix")
	fmt.Println("  go run cmd/run_balance_monitor.go -once -autofix")
	fmt.Println()
	fmt.Println("  # Show current balance health metrics")
	fmt.Println("  go run cmd/run_balance_monitor.go -health")
}

func showHealthMetrics(service *services.BalanceMonitoringService) {
	fmt.Println("📊 BALANCE HEALTH METRICS")
	fmt.Println("========================")

	health, err := service.GetBalanceHealth()
	if err != nil {
		log.Fatalf("❌ Failed to get balance health: %v", err)
	}

	fmt.Printf("Status: %s\n", health["status"])
	fmt.Printf("Total Accounts: %v\n", health["total_accounts"])
	fmt.Printf("Synchronized: %v\n", health["synchronized_accounts"])
	fmt.Printf("Unsynchronized: %v\n", health["unsynchronized_accounts"])
	fmt.Printf("Sync Percentage: %.2f%%\n", health["sync_percentage"])
	fmt.Printf("Total Difference Amount: %.2f\n", health["total_difference_amount"])
	fmt.Printf("Max Difference Amount: %.2f\n", health["max_difference_amount"])
	fmt.Printf("Last Check Time: %v\n", health["last_check_time"])

	status := health["status"].(string)
	switch status {
	case "OK":
		fmt.Println("\n✅ All accounts are properly synchronized!")
	case "WARNING":
		fmt.Println("\n⚠️  Some accounts have balance discrepancies (WARNING level)")
	case "ERROR":
		fmt.Println("\n🚨 Multiple accounts have balance discrepancies (ERROR level)")
	}
}

func runSingleCheck(service *services.BalanceMonitoringService, autoFix bool) {
	fmt.Println("🔍 Running single balance check...")

	result, err := service.CheckBalanceSynchronization()
	if err != nil {
		log.Fatalf("❌ Balance check failed: %v", err)
	}

	service.LogDiscrepancies(result)

	if autoFix && len(result.Discrepancies) > 0 {
		fmt.Println("\n🔧 Auto-fixing discrepancies...")
		if err := service.AutoFixDiscrepancies(result); err != nil {
			log.Printf("❌ Auto-fix failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("✅ Auto-fix completed successfully")
	}

	fmt.Printf("\n📊 SUMMARY:\n")
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Total Accounts: %d\n", result.TotalAccountsChecked)
	fmt.Printf("Synchronized: %d\n", result.SynchronizedAccounts)
	fmt.Printf("Unsynchronized: %d\n", result.UnsynchronizedAccounts)
	
	if result.Status != "OK" {
		os.Exit(1) // Exit with error code if there are issues
	}
}

func runPeriodicMonitoring(service *services.BalanceMonitoringService, interval int, autoFix bool) {
	fmt.Println("🔄 Starting Balance Monitoring Service")
	fmt.Println("=====================================")
	fmt.Printf("Check Interval: %d minutes\n", interval)
	fmt.Printf("Auto-fix Enabled: %v\n", autoFix)
	fmt.Println("Press Ctrl+C to stop monitoring...")
	fmt.Println()

	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Run initial check
	fmt.Println("Running initial check...")
	result, err := service.CheckBalanceSynchronization()
	if err != nil {
		log.Printf("❌ Initial check failed: %v", err)
	} else {
		service.LogDiscrepancies(result)
		if autoFix && len(result.Discrepancies) > 0 {
			if err := service.AutoFixDiscrepancies(result); err != nil {
				log.Printf("❌ Initial auto-fix failed: %v", err)
			}
		}
	}

	// Start periodic monitoring in a goroutine
	go service.RunPeriodicCheck(interval, autoFix)

	// Wait for shutdown signal
	<-c
	fmt.Println("\n🛑 Shutting down Balance Monitoring Service...")
	fmt.Println("✅ Service stopped gracefully")
}
