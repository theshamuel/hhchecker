package main

import (
	"bytes"
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

type Opts struct {
	URL           string `long:"url" env:"URL" required:"true" description:"url" default:"localhost"`
	From          string `long:"from" env:"EMAIL_FROM" required:"true" description:"the source email address"`
	To            string `long:"to" env:"EMAIL_TO" required:"true" description:"the target email address"`
	Cc            string `long:"cc" env:"EMAIL_CC" required:"true" description:"the cc email address"`
	Subject       string `long:"subject" env:"EMAIL_SUBJECT" required:"true" description:"the subject of email"`
	Text          string `long:"text" env:"EMAIL_TEXT" required:"true" description:"the text of email not more 255 letters"`
	TargetURL     string `long:"targetUrl" env:"TARGET_URL" required:"true" description:"the URL what you need to healthcheck"`
	MailgunAPIURL string `long:"mailgunApiUrl" env:"MAILGUN_API_URL" required:"true" description:"the mailgun API URL for sending notification"`
	BasicUser     string `long:"basicUser" env:"BASIC_USER" required:"true" description:"the user for mailgun api"`
	BasicPassword string `long:"basicPassword" env:"BASIC_PASSWORD" required:"true" description:"the password of user for mailgun api"`
}

var version = "unknown"

func main() {
	setupLogLevel(false)
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	log.Printf("[INFO] Starting Health checker for %s:%s ...\n", opts.TargetURL, version)
	var client = &http.Client{}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"from":    strings.NewReader(opts.From),
		"to":      strings.NewReader(opts.To),
		"cc":      strings.NewReader(opts.Cc),
		"subject": strings.NewReader(opts.Subject),
		"text":    strings.NewReader(opts.Text),
	}
	for range time.Tick(time.Minute * 60) {
		response, err := http.Get(opts.TargetURL)
		if err != nil || response.StatusCode != 200 {
			err := SendEmail(client, opts.MailgunAPIURL, values, opts.BasicUser, opts.BasicPassword)
			if err != nil {
				log.Printf("[ERROR] %+v", err)
			}
		}
	}
}

//SendEmail sending email via MailGun
func SendEmail(client *http.Client, url string, values map[string]io.Reader, basicUser, basicPassword string) (err error) {
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
	req.SetBasicAuth(basicUser, basicPassword)
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
