package main

import (
	"cdr/config"
	"cdr/processor"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		configPath = flag.String("config_path", "",
			"Config Folder path")
	//producerchan = make(chan objects.MailData, 100)
	)
	flag.Parse()
	//var configuration object.Configuration

	jsonFile, err := os.Open("config/charges.json")
	if err != nil {
		log.Fatal(err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}
	var result map[string][]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		log.Fatal(err)
	}
	charges := result["charge_table"]
	jsonFile.Close()
	configuration, err := config.New(*configPath)
	configuration.ChargeTable = make([]string, 0)
	for _, v := range charges {
		configuration.ChargeTable = append(configuration.ChargeTable, v.(string))
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatalf("Error while reading configuration")
	}
	processor.Config = configuration
	errChan := make(chan error)
	c := make(chan os.Signal, 1)
	//log.FileLogging(configuration.LogFilePath)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		receivedSignal := <-c
		logrus.Info("Recieved Signal to Stop  Application")
		errChan <- fmt.Errorf("%s", receivedSignal)
	}()
	wg := &sync.WaitGroup{}

	for _, v := range configuration.File {
		for _, val := range v.Queue {
			switch val.ProcessingType {
			case "flatfile":
				r := []rune(val.Delimeter)
				wg.Add(1)
				go processor.FlatfileProcessor(val.FileExtension, val.SuffixExtension, configuration.InputFilePath, r[0], wg, errChan)
			case "json":
				wg.Add(1)
				go processor.JsonProcessor(val.FileExtension, val.SuffixExtension, configuration.InputFilePath, wg, errChan)
			case "xml":
				wg.Add(1)
				go processor.XmlProcessor(val.FileExtension, val.SuffixExtension, configuration.InputFilePath, wg, errChan)
			}

		}
	}
	wg.Wait()

}
