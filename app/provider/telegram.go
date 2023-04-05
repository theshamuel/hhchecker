package provider

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// Telegram provider structure for sending email notification
type Telegram struct {
	BotAPIKey   string
	ChannelID   string
	ChannelName string
	Message     string
	Provider    Provider
}

// Send sending text message into public telegram channel
func (s *Telegram) Send() error {
	urlPattern := "https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s"
	channel := s.ChannelID
	if len(channel) == 0 && len(s.ChannelName) > 0 {
		channel = "@" + s.ChannelName
	}
	if len(channel) == 0 {
		return fmt.Errorf("channel ID and channel name were not found")
	}
	log.Printf("[DEBUG] telegram url: %s", fmt.Sprintf(urlPattern, s.BotAPIKey, channel, s.Message))
	req, err := http.NewRequest("GET", fmt.Sprintf(urlPattern, s.BotAPIKey, channel, s.Message), nil)
	if err != nil {
		return err
	}
	res, err := s.Provider.Client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Printf("[ERROR] Telegram response bad status: %s\n", res.Status)
		log.Printf("[ERROR] Telegram response bad body: %s\n", string(body))
	}
	return nil
}

// GetID get Provider ID
func (s *Telegram) GetID() ID {
	return s.Provider.GetID()
}
