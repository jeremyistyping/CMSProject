package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Payment Code Sequential Fix Tool")
	fmt.Println("===================================")

	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Step 1: Create payment_code_sequences table if not exists
	fmt.Println("📝 Step 1: Creating payment_code_sequences table...")
	err = database.CreatePaymentSequenceTables(db)
	if err != nil {
		log.Printf("❌ Failed to create payment sequence tables: %v", err)
		return
	}
	fmt.Println("✅ Payment sequence tables ready")

	// Step 2: Check current payment codes
	fmt.Println("\n📝 Step 2: Analyzing current payment codes...")
	var allPayments []models.Payment
	err = db.Order("created_at ASC").Find(&allPayments).Error
	if err != nil {
		log.Printf("❌ Failed to fetch payments: %v", err)
		return
	}

	fmt.Printf("Found %d payment records\n", len(allPayments))

	// Categorize payment codes
	var sequentialCodes []models.Payment
	var timestampCodes []models.Payment
	var otherCodes []models.Payment

	sequentialPattern := regexp.MustCompile(`^(RCV|PAY)/(\d{4})/(\d{2})/(\d{4})$`)
	timestampPattern := regexp.MustCompile(`TS\d{6}`)

	for _, payment := range allPayments {
		if sequentialPattern.MatchString(payment.Code) {
			sequentialCodes = append(sequentialCodes, payment)
		} else if timestampPattern.MatchString(payment.Code) {
			timestampCodes = append(timestampCodes, payment)
		} else {
			otherCodes = append(otherCodes, payment)
		}
	}

	fmt.Printf("📊 Code Analysis:\n")
	fmt.Printf("  Sequential codes (RCV/YYYY/MM/NNNN): %d\n", len(sequentialCodes))
	fmt.Printf("  Timestamp codes (contains TS): %d\n", len(timestampCodes))
	fmt.Printf("  Other/Invalid codes: %d\n", len(otherCodes))

	// Step 3: Initialize sequences from existing sequential codes
	fmt.Println("\n📝 Step 3: Initializing sequences from existing data...")
	err = initializeSequencesFromData(db)
	if err != nil {
		log.Printf("❌ Failed to initialize sequences: %v", err)
		return
	}

	// Step 4: Fix non-sequential payment codes
	fmt.Println("\n📝 Step 4: Fixing non-sequential payment codes...")
	
	totalToFix := len(timestampCodes) + len(otherCodes)
	if totalToFix == 0 {
		fmt.Println("✅ All payment codes are already sequential!")
		return
	}

	fmt.Printf("Need to fix %d payment codes\n", totalToFix)

	fixCount := 0
	
	// Fix timestamp codes
	for _, payment := range timestampCodes {
		newCode, err := generateSequentialCode(db, payment.CreatedAt)
		if err != nil {
			log.Printf("❌ Failed to generate code for payment %d: %v", payment.ID, err)
			continue
		}

		// Update payment code
		err = db.Model(&payment).Update("code", newCode).Error
		if err != nil {
			log.Printf("❌ Failed to update payment %d code: %v", payment.ID, err)
			continue
		}

		fmt.Printf("✅ Fixed payment %d: %s -> %s\n", payment.ID, payment.Code, newCode)
		fixCount++
	}

	// Fix other invalid codes  
	for _, payment := range otherCodes {
		newCode, err := generateSequentialCode(db, payment.CreatedAt)
		if err != nil {
			log.Printf("❌ Failed to generate code for payment %d: %v", payment.ID, err)
			continue
		}

		// Update payment code
		err = db.Model(&payment).Update("code", newCode).Error
		if err != nil {
			log.Printf("❌ Failed to update payment %d code: %v", payment.ID, err)
			continue
		}

		fmt.Printf("✅ Fixed payment %d: %s -> %s\n", payment.ID, payment.Code, newCode)
		fixCount++
	}

	fmt.Printf("\n🎉 Summary: Fixed %d out of %d payment codes\n", fixCount, totalToFix)

	// Step 5: Verify results
	fmt.Println("\n📝 Step 5: Verifying results...")
	verifyPaymentCodes(db)
}

