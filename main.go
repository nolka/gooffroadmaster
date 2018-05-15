package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/nolka/gooffroadmaster/component"
	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	util.EnsureDirectories()

	config := getConfig()

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.IsDebug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	var results = make(chan tgbotapi.MessageConfig)
	manager := component.NewComponentManager(bot, results)
	manager.RegisterComponent(component.NewTrackConverter(manager, util.GetRuntimePath()))
	manager.RegisterComponent(component.NewInteractiveMenu(manager))

	subscribeInterrupt(manager)

	go resultsSender(results, bot)
	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] => %s\n", update.Message.From.UserName, update.Message.Text)
		}
		go func() {
			manager.Dispatch(update)
		}()
	}
}

func subscribeInterrupt(manager *component.ComponentManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("SIG %s", sig.String())
			manager.Halt()
			os.Exit(1)
		}
	}()
}

func getConfig() *Config {
	f, err := os.Open("config.json")
	if err != nil {
		log.Printf("CONF OPEN ERR: %s", err)
		return nil
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("CONF READ ERR: %s", err)
		return nil
	}
	cfg := &Config{}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		log.Printf("CONF PARSE ERR: %s", err)
	}

	cfg.WorkDir = util.GetStartupPath()
	cfg.RuntimeDir = cfg.WorkDir + string(os.PathSeparator) + "runtime"
	return cfg
}

// func instantiate(p reflect.Type) ExecutableCommand {
// 	instance := reflect.New(p).Elem()
// 	return  instance.Interface().(ExecutableCommand)
// }

func getCommand(cmd string, args []string) ExecutableCommand {
	for c, instance := range EnumerateCommands() {
		if cmd == c {
			// instance := instantiate(typeName).(ExecutableCommand)
			instance.SetArgs(args)
			return instance
		}
	}
	return nil
}

func parseCommand(command string) ExecutableCommand {
	if !strings.HasPrefix(command, "/") {
		return nil
	}
	parts := strings.Split(command[1:], " ")
	cmd := parts[0]
	parts = parts[1:]

	instance := getCommand(cmd, parts)
	if instance == nil {
		log.Printf("Failed to find command handler for '%s'", cmd)
		return nil
	}
	return instance
}

func resultsSender(message chan tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	for message := range message {
		bot.Send(message)
	}
}

func handleChannelMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) {
	message := update.Message

	go resultsSender(results, bot)
	go func() {
		if strings.HasPrefix(message.Text, "/") {
			cmd := parseCommand(message.Text)
			if cmd == nil {
				return
			}
			result, err := cmd.Handle(message, bot)
			if err != nil {
				log.Printf("ERROR: %s", err)
				return
			}
			results <- result
			return
		}
	}()
}

func handlePrivateMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) {
	message := update.Message
	log.Println(message.Chat.Title)
}
