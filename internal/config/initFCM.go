package config

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

func InitFCM(conn string) error {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(conn))
	app, err := firebase.NewApp(ctx, &firebase.Config{}, opt)
	if err != nil {
		return err
	}
	fcmClient, err = app.Messaging(ctx)
	if err != nil {
		return err
	}
	return nil
}

func GetFCMClient() *messaging.Client {
	return fcmClient
}
