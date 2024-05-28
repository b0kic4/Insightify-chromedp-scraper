package utils

import (
	"context"
	"encoding/base64"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"google.golang.org/api/option"
)

func NewFirebaseClient(ctx context.Context) *storage.Client {
	encodedCredentials := os.Getenv("FIREBASE_CREDENTIALS_BASE64")
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		log.Fatalf("Failed to decode Firebase credentials: %v", err)
	}

	opt := option.WithCredentialsJSON(decodedBytes)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	storage, err := app.Storage(ctx)
	if err != nil {
		log.Fatalf("error initializing Firebase Storage: %v", err)
	}

	return storage
}
