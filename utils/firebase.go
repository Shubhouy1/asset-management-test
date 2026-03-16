package utils

import (
	"context"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App
var FirebaseAuth *auth.Client

func InitFirebase() error {

	serviceAccount := os.Getenv("FIREBASE_SERVICE_ACCOUNT")

	opt := option.WithCredentialsFile(serviceAccount)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return err
	}

	FirebaseApp = app
	FirebaseAuth = authClient

	return nil
}
