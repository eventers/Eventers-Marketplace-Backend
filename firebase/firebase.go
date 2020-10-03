package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
)

func VerifyIDToken(ctx context.Context, app *firebase.App, idToken string) (*auth.Token, error) {
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("VerifyIDToken: error getting Auth client: %v\n", err)
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("VerifyIDToken: error verifying ID token: %v\n", err)
	}

	return token, nil
}
