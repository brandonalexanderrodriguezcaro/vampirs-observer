
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/namsral/flag"
)
const  defaultUrl= "https://www.google.com/?hl=es"
const defaultTick = 60 * time.Second

//const url = "http://:80/api/v1/label/__name__/values"

type config struct {
	contentType string
	server      string
	statusCode  int
	tick        time.Duration
	url         string
	userAgent   string
}

func (c *config) init(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.String(flag.DefaultConfigFlagname, "", "Path to config file")

	var (
		statusCode  = flags.Int("status", 200, "Response HTTP status code")
		tick        = flags.Duration("tick", defaultTick, "Ticking interval")
		server      = flags.String("server", "", "Server HTTP header value")
		contentType = flags.String("content_type", "", "Content-Type HTTP header value")
		userAgent   = flags.String("user_agent", "", "User-Agent HTTP header value")
		url         = flags.String("url", defaultUrl, "Request URL")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.statusCode = *statusCode
	c.tick = *tick
	c.server = *server
	c.contentType = *contentType
	c.userAgent = *userAgent
	c.url = *url

	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	c := &config{}

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					c.init(os.Args)
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()

	if err := run(ctx, c, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *config, stdout io.Writer) error {
	c.init(os.Args)
	log.SetOutput(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(c.tick):
			resp, err := http.Get(c.url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			log.Printf("Body:\n %s", body)

			if resp.StatusCode != c.statusCode {
				log.Printf("Status code mismatch, got: %d\n", resp.StatusCode)
			}

		}
	}
}