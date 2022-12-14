package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/robinbraemer/event"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/component/codec/legacy"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type CustomConfig struct {
	Custom struct {
		PlayerCount int
		KickMessage string
	}
}

type CustomProxy struct {
	*proxy.Proxy
	customCfg CustomConfig
}

var legacyCodec = &legacy.Legacy{Char: legacy.AmpersandChar}

func loadCustomConfig(cfg *CustomConfig) {
	data, err := os.ReadFile("config.toml")

	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(data, &cfg)

	if err != nil {
		panic(err)
	}
}

func main() {
	proxy.Plugins = append(proxy.Plugins, proxy.Plugin{
		Name: "CustomProxy",
		Init: func(ctx context.Context, proxy *proxy.Proxy) error {
			fmt.Printf("Listening on %s\n", proxy.Config().Bind)
			return newCustomProxy(proxy).init()
		},
	})

	os.Setenv("GATE_VERBOSITY", "-1")
	gate.Execute()
}

func newCustomProxy(proxy *proxy.Proxy) *CustomProxy {
	var cfg CustomConfig

	loadCustomConfig(&cfg)

	return &CustomProxy{
		Proxy:     proxy,
		customCfg: cfg,
	}
}

// Init proxy
func (p *CustomProxy) init() error {
	// Register events
	event.Subscribe(p.Event(), 0, p.onPlayerLogin)
	event.Subscribe(p.Event(), 0, onServerPing(p.Config().Status.Motd, p.customCfg.Custom.PlayerCount))

	return nil
}

func (p *CustomProxy) onPlayerLogin(e *proxy.PostLoginEvent) {
	message, err := legacyCodec.Unmarshal([]byte(p.customCfg.Custom.KickMessage))

	if err != nil {
		panic("Error parsing kick message")
	}

	time := time.Now()

	fmt.Printf(
		"[%s]: %s (%s) tried to connect!\n",
		time.Format("02.01.2006 15:04"),
		e.Player().GameProfile().Name,
		e.Player().GameProfile().ID.String(),
	)

	e.Player().Disconnect(message)
}

func onServerPing(motd string, playerCount int) func(e *proxy.PingEvent) {
	message, err := legacyCodec.Unmarshal([]byte(motd))

	if err != nil {
		panic("Error parsing motd")
	}

	return func(e *proxy.PingEvent) {
		p := e.Ping()
		p.Description = message.(*component.Text)
		p.Players.Max = playerCount
		p.Players.Online = playerCount
	}
}
