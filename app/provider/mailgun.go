package provider

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

// Mailgun provider structure for sending email notification
type Mailgun struct {
	Domain   string
	APIKey   string
	Values   map[string]io.Reader
	Provider Provider
}

// Send sending email via MailGun
func (s *Mailgun) Send() (err error) {
	url := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", s.Domain)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range s.Values {
		var fw io.Writer
		if fw, err = w.CreateFormField(key); err != nil {
			return err
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	credentials := strings.Split(s.APIKey, ":")
	if len(credentials) != 2 {
		return fmt.Errorf("[ERROR]: MAILGUN_API_KEY is not valid")
	}
	req.SetBasicAuth(credentials[0], credentials[1])
	res, err := s.Provider.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Printf("[ERROR] Mailgun response bad status: %s\n", res.Status)
		log.Printf("[ERROR] Mailgun response bad body: %s\n", string(body))
	}

	return err
}

// GetID get Provider ID
func (s *Mailgun) GetID() ID {
	return s.Provider.GetID()
}
