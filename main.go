package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PhoneNumber struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EncryptedData string    `json:"encrypted_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type KMSClient struct {
	Client  *kms.KeyManagementClient
	KeyPath string
}

var (
	db *gorm.DB
	// Map to store different KMS clients for different purposes
	kmsClients = make(map[string]*KMSClient)
)

func initKMSClient(ctx context.Context, serviceAccountPath, keyPath, clientName string) error {
	// Read the service account key file
	serviceAccountJSON, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return fmt.Errorf("failed to read service account file for %s: %v", clientName, err)
	}

	// Create credentials from the JSON content
	credentials, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, "https://www.googleapis.com/auth/cloudkms")
	if err != nil {
		return fmt.Errorf("failed to load credentials for %s: %v", clientName, err)
	}

	// Create KMS client
	client, err := kms.NewKeyManagementClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return fmt.Errorf("failed to create KMS client for %s: %v", clientName, err)
	}

	// Store the client and key path
	kmsClients[clientName] = &KMSClient{
		Client:  client,
		KeyPath: keyPath,
	}

	return nil
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize KMS clients for different purposes
	ctx := context.Background()

	// Initialize phone number encryption KMS client
	if err := initKMSClient(
		ctx,
		os.Getenv("PHONE_ENCRYPTION_SERVICE_ACCOUNT_PATH"),
		os.Getenv("PHONE_ENCRYPTION_KEY_PATH"),
		"phone_encryption",
	); err != nil {
		log.Fatal(err)
	}

	// Auto-migrate database
	if err := db.AutoMigrate(&PhoneNumber{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

func encryptData(plaintext []byte) (string, error) {
	ctx := context.Background()
	kmsClient := kmsClients["phone_encryption"]
	if kmsClient == nil {
		return "", fmt.Errorf("phone encryption KMS client not initialized")
	}

	// Encrypt the data
	req := &kmspb.EncryptRequest{
		Name:      kmsClient.KeyPath,
		Plaintext: plaintext,
	}

	resp, err := kmsClient.Client.Encrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Return base64 encoded encrypted data
	return base64.StdEncoding.EncodeToString(resp.Ciphertext), nil
}

func decryptData(encryptedData string) (string, error) {
	ctx := context.Background()
	kmsClient := kmsClients["phone_encryption"]
	if kmsClient == nil {
		return "", fmt.Errorf("phone encryption KMS client not initialized")
	}

	// Decode base64 encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Decrypt the data
	req := &kmspb.DecryptRequest{
		Name:       kmsClient.KeyPath,
		Ciphertext: ciphertext,
	}

	resp, err := kmsClient.Client.Decrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(resp.Plaintext), nil
}

func main() {
	r := gin.Default()

	// POST endpoint to save encrypted phone number
	r.POST("/phone", func(c *gin.Context) {
		var input struct {
			PhoneNumber string `json:"phone_number" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Encrypt the phone number
		encryptedData, err := encryptData([]byte(input.PhoneNumber))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to encrypt phone number", "details": err.Error()})
			return
		}

		// Save to the database
		phone := PhoneNumber{
			ID:            uuid.New().String(),
			EncryptedData: encryptedData,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := db.Create(&phone).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to save phone number", "details": err.Error()})
			return
		}

		c.JSON(201, phone)
	})

	// GET endpoint to retrieve decrypted phone number
	r.GET("/phone/:id", func(c *gin.Context) {
		id := c.Param("id")

		var phone PhoneNumber
		if err := db.First(&phone, "id = ?", id).Error; err != nil {
			c.JSON(404, gin.H{"error": "The Phone number not found", "details": err.Error()})
			return
		}

		// Decrypt the phone number
		decryptedData, err := decryptData(phone.EncryptedData)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to decrypt phone number", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"id":           phone.ID,
			"phone_number": decryptedData,
			"created_at":   phone.CreatedAt,
			"updated_at":   phone.UpdatedAt,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
