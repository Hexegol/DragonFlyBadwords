package main

import (
	"fmt"
	"os"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := readConfig(log)
	if err != nil {
		log.Fatalln(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()
	badwords := []string{"ta gueule", "fils de pute", "salope"}
	srv.Listen()
	for srv.Accept(func(p *player.Player) { p.Handle(NewMyHandler(p, badwords)) }) {
	}
}

func Contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

type MyHandler struct {
	player.NopHandler
	p        *player.Player
	badwords []string
}

func NewMyHandler(p *player.Player, badwords []string) *MyHandler {
	return &MyHandler{p: p, badwords: badwords}
}

func (m *MyHandler) HandleChat(ctx *event.Context, message *string) {
	if Contains(m.badwords, *message) {
		ctx.Cancel()
		fmt.Println("Player", m.p.Name(), "tried to send:", *message)
		m.p.Message("you cant say '", *message, "'")
	}
}

func readConfig(log server.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
