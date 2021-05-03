// otear -> scan in Spanish
// scans the stream of messages from ssh-chat
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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

type Options struct {
	ServerAddr       string   `short:"s" long:"server-addr" description:"ssh-chat server address to connect to" required:"true"`
	BotName          string   `short:"n" long:"name" description:"name of the bot listener bot" required:"true"`
	Keywords         []string `short:"k" long:"keywords" description:"keywords to monitor" required:"true"`
	PushoverToken    string   `short:"p" long:"pushover-token" description:"token used to send notifications to pushover" required:"true"`
	PushoverGroupKey string   `short:"g" long:"pushover-group" description:"pushover group to notify" required:"true"`

	Verbose         bool   `short:"v" long:"verbose" description:"print verbose messages"`
	LogTimeLocation string `short:"z" long:"log-tz" description:"timezone for log messages" default:"America/Vancouver"`
}

func run() error {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		return nil
	}
	err = setupLogger(opts)
	if err != nil {
		return err
	}

	for {
		client, err := sshclient.CreateClient(opts.ServerAddr, opts.BotName)
		if err != nil {
			return errors.Wrapf(err, "connect failed")
		}
		defer client.Close()

		readSomething := false

		err = handle(opts, func() (string, error) {
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

func handle(opts Options, readLine func() (string, error)) error {
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

		found := checkKeyword(msg, opts.Keywords)

		if !found {
			continue
		}
		lg.WithFields(lg.Fields{"from": from, "message": msg}).
			Info("Notifying for message from")
		sendPushoverNotification(opts, from, msg)
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

func sendPushoverNotification(opts Options, from, message string) {
	params := url.Values{}
	params.Add("token", opts.PushoverToken)
	params.Add("user", opts.PushoverGroupKey)
	params.Add("title", "SSH Chat Mention")
	params.Add("message", fmt.Sprintf("%s said: %s", from, message))

	resp, err := http.PostForm(pushoverMessageURL, params)
	if err != nil {
		lg.Warn("Unable to publish to pushover", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		lg.WithField("status", resp.Request.Response.Status).
			Warn("Unexpected status code from pushover", resp)
		return
	}
}
func setupLogger(opts Options) error {
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
