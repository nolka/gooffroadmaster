package main

type Config struct {
	IsDebug    bool
	Token      string
	WorkDir    string
	RuntimeDir string
}

type ConversionInfo struct {
	Id        string
	ConvertTo string
}
