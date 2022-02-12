package main

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var opts struct {
	URL     string        `long:"url" env:"URL" required:"true" description:"the URL what you need to healthcheck"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"300s" description:"the timeout for health probe in seconds"`

	Email struct {
		Enabled       bool   `long:"enabled" env:"ENABLED" description:"enable email mailgun provider"`
		From          string `long:"from" env:"FROM"description:"the source email address"`
		To            string `long:"to" env:"TO" description:"the target email address"`
		Cc            string `long:"cc" env:"CC" description:"the cc email address"`
		Subject       string `long:"subject" env:"SUBJECT" description:"the subject of email"`
		Text          string `long:"text" env:"TEXT" description:"the text of email not more 255 letters"`
		MailgunAPIURL string `long:"mailgunApiUrl" env:"MAILGUN_API_URL" description:"the mailgun API URL for sending notification"`
		MailgunAPIKey string `long:"mailgunApiKey" env:"MAILGUN_API_KEY" description:"the token for mailgun api"`
	} `group:"email" namespace:"email" env-namespace:"EMAIL"`

	Telegram struct {
		Enabled   bool   `long:"enabled" env:"ENABLED" description:"enable telegram provider"`
		BotAPIKey string `long:"botApiKey" env:"BOT_API_KEY" required:"true" description:"the telegram bot api key"`
		ChannelId string `long:"channelId" env:"CHANNEL_ID" description:"the channel id without leading symbol @"`
		Message   string `long:"message" env:"MESSAGE" description:"the text message not more 255 letters"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
}

var version = "unknown"

func main() {
	setupLogLevel(false)
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	log.Printf("[INFO] Starting Health checker for %s:%s ...\n", opts.URL, version)
	var client = &http.Client{}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"from":    strings.NewReader(opts.Email.From),
		"to":      strings.NewReader(opts.Email.To),
		"cc":      strings.NewReader(opts.Email.Cc),
		"subject": strings.NewReader(opts.Email.Subject),
		"text":    strings.NewReader(opts.Email.Text),
	}
	for range time.Tick(time.Second * opts.Timeout) {
		response, err := http.Get(opts.URL)
		if err != nil || response.StatusCode != 200 {
			if opts.Email.Enabled {
				err := SendEmail(client, opts.Email.MailgunAPIURL, values, opts.Email.MailgunAPIKey)
				if err != nil {
					log.Printf("[ERROR] error occurs during sending email: %+v", err)
				}
			}
			if opts.Telegram.Enabled {
				err := SendTelegramMessage(client, opts.Telegram.BotAPIKey, opts.Telegram.ChannelId, opts.Telegram.Message)
				if err != nil {
					log.Printf("[ERROR] error occurs during sending telegram message: %+v", err)
				}
			}
		}
	}
}

//SendTelegramMessage sending text message into public telegram channel
func SendTelegramMessage(client *http.Client, botAPIKey, channelId, message string) error {
	urlPattern := "https://api.telegram.org/bot%s/sendMessage?chat_id=@%s&text=%s"
	req, err := http.NewRequest("GET", fmt.Sprintf(urlPattern, botAPIKey, channelId, message), nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		log.Printf("[ERROR] Telegram response bad status: %s\n", res.Status)
		log.Printf("[ERROR] Telegram response bad body: %s\n", string(body))
	}
	return nil
}

//SendEmail sending email via MailGun
func SendEmail(client *http.Client, url string, values map[string]io.Reader, apiKey string) (err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
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
	credentials := strings.Split(apiKey, ":")
	if len(credentials) != 2 {
		return fmt.Errorf("[ERROR]: MAILGUN_API_KEY is not valid")
	}
	req.SetBasicAuth(credentials[0], credentials[1])
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		log.Printf("[ERROR] MailGun response bad status: %s\n", res.Status)
		log.Printf("[ERROR] MailGun response bad body: %s\n", string(body))
	}

	return err
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
