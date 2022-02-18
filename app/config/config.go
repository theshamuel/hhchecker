package config

import (
	"fmt"
	"github.com/theshamuel/hhchecker/provider"
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
	var fileConf struct {
		URL       string        `yaml:"url"`
		Timeout   time.Duration `yaml:"timeout,omitempty"`
		MaxAlerts int8          `yaml:"max-alerts,omitempty"`
		Debug     bool          `yaml:"debug,omitempty"`
	}
	f, err := os.Open(s.FileName)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", s.FileName, err)
	}
	defer f.Close() //nolint gosec
	if err = yaml.NewDecoder(f).Decode(&fileConf); err != nil {
		return nil, fmt.Errorf("can't parse %s: %w", s.FileName, err)
	}

	return &CommonOpts{
		URL:       fileConf.URL,
		Timeout:   fileConf.Timeout,
		MaxAlerts: fileConf.MaxAlerts,
		Debug:     fileConf.Debug,
	}, nil
}
func (s *Config) GetProviders(client *http.Client) ([]provider.Interface, error) {
	s.Lock()
	defer s.Unlock()
	var fileConf struct {
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
				API struct {
					URL string `yaml:"url,omitempty"`
					Key string `yaml:"key,omitempty"`
				} `yaml:"api,omitempty"`
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

	f, err := os.Open(s.FileName)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", s.FileName, err)
	}
	defer f.Close() //nolint gosec
	if err = yaml.NewDecoder(f).Decode(&fileConf); err != nil {
		return nil, fmt.Errorf("can't parse %s: %w", s.FileName, err)
	}
	var providers []provider.Interface

	if fileConf.Email.Enabled {
		providers = append(providers, &provider.Mailgun{
			Values: map[string]io.Reader{
				"from":    strings.NewReader(fileConf.Email.From),
				"to":      strings.NewReader(fileConf.Email.To),
				"cc":      strings.NewReader(fileConf.Email.Cc),
				"subject": strings.NewReader(fileConf.Email.Subject),
				"text":    strings.NewReader(fileConf.Email.Text),
			},
			URL:    fileConf.URL,
			APIKey: fileConf.Email.Mailgun.API.Key,
			Provider: provider.Provider{
				ID:     provider.PIDMailgun,
				Client: client,
			},
		})
	}

	if fileConf.Telegram.Enabled {
		providers = append(providers, &provider.Telegram{
			BotAPIKey:   fileConf.Telegram.BotAPIKey,
			ChannelID:   fileConf.Telegram.Channel.ID,
			ChannelName: fileConf.Telegram.Channel.Name,
			Message:     fileConf.Telegram.Message,
			Provider: provider.Provider{
				ID:     provider.PIDTelegram,
				Client: client,
			},
		})
	}

	return providers, nil
}
