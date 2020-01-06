// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/drone/signal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"

	"github.com/go-ini/ini"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"

	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/rpc"

	"github.com/coderun-top/coderun/src/utils"
	"github.com/coderun-top/coderun/src/version"
)

const (
	defaultClientMaxReceiveMessageSize = 1024 * 1024 * 100
)

// user model: agent for coderun for user to deploy heself
// dougo model: agent for dougo for all user
var agentType string = "user"

func main() {
	app := cli.NewApp()
	app.Name = "dougo-agent"
	app.Version = version.Version.String()
	app.Usage = "dougo agent"
	app.Action = loop
	app.Commands = []cli.Command{
		{
			Name:   "ping",
			Usage:  "ping the agent",
			Action: pinger,
		},
	}

	paramPrefix := "DOUGO"
	if agentType == "user" {
		paramPrefix = "CODERUN"
	}

	flags := []cli.Flag{
		cli.BoolFlag{
			EnvVar: fmt.Sprintf("%s_DEBUG", paramPrefix),
			Name:   "debug",
			Usage:  "enable agent debug mode",
		},
		cli.BoolFlag{
			EnvVar: fmt.Sprintf("%s_DEBUG_PRETTY", paramPrefix),
			Name:   "pretty",
			Usage:  "enable pretty-printed debug output",
		},
		cli.BoolTFlag{
			EnvVar: fmt.Sprintf("%s_DEBUG_NOCOLOR", paramPrefix),
			Name:   "nocolor",
			Usage:  "disable colored debug output",
		},
		cli.StringFlag{
			EnvVar: fmt.Sprintf("%s_PLATFORM", paramPrefix),
			Name:   "platform",
			Usage:  "restrict builds by platform conditions",
			Value:  "linux/amd64",
		},
		cli.DurationFlag{
			EnvVar: fmt.Sprintf("%s_KEEPALIVE_TIME", paramPrefix),
			Name:   "keepalive-time",
			Usage:  "after a duration of this time of no activity, the agent pings the server to check if the transport is still alive",
		},
		cli.DurationFlag{
			EnvVar: fmt.Sprintf("%s_KEEPALIVE_TIMEOUT", paramPrefix),
			Name:   "keepalive-timeout",
			Usage:  "after pinging for a keepalive check, the agent waits for a duration of this time before closing the connection if no activity",
			Value:  time.Second * 20,
		},
		cli.StringFlag{
			EnvVar: fmt.Sprintf("%s_SERVER", paramPrefix),
			Name:   "server",
			Usage:  "server address",
			Value:  "https://g.coderun.top/",
		},
		cli.StringFlag{
			EnvVar: fmt.Sprintf("%s_NAME", paramPrefix),
			Name:   "name",
			Usage:  "agent name",
		},
		cli.StringFlag{
			EnvVar: fmt.Sprintf("%s_TAGS", paramPrefix),
			Name:   "tags",
			Usage:  "coderun agent tags",
		},
		cli.StringFlag{
			EnvVar: fmt.Sprintf("%s_SECRET", paramPrefix),
			Name:   "secret",
			Usage:  "server-agent shared secret key",
		},
		cli.BoolTFlag{
			EnvVar: fmt.Sprintf("%s_HEALTHCHECK", paramPrefix),
			Name:   "healthcheck",
			Usage:  "enable healthcheck endpoint",
		},
	}

	if agentType == "dougo" {
		flags = append(flags, []cli.Flag{
			cli.IntFlag{
				EnvVar: fmt.Sprintf("%s_MAX_PROCS", paramPrefix),
				Name:   "max-procs",
				Usage:  "agent parallel builds",
				Value:  1,
			},
		}...)
	}
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loop(c *cli.Context) error {
	log.Info().Msg("Agent running in " + agentType + " model.")

	maxProcs := 1
	platform := c.String("platform")
	name := c.String("name")
	server := c.String("server")
	tags := c.String("tags")
	secret := c.String("secret")

	if server == "" {
		return fmt.Errorf("server does not null")
	}

	// format server
	if agentType == "user" {
		u, err := url.Parse(server)
		if err != nil {
			return fmt.Errorf("server parse error, %s", err)
		}

		server = u.Host
		if server == "g.coderun.top" {
			server += ":31617"
		} else if server == "dev.crun.top" {
			server += ":31607"
		}
	}

	// get agent id
	clientID := ""
	agentFile := "/etc/coderun-agent/config"
	if agentType == "dougo" {
		agentFile = "/etc/dougo-agent/config"
	}
	cfg, err := ini.Load(agentFile)
	if err != nil {
		log.Info().Msg("Connot find file: " + agentFile + ", " + err.Error())
		cfg = ini.Empty()
		clientID = utils.GeneratorId()
		cfg.Section("").Key("client_id").SetValue(clientID)
		cfg.SaveTo(agentFile)
	} else {
		clientID = cfg.Section("").Key("client_id").String()
	}

	log.Info().Msg("Client ID: " + clientID)

	// 获取ip地址
	addr := ""

	if len(name) == 0 {
		name, _ = os.Hostname()
	}

	if agentType == "dougo" {
		maxProcs = c.Int("max-procs")
	}

	filter := rpc.Filter{
		Labels: map[string]string{
			"type":      agentType,
			"client_id": clientID,
			"platform":  platform,
			"name":      name,
			"addr":      addr,
			"key":       secret,
			"tags":      tags,
			"max_procs": strconv.Itoa(maxProcs),
		},
	}

	if c.BoolT("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	if c.Bool("pretty") {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: c.BoolT("nocolor"),
			},
		)
	}

	counter.Polling = maxProcs
	counter.Running = 0

	if c.BoolT("healthcheck") {
		go http.ListenAndServe(":3000", nil)
	}

	// TODO pass version information to grpc server
	// TODO authenticate to grpc server

	// grpc.Dial(target, ))

	conn, err := grpc.Dial(
		server,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(defaultClientMaxReceiveMessageSize),
		),
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&credentials{
			username: "x-oauth-basic",
			password: "dougo1234abcd",
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.Duration("keepalive-time"),
			Timeout: c.Duration("keepalive-timeout"),
		}),
	)

	if err != nil {
		return err
	}
	defer conn.Close()

	client := rpc.NewGrpcClient(conn)

	sigterm := abool.New()
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("name", name),
	)
	ctx = signal.WithContextFunc(ctx, func() {
		println("ctrl+c received, terminating process")
		sigterm.Set()
	})

	var wg sync.WaitGroup
	parallel := maxProcs
	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for {
				if sigterm.IsSet() {
					return
				}
				r := runner{
					client: client,
					filter: filter,
					name:   name,
				}
				if err := r.run(ctx); err != nil {
					log.Error().Err(err).Msg("pipeline done with error")
					return
				}
			}
		}()
	}

	wg.Wait()
	return nil
}
