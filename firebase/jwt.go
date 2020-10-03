package firebase

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	validationErrorExpired = "Token is expired"
)

var CertsAPIEndpoint = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

func checkInterval(claims jwt.MapClaims, interval time.Duration) bool {
	var ok bool
	switch exp := claims["exp"].(type) {
	case float64:
		t1 := time.Unix(int64(exp), 0)
		fmt.Printf("\n%v\n", t1)
		t2 := time.Now().Add(interval * -1)
		ok = t2.Before(t1)
	case json.Number:
		v, _ := exp.Int64()
		t1 := time.Unix(int64(v), 0)
		t2 := time.Now().Add(interval * -1)
		ok = t2.Before(t1)
	}

	return ok
}

func VerifyJWTIDToken(t, projectID string, interval time.Duration) (uid string, ok bool) {
	parsed, err := jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		cert, err := getCertificateFromToken(t)
		if err != nil {
			return "", err
		}
		publicKey, err := readPublicKey(cert)
		if err != nil {
			return "", err
		}
		return publicKey, nil
	})

	if err != nil && err.Error() == validationErrorExpired {
		claims, valid := parsed.Claims.(jwt.MapClaims)
		if !valid {
			err = fmt.Errorf("VerifyJWTIDToken: could not parse claims")
			return
		}

		if ok = checkInterval(claims, interval); ok {
			err = nil
			uid, ok = claims["sub"].(string)
			if !ok {
				return
			}
			return
		}
	}

	if err != nil {
		return
	}

	ok = parsed.Valid
	if !ok {
		return
	}

	if parsed.Header["alg"] != "RS256" {
		ok = false
		return
	}

	ok, uid = verifyPayload(parsed, projectID)
	return
}

func getCertificates() (certs map[string]string, err error) {
	res, err := http.Get(CertsAPIEndpoint)
	if err != nil {
		return
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	json.Unmarshal(data, &certs)
	return
}

func getCertificate(kid string) (cert []byte, err error) {
	certs, err := getCertificates()
	if err != nil {
		return
	}

	certString := certs[kid]
	cert = []byte(certString)
	err = nil

	return
}

func getCertificateFromToken(token *jwt.Token) ([]byte, error) {
	kid, ok := token.Header["kid"]
	if !ok {
		return []byte{}, errors.New("kid not found")
	}

	kidString, ok := kid.(string)
	if !ok {
		return []byte{}, errors.New("kid cast error to string")
	}

	return getCertificate(kidString)
}

func verifyPayload(t *jwt.Token, projectID string) (ok bool, uid string) {
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	claimsAud, ok := claims["aud"].(string)
	if claimsAud != projectID || !ok {
		return
	}

	iss := "https://securetoken.google.com/" + projectID
	claimsIss, ok := claims["iss"].(string)
	if claimsIss != iss || !ok {
		return
	}

	uid, ok = claims["sub"].(string)
	if !ok {
		return
	}

	now := time.Now()

	authTime, ok := claims["auth_time"].(float64)
	if !ok {
		return
	}

	tm := time.Unix(int64(authTime), 0)

	if !tm.Before(now) {
		return false, ""
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return
	}

	iaTime := time.Unix(int64(iat), 0)

	if !iaTime.Before(now) {
		return false, ""
	}

	if claims.Valid() != nil {
		return false, ""
	}

	return
}

func readPublicKey(cert []byte) (*rsa.PublicKey, error) {
	publicKeyBlock, _ := pem.Decode(cert)

	if publicKeyBlock == nil {
		return nil, errors.New("invalid public key data")
	}

	if publicKeyBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid public key type: %s", publicKeyBlock.Type)
	}

	c, err := x509.ParseCertificate(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := c.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}

	return publicKey, nil
}
