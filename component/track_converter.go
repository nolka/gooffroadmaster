package component

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nolka/gogpslib"
	"github.com/nolka/gogpslib/writer"
	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

func NewTrackConverter(manager *ComponentManager, startdir string) *TrackConverter {
	c := &TrackConverter{}
	util.LoadConfig(c)
	c.Init(manager)
	c.RuntimeDir = startdir
	return c
}

type TrackConverter struct {
	Manager             *ComponentManager `json:"-"`
	Id                  int               `json:"-"`
	RuntimeDir          string            `json:"runtime_dir"`
	ConverterBinaryPath string            `json:"binary_path"`
}

func (t *TrackConverter) SetId(id int) {
	t.Id = id
}

func (t *TrackConverter) Init(manager *ComponentManager) {
	t.Manager = manager
}

func (t *TrackConverter) GetName() string {
	return "GPS Track Converter"
}

func (t *TrackConverter) PrepareData(d ...string) string {
	return fmt.Sprintf("%d|%s", t.Id, strings.Join(d, "|"))
}

func (t *TrackConverter) HandleMessage(update tgbotapi.Update) {
	message := update.Message
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
				var data string = t.PrepareData(doc.FileID, fmt)
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Сделать "+fmt, data))
			}

			markup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					buttons...,
				),
			)
			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			msg.ReplyToMessageID = message.MessageID
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = markup
			msg.Text = "Могу сконвертировать этот файл в один из следующих форматов:"
			t.Manager.Results <- msg
		}
	}
}

func (t *TrackConverter) HandleCallback(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		log.Printf("ERR Callback data empty\n")
		return
	}
	parts := strings.Split(update.CallbackQuery.Data, "|")
	cmd, fileID, destFormat := parts[0], parts[1], parts[2]

	log.Printf("%s, %s, %s", cmd, fileID, destFormat)
	file, err := t.Manager.Bot.GetFile(tgbotapi.FileConfig{fileID})
	if err != nil {
		log.Println("GET FILE ERR: " + err.Error())
	}

	srcFn := t.RuntimeDir + "/" + update.CallbackQuery.Message.ReplyToMessage.Document.FileName
	util.DownloadFile(file.Link(t.Manager.Bot.Token), srcFn)
	if err != nil {
		log.Printf("Error downloading file: %s\n", err)
	}

	newName := convertFile(srcFn, destFormat)
	log.Printf("Sending file: %s", newName)
	doc := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.ReplyToMessage.Chat.ID, newName)
	doc.ReplyToMessageID = update.CallbackQuery.Message.ReplyToMessage.MessageID
	t.Manager.Bot.Send(doc)

	os.Remove(newName)
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
	defer os.Remove(srcFile)

	d.Write([]byte(w.Content))

	return destFileName
}
