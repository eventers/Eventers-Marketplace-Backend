package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"firebase.google.com/go"
	"google.golang.org/api/option"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	verifyCustomTokenURL = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyCustomToken?key=%s"
)

var hasTestData bool

func init() {
	if _, err := os.Stat("testdata"); !os.IsNotExist(err) {
		hasTestData = true
	}

	if _, err := os.Stat("/testdata"); !os.IsNotExist(err) {
		hasTestData = true
	}

	if _, err := os.Stat("./testdata"); !os.IsNotExist(err) {
		hasTestData = true
	}
}

func TestVerifyIDToken(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("testdata/serviceAccountKey.json"))
	require.Nil(t, err, "expected err to be nil")

	client, err := app.Auth(ctx)
	require.Nil(t, err, "expected err to be nil")

	customToken, err := client.CustomToken(ctx, "aWBjlMGvUYYp6HheJPryIS0VwME2")
	require.Nil(t, err, "expected err to be nil")

	idToken, err := signInWithCustomToken(customToken)
	require.Nil(t, err, "expected err to be nil")

	token, err := VerifyIDToken(ctx, app, idToken)

	require.Nil(t, err, "expected err to be nil")
	assert.Equal(t, token.Issuer, "https://securetoken.google.com/eventers-dev")
	assert.Equal(t, token.Subject, "r3q9th49t7g97")
	assert.Equal(t, token.UID, "r3q9th49t7g97")
}

func TestVerifyIDTokenFailsToGetAuthClient(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("testdata/serviceAccountKeyInvalid.json"))
	require.Nil(t, err, "expected err to be nil")

	token, err := VerifyIDToken(ctx, app, "")
	require.Nil(t, token, "expected token to be nil")

	assert.Containsf(t, err.Error(), "VerifyIDToken: error getting Auth client:", "")
}
func TestVerifyIDTokenFailsToVerifyToken(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("testdata/serviceAccountKey.json"))
	require.Nil(t, err, "expected err to be nil")

	token, err := VerifyIDToken(ctx, app, "")
	require.Nil(t, token, "expected token to be nil")

	assert.Containsf(t, err.Error(), "VerifyIDToken: error verifying ID token:", "")
}

func signInWithCustomToken(token string) (string, error) {
	req, err := json.Marshal(map[string]interface{}{
		"token":             token,
		"returnSecureToken": true,
	})
	if err != nil {
		return "", err
	}

	apiKey, err := apiKey()
	if err != nil {
		return "", err
	}
	resp, err := postRequest(fmt.Sprintf(verifyCustomTokenURL, apiKey), req)
	if err != nil {
		return "", err
	}
	var respBody struct {
		IDToken string `json:"idToken"`
	}
	if err := json.Unmarshal(resp, &respBody); err != nil {
		return "", err
	}
	return respBody.IDToken, err
}

func apiKey() (string, error) {
	b, err := ioutil.ReadFile("testdata/apiKey.txt")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func postRequest(url string, req []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status code: %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