func initializeSequencesFromData(db *gorm.DB) error {
	// Get all sequential payment codes to determine current sequences
	var results []struct {
		Prefix string
		Year   int
		Month  int
		MaxSeq int
	}

	query := `
		SELECT 
			SUBSTRING_INDEX(SUBSTRING_INDEX(code, '/', 1), '/', -1) as prefix,
			CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(code, '/', 2), '/', -1) AS UNSIGNED) as year,
			CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(code, '/', 3), '/', -1) AS UNSIGNED) as month,
			MAX(CAST(SUBSTRING_INDEX(code, '/', -1) AS UNSIGNED)) as max_seq
		FROM payments 
		WHERE code REGEXP '^(RCV|PAY)/[0-9]{4}/[0-9]{2}/[0-9]{4}$'
		AND deleted_at IS NULL
		GROUP BY prefix, year, month
	`

	err := db.Raw(query).Scan(&results).Error
	if err != nil {
		return err
	}

	fmt.Printf("Found %d existing sequence groups\n", len(results))

	for _, result := range results {
		sequence := models.PaymentCodeSequence{
			Prefix:         result.Prefix,
			Year:           result.Year,
			Month:          result.Month,
			SequenceNumber: result.MaxSeq,
		}

		// Use UPSERT
		err = db.Where("prefix = ? AND year = ? AND month = ?", 
			result.Prefix, result.Year, result.Month).
			Assign(&sequence).FirstOrCreate(&sequence).Error
		if err != nil {
			log.Printf("❌ Failed to initialize sequence %s/%d/%02d: %v", 
				result.Prefix, result.Year, result.Month, err)
		} else {
			fmt.Printf("✅ Initialized sequence %s/%d/%02d = %d\n", 
				result.Prefix, result.Year, result.Month, result.MaxSeq)
		}
	}

	return nil
}

func generateSequentialCode(db *gorm.DB, createdAt time.Time) (string, error) {
	// Determine prefix based on created date and typical usage
	// For simplicity, assume most payments are receivables (RCV)
	prefix := "RCV"
	year := createdAt.Year()
	month := int(createdAt.Month())

	// Get or create sequence
	var sequence models.PaymentCodeSequence
	err := db.Where("prefix = ? AND year = ? AND month = ?", prefix, year, month).
		First(&sequence).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sequence starting at 1
			sequence = models.PaymentCodeSequence{
				Prefix:         prefix,
				Year:           year,
				Month:          month,
				SequenceNumber: 1,
			}
			err = db.Create(&sequence).Error
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		// Increment existing sequence
		sequence.SequenceNumber++
		err = db.Save(&sequence).Error
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s/%04d/%02d/%04d", prefix, year, month, sequence.SequenceNumber), nil
}

func verifyPaymentCodes(db *gorm.DB) {
	var allPayments []models.Payment
	db.Order("created_at ASC").Find(&allPayments)

	sequentialPattern := regexp.MustCompile(`^(RCV|PAY)/(\d{4})/(\d{2})/(\d{4})$`)
	timestampPattern := regexp.MustCompile(`TS\d{6}`)

	sequential := 0
	timestamp := 0
	other := 0

	for _, payment := range allPayments {
		if sequentialPattern.MatchString(payment.Code) {
			sequential++
		} else if timestampPattern.MatchString(payment.Code) {
			timestamp++
		} else {
			other++
		}
	}

	fmt.Printf("📊 Final Results:\n")
	fmt.Printf("  Sequential codes: %d\n", sequential)
	fmt.Printf("  Timestamp codes: %d\n", timestamp)
	fmt.Printf("  Other codes: %d\n", other)

	if timestamp == 0 && other == 0 {
		fmt.Println("🎉 All payment codes are now sequential!")
	} else {
		fmt.Printf("⚠️  Still have %d non-sequential codes\n", timestamp+other)
	}
}