package config

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

type FirebaseApp interface {
	Messaging(context.Context) (*messaging.Client, error)
}

var firebaseNewApp func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (FirebaseApp, error)

func init() {
	firebaseNewApp = func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (FirebaseApp, error) {
		app, err := firebase.NewApp(ctx, config, opts...)
		if err != nil {
			var nilApp FirebaseApp = nil
			return nilApp, err
		}
		return app, nil
	}
}

var FirebaseNewApp = &firebaseNewApp

func InitFCM(conn string) error {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(conn))

	app, err := firebaseNewApp(ctx, nil, opt) // Pass nil config explicitly
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

func ResetFCMState() { // Make it exported if tests are in different package
	fcmClient = nil
	firebaseNewApp = func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (FirebaseApp, error) {
		app, err := firebase.NewApp(ctx, config, opts...)
		if err != nil {
			var nilApp FirebaseApp = nil
			return nilApp, err
		}
		return app, nil
	}
}
