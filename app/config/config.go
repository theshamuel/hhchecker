package config

import (
	"fmt"
	"github.com/theshamuel/hhchecker/app/provider"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	FileName string
	sync.Mutex
	File *File
}

type File struct {
	URL       string        `yaml:"url"`
	Timeout   time.Duration `yaml:"timeout,omitempty"`
	MaxAlerts int8          `yaml:"max-alerts,omitempty"`
	Debug     bool          `yaml:"debug,omitempty"`
	Email     struct {
		Enabled bool   `yaml:"enabled,omitempty"`
		From    string `yaml:"from,omitempty"`
		To      string `yaml:"to,omitempty"`
		Cc      string `yaml:"cc,omitempty"`
		Subject string `yaml:"subject,omitempty"`
		Text    string `yaml:"text,omitempty"`
		Mailgun struct {
			Domain string `yaml:"domain,omitempty"`
			APIKey string `yaml:"api-key,omitempty"`
		} `yaml:"mailgun,omitempty"`
	} `yaml:"email,omitempty"`
	Telegram struct {
		Enabled   bool   `yaml:"enabled,omitempty"`
		BotAPIKey string `yaml:"bot-api-key,omitempty"`
		Message   string `yaml:"message,omitempty"`
		Channel   struct {
			Name string `yaml:"name,omitempty"`
			ID   string `yaml:"id,omitempty"`
		} `yaml:"channel,omitempty"`
	} `yaml:"telegram,omitempty"`
}

type CommonOpts struct {
	URL       string        `long:"url" env:"URL" description:"the URL what you need to healthcheck"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" default:"300s" description:"the timeout for health probe in seconds"`
	MaxAlerts int8          `long:"max-alerts" env:"MAX_ALERTS" default:"3" description:"the max count of alerts in sequence"`
	Debug     bool          `long:"debug" env:"DEBUG" description:"debug mode"`
}

func (s *Config) GetCommon() (*CommonOpts, error) {
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(s.FileName)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", s.FileName, err)
	}
	defer f.Close()
	if err = yaml.NewDecoder(f).Decode(&s.File); err != nil {
		return nil, fmt.Errorf("can't parse %s: %w", s.FileName, err)
	}

	return &CommonOpts{
		URL:       s.File.URL,
		Timeout:   s.File.Timeout,
		MaxAlerts: s.File.MaxAlerts,
		Debug:     s.File.Debug,
	}, nil
}

func (s *Config) GetProviders(client *http.Client) ([]provider.Interface, error) {
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(s.FileName)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", s.FileName, err)
	}
	defer f.Close()
	if err = yaml.NewDecoder(f).Decode(&s.File); err != nil {
		return nil, fmt.Errorf("can't parse %s: %w", s.FileName, err)
	}
	var providers []provider.Interface

	if s.File.Email.Enabled {
		providers = append(providers, &provider.Mailgun{
			Values: map[string]io.Reader{
				"from":    strings.NewReader(s.File.Email.From),
				"to":      strings.NewReader(s.File.Email.To),
				"cc":      strings.NewReader(s.File.Email.Cc),
				"subject": strings.NewReader(s.File.Email.Subject),
				"text":    strings.NewReader(s.File.Email.Text),
			},
			Domain: s.File.Email.Mailgun.Domain,
			APIKey: s.File.Email.Mailgun.APIKey,
			Provider: provider.Provider{
				ID:     provider.PIDMailgun,
				Client: client,
			},
		})
	}

	if s.File.Telegram.Enabled {
		providers = append(providers, &provider.Telegram{
			BotAPIKey:   s.File.Telegram.BotAPIKey,
			ChannelID:   s.File.Telegram.Channel.ID,
			ChannelName: s.File.Telegram.Channel.Name,
			Message:     s.File.Telegram.Message,
			Provider: provider.Provider{
				ID:     provider.PIDTelegram,
				Client: client,
			},
		})
	}

	return providers, nil
}
