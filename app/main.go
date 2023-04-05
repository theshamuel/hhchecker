package main

import (
	"fmt"
	"github.com/hashicorp/logutils"
	"github.com/theshamuel/go-flags"
	"github.com/theshamuel/hhchecker/app/config"
	"github.com/theshamuel/hhchecker/app/provider"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var opts struct {
	config.CommonOpts
	Email struct {
		Enabled       bool   `long:"enabled" env:"ENABLED" description:"enable email mailgun provider"`
		From          string `long:"from" env:"FROM" description:"the source email address"`
		To            string `long:"to" env:"TO" description:"the target email address"`
		Cc            string `long:"cc" env:"CC" description:"the cc email address"`
		Subject       string `long:"subject" env:"SUBJECT" description:"the subject of email"`
		Text          string `long:"text" env:"TEXT" description:"the text of email not more 255 letters"`
		Domain        string `long:"domain" env:"DOMAIN" description:"the mailgun API URL for sending notification"`
		MailgunAPIKey string `long:"mailgunApiKey" env:"MAILGUN_API_KEY" description:"the token for mailgun api"`
	} `group:"email" namespace:"email" env-namespace:"EMAIL"`

	Telegram struct {
		Enabled     bool   `long:"enabled" env:"ENABLED" description:"enable telegram provider"`
		BotAPIKey   string `long:"botApiKey" env:"BOT_API_KEY" description:"the telegram bot api key"`
		ChannelName string `long:"channelName" env:"CHANNEL_NAME" description:"the channel name without leading symbol @ for public channel only"`
		ChannelID   string `long:"channelId" env:"CHANNEL_ID" description:"the channel id for private channel only"`
		Message     string `long:"message" env:"MESSAGE" description:"the text message not more 255 letters"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Config struct {
		Enabled  bool   `long:"enabled" env:"ENABLED" description:"enable getting parameters from config. In that case all parameters will be read only form config"`
		FileName string `long:"file-name" env:"FILE_NAME" default:"hhchecker.yml" description:"config file name"`
	} `group:"config" namespace:"config" env-namespace:"CONFIG"`
}

var version = "unknown"

func main() {
	parseFlags()
	if opts.Config.Enabled {
		cnf := &config.Config{
			FileName: opts.Config.FileName,
		}
		var err error
		var co *config.CommonOpts
		if co, err = cnf.GetCommon(); err != nil {
			panic(fmt.Errorf("[ERROR] can not read config file, %w", err))
		}
		opts.URL = co.URL
		opts.Debug = co.Debug
		opts.Timeout = co.Timeout
		opts.MaxAlerts = co.MaxAlerts
		log.Printf("[DEBUG] config: %+v", cnf.File)
	}

	setupLogLevel(opts.CommonOpts.Debug)

	log.Printf("[INFO] Starting Health checker for %s:%s ...\n", opts.CommonOpts.URL, version)
	var client = &http.Client{Timeout: 3 * time.Second}
	var maxAlerts = int8(0)
	var providers []provider.Interface

	if opts.Email.Enabled {
		providers = append(providers, &provider.Mailgun{
			Values: map[string]io.Reader{
				"from":    strings.NewReader(opts.Email.From),
				"to":      strings.NewReader(opts.Email.To),
				"cc":      strings.NewReader(opts.Email.Cc),
				"subject": strings.NewReader(opts.Email.Subject),
				"text":    strings.NewReader(opts.Email.Text),
			},
			Domain: opts.Email.Domain,
			APIKey: opts.Email.MailgunAPIKey,
			Provider: provider.Provider{
				ID:     provider.PIDMailgun,
				Client: client,
			},
		})
	}

	if opts.Telegram.Enabled {
		providers = append(providers, &provider.Telegram{
			BotAPIKey:   opts.Telegram.BotAPIKey,
			ChannelID:   opts.Telegram.ChannelID,
			ChannelName: opts.Telegram.ChannelName,
			Message:     opts.Telegram.Message,
			Provider: provider.Provider{
				ID:     provider.PIDTelegram,
				Client: client,
			},
		})
	}

	if opts.Config.Enabled {
		var err error
		cnf := config.Config{
			FileName: opts.Config.FileName,
		}
		if providers, err = cnf.GetProviders(client); err != nil {
			panic(fmt.Errorf("[ERROR] can not read config file, %w", err))
		}
	}

	log.Printf("[DEBUG] cli options: %+v", opts)

	for range time.Tick(opts.CommonOpts.Timeout) {
		response, err := http.Get(opts.CommonOpts.URL)
		log.Printf("[DEBUG] response: %+v", response)
		if err != nil || response.StatusCode != 200 {
			if maxAlerts >= opts.CommonOpts.MaxAlerts {
				for _, provider := range providers {
					if err := provider.Send(); err != nil {
						log.Printf("[ERROR] error occurs during sending [%s] message: %+v", provider.GetID(), err)
					}
				}
				maxAlerts = 0
			}
			maxAlerts++
		} else if response != nil && response.StatusCode == 200 {
			maxAlerts = 0
		}
	}
}

func parseFlags() {
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func setupLogLevel(debug bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}
	log.SetFlags(log.Ldate | log.Ltime)

	if debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}
	log.SetOutput(filter)
}

func getStackTrace() string {
	maxSize := 7 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] Singal QUITE is cought , stacktrace [\n%s", getStackTrace())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
