// Copyright (c) 2023 Adam Prakash Stringer
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted (subject to the limitations in the disclaimer
// below) provided that the following conditions are met:
//
//     * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//
//     * Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer in the
//     documentation and/or other materials provided with the distribution.
//
//     * Neither the name of the copyright holder nor the names of its
//     contributors may be used to endorse or promote products derived from this
//     software without specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY
// THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR
// BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"
	"github.com/tauraamui/bluepanda/internal/logging"
	"github.com/tauraamui/bluepanda/internal/service"
	"github.com/tauraamui/bluepanda/pkg/client"
)

type args struct {
	Proto    string `arg:"--proto" default:"grpc"`
	LogLevel string `arg:"--loglevel" default:"info"`
	Port     int    `arg:"--port" default:"3000"`
}

func (args) Version() string {
	return "bluepanda v0.0.0"
}

type newServerFunc func(logging.Logger) (service.Server, error)

func run(log logging.Logger, newServer newServerFunc, opts args) {
	svr, err := newServer(log)
	if err != nil {
		log.Fatal().Msgf("error: %s", err)
	}
	log.Info().Msgf("%s starting %s service", opts.Version(), svr.Type())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	port := strconv.Itoa(opts.Port)
	addr := fmt.Sprintf(":%s", port)
	log.Info().Msgf("listening @ %s", addr)

	go func() {
		if err := svr.Listen(addr); err != nil {
			log.Fatal().Msgf("error: %s", err)
		}
	}()

	log.Info().Msg("bluepanda started, waiting for interrupt...")

	go func() {
		time.Sleep(3 * time.Second)
		client.Run()
	}()

	<-interrupt

	log.Info().Msg("shutting down gracefully...")
	if err := svr.Cleanup(log); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	if err := svr.ShutdownWithTimeout(60 * time.Second); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	log.Info().Msg("shut down... done")
}

func main() {
	var args args
	p := arg.MustParse(&args)

	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		p.Fail(fmt.Sprintf("unrecognised log level %s", args.LogLevel))
	}

	zerolog.SetGlobalLevel(logLevel)
	log := logging.New()

	proto := strings.ToLower(args.Proto)
	switch proto {
	case "http":
		run(log, service.NewHTTP, args)
	case "grpc":
		run(log, service.NewRPC, args)
	default:
		p.Fail(fmt.Sprintf("unrecognised protocol: %s", proto))
	}
}
