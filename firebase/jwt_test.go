package firebase

import (
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestVerifyJWTIDToken(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	idToken := idToken(t)
	idToken=`eyJhbGciOiJSUzI1NiIsImtpZCI6IjhjZjBjNjQyZDQwOWRlODJlY2M5MjI4ZTRiZDc5OTkzOTZiNTY3NDAiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiVmlrcmFtIFNpbmdoIiwicGljdHVyZSI6Imh0dHBzOi8vZ3JhcGguZmFjZWJvb2suY29tLzExOTUwNTI3Mjc5MDk2My9waWN0dXJlIiwiaXNzIjoiaHR0cHM6Ly9zZWN1cmV0b2tlbi5nb29nbGUuY29tL2V2ZW50ZXJzLWRldiIsImF1ZCI6ImV2ZW50ZXJzLWRldiIsImF1dGhfdGltZSI6MTU4MzU2NjI3NCwidXNlcl9pZCI6IkJPZ0FWV1Axa05YNUsydDIwWjI2ZWNFQWtpNjIiLCJzdWIiOiJCT2dBVldQMWtOWDVLMnQyMFoyNmVjRUFraTYyIiwiaWF0IjoxNTgzNTY2NTIwLCJleHAiOjE1ODM1NzAxMjAsImVtYWlsIjoidmlrcmFtc2luZ2hAZXZlbnRlcnNhcHAuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7ImZhY2Vib29rLmNvbSI6WyIxMTk1MDUyNzI3OTA5NjMiXSwiZW1haWwiOlsidmlrcmFtc2luZ2hAZXZlbnRlcnNhcHAuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoiZmFjZWJvb2suY29tIn19.urqZZrw088NnReFS9nCp_sAVPcPQmSRB2hwJwAiMfBkGPWMruaJJ2crcQ0iykQClznL6_mKDJ7raFC-itQvTyV5Je_ZChJE1XQz0iKF23upXtOaJaa_T4Laq-oaCMDaj6Gwf9Y3yjkq1uysDn9BszM7P6HvgCwdfyV8Hk6KfnIYNkng3a5oTM58Nwbc4pWj4CuqX`
	uid, ok := VerifyJWTIDToken(idToken, "eventers-dev", time.Second*10000000)
	require.True(t, ok)

	assert.Equal(t, "foobarbaz", uid)
}

func TestVerifyJWTIDTokenFailsForInvalidKid(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	idToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjBlYTNmN2EwMjQ4YmU0ZTBkZjAyYWVlZWIyMGIxZDJlMmI3ZjI0NzQiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZXZlbnRlcnMtZGV2IiwiYXVkIjoiZXZlbnRlcnMtZGV2IiwiYXV0aF90aW1lIjoxNTgyOTEwODAyLCJ1c2VyX2lkIjoiZm9vYmFyYmF6Iiwic3ViIjoiZm9vYmFyYmF6IiwiaWF0IjoxNTgyOTEwODAyLCJleHAiOjE1ODI5MTQ0MDIsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnt9LCJzaWduX2luX3Byb3ZpZGVyIjoiY3VzdG9tIn19.xNpLGagqlEV8F3O93l-LuKbpIUbdonFBfm-VYQLnBDhWq23CVgOHDU7pZL4GZUQYrQtFr-qtnczBY0CKDDJ8XDCTIFbj8NbiyAiRo7TJLIcRmeDuWj-fPPwuo8bdNFWsdhKl1q-bqTYf5_XOcViZcTtifOTj8JHR7qF1UryVSv4Ji5_8xd1eqCvC1AR23wqeujEDJ1VBhMYlVhGN0SP-lBBrdvADZh6wFsQRjCe6x6muxuvZvNo3WPeDOr0W5YD9gyKuJrdE1o_mHQaoHP_kiogXM4NGumX4Q8A_r_G0_EPnBziA8dTINfCLvmm9VGUyMkATdlOqPo5Hvo62AIGOHw"

	uid, ok := VerifyJWTIDToken(idToken, "", time.Second*1)
	require.False(t, ok)

	assert.Equal(t, "", uid)
}

func TestVerifyJWTIDTokenFailsWhenNoKid(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	idToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZXZlbnRlcnMtZGV2IiwiYXVkIjoiZXZlbnRlcnMtZGV2IiwiYXV0aF90aW1lIjoxNTUyMzIxNTY4LCJ1c2VyX2lkIjoiZm9vYmFyYmF6Iiwic3ViIjoiZm9vYmFyYmF6IiwiaWF0IjoxNTUyMzIxNTY4LCJleHAiOjE1NTIzMjUxNjgsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnt9LCJzaWduX2luX3Byb3ZpZGVyIjoiY3VzdG9tIn19.jUR50OoXaBHSBscon61n3nqsehRG4XWZg2HL5MNhIbJkGjm2rsLPgKv62aLlG69OVvnCCO15pfMRJb_Hz4drJ8cCiLR1oMG4Kl7kzqoUd1DsB85rSQRQOj87Vw-2Y1LultDdkKimop9EQ3VU3Q4UkZZX2IA2_iKyKg8ckk-48VOK0--QjvOKqR2mFsm037CiQc_pw2SLxyKik8gN1auF-7teoMtb_QvQmuF2Ei24llLkV71llWgYEO2qE5wRPD-R7qetlu6cpxqDDvkI7x6m0fMUNGCVTc2B70Gol7PpCEJECCAuQKWSEIb-tAwtquPwtaeDuXrTrw1jl-ASJ90STA"

	uid, ok := VerifyJWTIDToken(idToken, "", time.Second*1)
	require.False(t, ok)

	assert.Equal(t, "", uid)
}

func TestVerifyJWTIDTokenFailsWhenEmptyJWTToken(t *testing.T) {
	if !hasTestData {
		t.Skipf("TestVerifyToken: skipping test as the testdata does not exists")
	}

	uid, ok := VerifyJWTIDToken("", "", time.Second*1)
	require.False(t, ok)

	assert.Equal(t, "", uid)
}

func idToken(t *testing.T) string {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("testdata/serviceAccountKey.json"))
	require.Nil(t, err, "expected err to be nil")

	client, err := app.Auth(ctx)
	require.Nil(t, err, "expected err to be nil")

	customToken, err := client.CustomToken(ctx, "foobarbaz")
	require.Nil(t, err, "expected err to be nil")

	idToken, err := signInWithCustomToken(customToken)
	require.Nil(t, err, "expected err to be nil")

	return idToken
}
