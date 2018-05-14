package component

import (
	"bytes"
	"fmt"
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

func NewTrackConverter(manager *ComponentManager, startdir string) *TrackConverter {
	c := &TrackConverter{}
	util.LoadConfig(c)
	c.Init(manager)
	c.RuntimeDir = startdir
	return c
}

type TrackConverter struct {
	Manager    *ComponentManager `json:"-"`
	Id         int               `json:"-"`
	RuntimeDir string            `json:"runtime_dir"`
	BinaryName string            `json:"binary_name"`
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

		if !util.FileExists(t.GetGpsbabelPath()) {
			log.Printf("Gpsbabel application is not found!")
			return
		}

		if !t.IsKnownFormat(path.Ext(doc.FileName)) {
			return
		}

		var buttons []tgbotapi.InlineKeyboardButton
		var known_formats = []string{
			".plt",
			".gpx",
			".kml",
			".kmz",
		}

		for _, format := range known_formats {
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
	fmtmap := t.GetKnownFormatsMap()
	keys := make([]string, 0, len(fmtmap))
	for k := range fmtmap {
		keys = append(keys, k)
	}

	_, ok := fmtmap[format]
	if ok {
		return true
	}
	return false
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

	newName, err, _, _ := t.convertUsingGpsBabel(srcFn, destFormat)
	if err != nil {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
		msg.Text = "Не удалось сконвертировать файл :("
		t.Manager.Results <- msg
	}
	log.Printf("Sending file: %s", newName)
	doc := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.ReplyToMessage.Chat.ID, newName)
	doc.ReplyToMessageID = update.CallbackQuery.Message.ReplyToMessage.MessageID
	t.Manager.Bot.Send(doc)

	os.Remove(newName)
}

func (t *TrackConverter) TrackToArguments(srcFile, dstFormat string) (string, string, string, string) {
	formatMap := t.GetKnownFormatsMap()

	srcFormat := strings.ToLower(formatMap[path.Ext(srcFile)])
	fileName := strings.TrimSuffix(path.Base(srcFile), "."+srcFormat)
	destFileName := fmt.Sprintf("%s%s%s", t.RuntimeDir, string(os.PathSeparator), fileName+dstFormat)
	return srcFormat, srcFile, formatMap[dstFormat], destFileName
}

func (t *TrackConverter) convertUsingGpsBabel(srcFile, dstFormat string) (string, error, string, string) {
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
		return "", err, stdout.String(), stderr.String()
	}

	log.Printf("Successfully converted. Output is:\n%s", stdout.String())
	return dstFileName, nil, stdout.String(), stderr.String()
}

func (t *TrackConverter) GetGpsbabelPath() string {
	return t.RuntimeDir + string(os.PathSeparator) + t.BinaryName
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
