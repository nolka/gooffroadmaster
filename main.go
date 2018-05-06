package main

import (
	"encoding/json"
	"github.com/nolka/gogpslib"
	"github.com/nolka/gogpslib/writer"
	"gopkg.in/telegram-bot-api.v4"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	config := getConfig()

	log.Printf("Runtime dir: %s\n", config.RuntimeDir)

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	var results = make(chan tgbotapi.MessageConfig)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallback(&update, bot, results, config)
		}
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		switch update.Message.Chat.Type {
		case "private":
			handlePrivateMessage(&update, bot, results)
		default:
			handleChannelMessage(&update, bot, results)
		}

	}
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

	cfg.WorkDir = getStartupPath()
	cfg.RuntimeDir = cfg.WorkDir + string(os.PathSeparator) + "runtime";
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

func handleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig, cfg *Config) {
	if update.CallbackQuery == nil {
		log.Printf("ERR Callback data empty\n")
		return
	}
	parts := strings.Split(update.CallbackQuery.Data, "|")
	cmd, fileId, destFormat := parts[0], parts[1], parts[2]

	log.Printf("%s, %s, %s", cmd, fileId, destFormat)
	file, err := bot.GetFile(tgbotapi.FileConfig{fileId})
	if err != nil {
		log.Println("GET FILE ERR: " + err.Error())
	}

	response, err := http.Get(file.Link(cfg.Token))
	if err != nil {
		log.Println("DL FILE ERR: " + err.Error())
	}
	defer response.Body.Close()

	var srcFn string = cfg.RuntimeDir + "/" + update.CallbackQuery.Message.ReplyToMessage.Document.FileName
	out, err := os.Create(srcFn)
	if err != nil {
		log.Println("MK FILE ERR: " + err.Error())
	}
	defer out.Close()

	n, err := io.Copy(out, response.Body)
	if err != nil {
		log.Println("COPY FILE ERR: " + err.Error())
	}

	log.Printf("File downloaded success. Bytes read: %d", n)

	newName := convertFile(srcFn, destFormat)
	log.Printf("Sending file: %s", newName)
	doc := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.ReplyToMessage.Chat.ID, newName)
	doc.ReplyToMessageID = update.CallbackQuery.Message.ReplyToMessage.MessageID
	bot.Send(doc)
}

func convertFile(srcFile string, dstFormat string) string {
	converters := map[string]gogpslib.FormatReaderWriter{
		".plt": &gogpslib.PltFormat{},
		".gpx": &gogpslib.GpxFormat{},
	}

	ext := filepath.Ext(srcFile)
	basename := filepath.Base(srcFile)
	newName := strings.TrimSuffix(basename, filepath.Ext(basename)) + dstFormat

	src := converters[ext]
	src.Read(srcFile)

	dst := converters[dstFormat]
	dst.SetSegments(src.GetSegments())

	w := writer.CreateStringWriter()
	dst.WriteSegments(w)
	w.Write()

	abs, _ := filepath.Abs(srcFile)
	destFileName := filepath.Dir(abs) + string(os.PathSeparator) + newName

	d, err := os.Create(destFileName)
	if err != nil {
		log.Printf("Error creating dest file name: %s", destFileName)
	}
	defer d.Close()

	d.Write([]byte(w.Content))

	return destFileName
}

func handleChannelMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) {
	message := update.Message

	go resultsSender(results, bot)
	go func() {
		if strings.HasPrefix(message.Text, "/") {
			cmd := parseCommand(message.Text)
			if cmd == nil {
				return;
			}
			result, err := cmd.Handle(message, bot)
			if err != nil {
				log.Printf("ERROR: %s", err)
				return
			}
			results <- result
			return
		}

		if message.Document != nil {
			doc := message.Document

			if strings.HasSuffix(doc.FileName, ".gpx") || strings.HasSuffix(doc.FileName, ".plt") {

				var buttons []tgbotapi.InlineKeyboardButton
				var known_formats = []string{
					".plt",
					".gpx",
				}

				for _, fmt := range known_formats {
					if strings.HasSuffix(doc.FileName, fmt) {
						continue
					}
					var data string = "convert|" + doc.FileID + "|" + fmt
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("конвертнуть в "+fmt, data))
				}

				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Дать в пердак", "give_anal"))

				markup := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						buttons...
					),
				)
				msg := tgbotapi.NewMessage(message.Chat.ID, "")
				msg.ReplyToMessageID = message.MessageID
				msg.ParseMode = "HTML"
				msg.ReplyMarkup = markup
				msg.Text = "Wanna some conversions?"
				results <- msg
			}
		}
	}()
}

func handlePrivateMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) {
	message := update.Message
	log.Println(message.Chat.Title)
}
