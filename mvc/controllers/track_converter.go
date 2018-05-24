package controllers

import (
	"bytes"
	"fmt"
	"github.com/nolka/gooffroadmaster/mvc"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/nolka/gogpslib"
	"github.com/nolka/gogpslib/writer"
	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	defaultConverterID = 2
)

type conversionCallback func(srcFile string, destFormat string) (string, error)

func NewTrackConverter(manager *mvc.Router, runtimeDir string) *TrackConverter {
	c := &TrackConverter{}
	util.LoadConfig(c)
	c.Init(manager)
	if c.RuntimeDir == "" {
		c.RuntimeDir = runtimeDir
	}
	if c.ConverterId == 0 {
		c.ConverterId = defaultConverterID
	}
	return c
}

type TrackConverter struct {
	Manager     *mvc.Router `json:"-"`
	Id          int     `json:"-"`
	RuntimeDir  string  `json:"runtime_dir"`
	BinaryName  string  `json:"binary_name"`
	ConverterId int     `json:"converter_id"`
}

func (t *TrackConverter) SetId(id int) {
	t.Id = id
}

func (t *TrackConverter) Init(manager *mvc.Router) {
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

		if !util.FileExists(t.GetGpsbabelPath()) {
			log.Printf("Gpsbabel application is not found!")
			return
		}

		if !t.IsKnownFormat(path.Ext(doc.FileName)) {
			return
		}

		var buttons []tgbotapi.InlineKeyboardButton

		for _, format := range t.GetKnownFormatsMap() {
			if strings.HasSuffix(doc.FileName, format) {
				continue
			}
			var data string = t.PrepareData(doc.FileID, format)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Сделать "+format, data))
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

func (t *TrackConverter) GetKnownFormatsMap() map[string]string {
	return map[string]string{
		".kml": "kml",
		".kmz": "kmz",
		".plt": "ozi",
		".gpx": "gpx",
	}
}

func (t *TrackConverter) IsKnownFormat(format string) bool {
	for ext, _ := range t.GetKnownFormatsMap() {
		if format == ext {
			return true
		}
	}
	return false
}

func (t *TrackConverter) HandleCallback(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		log.Println("ERR Callback data empty")
		return
	}
	parts := strings.Split(update.CallbackQuery.Data, "|")
	cmd, fileID, destFormat := parts[0], parts[1], parts[2]

	log.Printf("%s, %s, %s", cmd, fileID, destFormat)
	file, err := t.Manager.Bot.GetFile(tgbotapi.FileConfig{fileID})
	if err != nil {
		log.Println("GET FILE ERR: " + err.Error())
		return
	}

	srcFileName := t.RuntimeDir + "/" + update.CallbackQuery.Message.ReplyToMessage.Document.FileName
	util.DownloadFile(file.Link(t.Manager.Bot.Token), srcFileName)
	if err != nil {
		log.Printf("Error downloading file: %s\n", err)
		return
	}

	var newFileName string
	switch t.ConverterId {
	case 1:
		{
			newFileName, err = t.convert(srcFileName, destFormat, t.ConvertInternalFile)
			break
		}
	case 2:
	default:
		{
			newFileName, err = t.convert(srcFileName, destFormat, t.ConvertUsingGpsBabel)
			break
		}
	}

	if err != nil {
		log.Printf("Failed to convert file: %s\n", err)
		return
	}

	log.Printf("Sending file: %s", newFileName)
	doc := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.ReplyToMessage.Chat.ID, newFileName)
	doc.ReplyToMessageID = update.CallbackQuery.Message.ReplyToMessage.MessageID
	t.Manager.Bot.Send(doc)

	// Here we can make some tracks cache
	os.Remove(newFileName)
	os.Remove(srcFileName)
}

func (t *TrackConverter) convert(srcFile, destFormat string, converter conversionCallback) (string, error) {
	return converter(srcFile, destFormat)
}

func (t *TrackConverter) TrackToArguments(srcFile, dstFormat string) (string, string, string, string) {
	formatMap := t.GetKnownFormatsMap()

	srcFormat := strings.ToLower(formatMap[path.Ext(srcFile)])
	fileName := strings.TrimSuffix(path.Base(srcFile), "."+srcFormat)
	destFileName := fmt.Sprintf("%s%s%s", t.RuntimeDir, string(os.PathSeparator), fileName+dstFormat)
	return srcFormat, srcFile, formatMap[dstFormat], destFileName
}

func (t *TrackConverter) ConvertUsingGpsBabel(srcFile, dstFormat string) (string, error) {
	srcFormat, srcFile, dstFormat, dstFileName := t.TrackToArguments(srcFile, dstFormat)

	cmd := exec.Command(t.GetGpsbabelPath(), "-i", srcFormat, "-f", srcFile, "-o", dstFormat, "-F", dstFileName)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	cmd.Wait()
	if err != nil {
		log.Printf("CONVERT ERR: %s\n%s\n%s\n", err.Error(), stdout.String(), stderr.String())
		return "", err
	}

	log.Printf("Successfully converted. Output is:\n%s", stdout.String())
	return dstFileName, nil
}

func (t *TrackConverter) GetGpsbabelPath() string {
	return t.RuntimeDir + string(os.PathSeparator) + t.BinaryName
}

func (t *TrackConverter) ConvertInternalFile(srcFile, dstFormat string) (string, error) {
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
		return "", fmt.Errorf("Error creating dest file name: %s", destFileName)
	}
	defer d.Close()
	defer os.Remove(srcFile)

	d.Write([]byte(w.Content))

	return destFileName, nil
}
