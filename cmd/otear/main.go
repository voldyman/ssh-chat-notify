// otear -> scan in Spanish
// scans the stream of messages from ssh-chat
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	config "github.com/gookit/config/v2"
	jcfg "github.com/gookit/config/v2/json"
	flags "github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	lg "github.com/sirupsen/logrus"
	sshclient "github.com/voldyman/ssh-chat-notify/client"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

type MentionConfig struct {
	Name             string   `mapstructure:"name"`
	Keywords         []string `mapstructure:"keywords"`
	PushoverToken    string   `mapstructure:"pushover-token"`
	PushoverGroupKey string   `mapstructure:"pushover-group"`
}

type Config struct {
	ServerAddr  string          `mapstructure:"server-addr"`
	BotName     string          `mapstructure:"name"`
	MentionCfgs []MentionConfig `mapstructure:"mentions"`
}

type CliOptions struct {
	Cfg             string `short:"c" long:"config" description:"location of the config file" default:"config.json"`
	Verbose         bool   `short:"v" long:"verbose" description:"print verbose messages"`
	LogTimeLocation string `short:"z" long:"log-tz" description:"timezone for log messages" default:"America/Vancouver"`
}

func run() error {
	var opts CliOptions
	_, err := flags.Parse(&opts)
	if err != nil {
		return err
	}
	err = setupLogger(opts)
	if err != nil {
		return err
	}
	cfg, err := loadConfig(opts.Cfg)
	if err != nil {
		return err
	}
	for {
		client, err := sshclient.CreateClient(cfg.ServerAddr, cfg.BotName)
		if err != nil {
			return errors.Wrapf(err, "connect failed")
		}
		defer client.Close()

		readSomething := false

		err = handle(cfg.MentionCfgs, func() (string, error) {
			line, err := client.ScanLine()
			if err == nil {
				readSomething = true
			}
			return line, err
		})

		if err == nil {
			lg.Info("handle returned without error")
			continue
		}
		if !readSomething {
			return errors.Wrapf(err, "failed without reading from server")
		}
		lg.Warn("message handler failed, retrying:", err)
		time.Sleep(10 * time.Second)
	}
}

func loadConfig(file string) (*Config, error) {
	config.AddDriver(jcfg.Driver)
	err := config.LoadFiles(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load config file: %w", err)
	}
	var cfg Config
	err = config.BindStruct("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to bind config struct: %w", err)
	}

	return &cfg, nil
}

func handle(mentions []MentionConfig, readLine func() (string, error)) error {
	for {
		cline, err := readLine()
		if err != nil {
			return errors.Wrapf(err, "read failed")
		}
		line := strings.TrimSpace(cline)

		lg.WithField("line", line).Debug("Scanned line")

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "*") {
			lg.Debug("Ignoring system message")
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			lg.WithField("line", line).
				Warn("Message does not have a sender, ignoring")
			continue
		}
		from := strings.TrimSpace(parts[0])
		msg := strings.TrimSpace(parts[1])

		for _, mcfg := range mentions {

			if checkKeyword(msg, mcfg.Keywords) {
				lg.WithFields(lg.Fields{"from": from, "message": msg, "cfg": mcfg.Name}).
					Info("Notifying for message")
				sendPushoverNotification(from, msg, pushoverCfg{
					Token:    mcfg.PushoverToken,
					GroupKey: mcfg.PushoverGroupKey,
				})
			}
		}

	}
}

func checkKeyword(str string, keywords []string) bool {
	for _, word := range keywords {
		if strings.Contains(str, word) {
			lg.WithField("word", word).
				Info("Found keyword")

			return true
		}
	}
	return false
}

const pushoverMessageURL = "https://api.pushover.net/1/messages.json"

type pushoverCfg struct {
	Token    string
	GroupKey string
}

func sendPushoverNotification(from, message string, po pushoverCfg) {
	params := url.Values{}
	params.Add("token", po.Token)
	params.Add("user", po.GroupKey)
	params.Add("title", "SSH Chat Mention")
	params.Add("message", fmt.Sprintf("%s said: %s", from, message))

	resp, err := http.PostForm(pushoverMessageURL, params)
	if err != nil {
		lg.Warn("Unable to publish to pushover", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		lg.WithField("status", resp.StatusCode).
			Warn("Unexpected status code from pushover", resp)
		return
	}
}
func setupLogger(opts CliOptions) error {
	if opts.Verbose {
		lg.SetLevel(lg.DebugLevel)
	} else {
		lg.SetLevel(lg.InfoLevel)
	}
	timeZone, err := time.LoadLocation(opts.LogTimeLocation)
	if err != nil {
		return errors.Wrapf(err, "unable to setup logger for location %s", opts.LogTimeLocation)
	}

	lg.SetFormatter(newTzFormatter(timeZone))

	return nil
}

type TZLogFormatter struct {
	location *time.Location
	fmtr     *lg.TextFormatter
}

func newTzFormatter(loc *time.Location) lg.Formatter {
	return &TZLogFormatter{
		location: loc,
		fmtr: &lg.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC822,
		},
	}
}

func (tz *TZLogFormatter) Format(e *lg.Entry) ([]byte, error) {
	e.Time = e.Time.In(tz.location)
	return tz.fmtr.Format(e)
}
