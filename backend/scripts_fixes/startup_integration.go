package main

import (
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

// StartupIntegration initializes the automatic balance synchronization system
func StartupIntegration() {
	log.Println("🚀 Initializing automatic balance synchronization system...")

	// Connect to database
	db := database.ConnectDB()

	// Initialize balance sync service
	balanceSync := services.NewBalanceSyncService(db)

	// Run initial integrity check and sync if needed
	log.Println("🔍 Running startup balance integrity check...")
	isConsistent, err := balanceSync.VerifyBalanceIntegrity()
	if err != nil {
		log.Printf("❌ Error during startup integrity check: %v", err)
	} else if !isConsistent {
		log.Println("⚠️  Balance inconsistencies detected on startup, running full synchronization...")
		err = balanceSync.SyncAccountBalancesFromSSOT()
		if err != nil {
			log.Printf("❌ Error during startup balance sync: %v", err)
		} else {
			log.Println("✅ Startup balance synchronization completed successfully")
		}
	} else {
		log.Println("✅ All balances are consistent on startup")
	}

	// Start periodic balance sync in background (every 30 minutes)
	go func() {
		log.Println("🔄 Starting periodic balance sync service (every 30 minutes)...")
		balanceSync.SchedulePeriodicSync(30)
	}()

	log.Println("🎯 Balance synchronization system initialized successfully")
}

func main() {
	StartupIntegration()
	
	// Keep the service running
	log.Println("💤 Balance sync service is running in background...")
	select {} // Block forever
}