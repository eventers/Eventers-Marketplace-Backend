package twilio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Sender interface {
	Send(string, string) (string, error)
}

type smsSender struct {
	AccountSID string
	AuthToken  string
	URL        string
	From       string
	HTTPClient http.Client
}

func NewSender(acSID, authToken, url, from string) Sender {
	return &smsSender{
		AccountSID: acSID,
		AuthToken:  authToken,
		URL:        fmt.Sprintf("%s/%s/Messages.json", url, acSID),
		From:       from,
	}
}

func (s *smsSender) Send(to, message string) (string, error) {
	v := url.Values{}
	v.Set("To", to)
	v.Set("From", s.From)
	v.Set("Body", message)

	statusCode, sid, err := s.post(v)
	if err != nil {
		return "", fmt.Errorf("send: error sending sms: status code: %d: err: %s", statusCode, err)
	}
	return *sid, nil
}

func (s *smsSender) post(values url.Values) (*int, *string, error) {
	req, err := http.NewRequest(http.MethodPost, s.URL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, nil, err
	}

	req.SetBasicAuth(s.AccountSID, s.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("post: error reading sms body: %s", err)
		}

		var data map[string]interface{}
		err := json.Unmarshal(bodyBytes, &data)

		if err != nil {
			return nil, nil, fmt.Errorf("post: error unmarshallin g response body: %s", err)
		}
		sid := data["sid"].(string)
		return &res.StatusCode, &sid, nil
	}

	return &res.StatusCode, nil, fmt.Errorf("post: error making post request: %v", err)

}
