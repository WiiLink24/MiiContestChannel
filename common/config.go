package common

import (
	"encoding/xml"
	"log"
	"os"
)

type Config struct {
	Username        string `xml:"username"`
	Password        string `xml:"password"`
	DatabaseAddress string `xml:"databaseAddress"`
	DatabaseName    string `xml:"databaseName"`
	Address         string `xml:"address"`
	AssetsPath      string `xml:"assetsPath"`
	ServerURL	   	string `xml:"serverURL"`
}

func CheckError(err error) {
	if err != nil {
		log.Fatalf("Mii Contest Channel server has encountered a fatal error! Reason: %v\n", err)
	}
}

func GetConfig() Config {
	data, err := os.ReadFile("config.xml")
	CheckError(err)

	var config Config
	err = xml.Unmarshal(data, &config)

	return config
}
