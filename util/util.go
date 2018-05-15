package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func GetStartupPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return dir
}

func GetRuntimePath() string {
	return fmt.Sprintf("%s%sruntime", GetStartupPath(), string(os.PathSeparator))
}

func GetConfigPath() string {
	return fmt.Sprintf("%s%sconfig", GetStartupPath(), string(os.PathSeparator))
}

func EnsureDirectories() {
	dirs := []string{
		GetRuntimePath(),
		GetConfigPath(),
	}
	for _, d := range dirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			os.Mkdir(d, os.ModePerm)
		}
	}
}

func LoadConfig(i interface{}) {
	log.Printf("Loading config for %s", reflect.TypeOf(i).Elem().Name())
	file, err := os.Open(fmt.Sprintf("%s%s%s.json", GetConfigPath(), string(os.PathSeparator), reflect.TypeOf(i).Elem().Name()))
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	cfgbytes, _ := ioutil.ReadAll(file)

	err = json.Unmarshal(cfgbytes, &i)
	if err != nil {
		log.Printf("CFG ERR: %s\n", err)
		return
	}
}

func SaveConfig(i interface{}) {
	log.Printf("Saving config for %s", reflect.TypeOf(i).Elem().Name())
	file, err := os.Create(fmt.Sprintf("%s%s%s.json", GetConfigPath(), string(os.PathSeparator), reflect.TypeOf(i).Elem().Name()))
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	j, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Printf("CFG SAVE ERR: %s\n", err)
		return
	}

	file.Write(j)
}

func DownloadFile(url, dest string) (int64, error) {
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := io.Copy(out, response.Body)
	if err != nil {
		return 0, err
	}

	log.Printf("File downloaded success. Bytes read: %d\n", n)
	return n, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == os.ErrNotExist {
		log.Printf("File %s does not exists!", path)
		return false
	}
	return true
}

func MakePath(pathparts ...string) string {
	return strings.Join(pathparts, string(os.PathSeparator))
}
