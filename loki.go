package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"time"
	"y-clouds.com/tarantula/ampq"
	"y-clouds.com/tarantula/capture"
	"y-clouds.com/tarantula/oss"
)

var confFile = flag.String("c", "./conf.ini", "Snapshot tool configuration file.")

// AppConf is the config of app
type AppConf struct {
	AppMode      string
	ConsumeQueue string
	PublishQueue string
	AmpqConf     *ampq.Rabbit
	SeleniumConf *capture.Selenium
	OssConf      *oss.AliOss
}

var appConf = new(AppConf)

// setAppConf is used to set config params of the app
func setAppConf() {
	cfg, err := ini.Load(*confFile)

	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}
	appMode := cfg.Section("").Key("AppMode").String()
	appConf.AppMode = appMode

	// queue
	consumeQueue := cfg.Section("Queue").Key("Consume").String()
	appConf.ConsumeQueue = consumeQueue
	publishQueue := cfg.Section("Queue").Key("Publish").String()
	appConf.PublishQueue = publishQueue

	// rabbit conf
	ampqConf := new(ampq.Rabbit)
	err = cfg.Section("Rabbit").MapTo(ampqConf)
	if err != nil {
		log.Fatalf("Missing Rabbit MQ configuration parameters: %v", err)
	}
	appConf.AmpqConf = ampqConf

	// selenium conf
	seleniumConf := new(capture.Selenium)
	err = cfg.Section("Selenium").MapTo(seleniumConf)
	if err != nil {
		log.Fatalf("Missing selenium parameters: %v", err)
	}
	appConf.SeleniumConf = seleniumConf

	// oss conf
	aliOss := new(oss.AliOss)
	err = cfg.Section("OSS").MapTo(aliOss)
	if err != nil {
		log.Fatalf("Missing Ali oss configurationparameters: %v", err)
	}
	appConf.OssConf = aliOss
}

// GetEbayWebScreenshots is start to get tarantula
func getEbayWebScreenshots(asin string) (float32, []byte, string) {
	var seleniumConf = appConf.SeleniumConf
	ebay := &capture.Ebay{
		Asin:       asin,
		DriverPath: seleniumConf.DriverPath,
		Port:       seleniumConf.Port,
	}

	return ebay.WebScreenshots()
}

// uploadScreenshots Upload images to oss
func uploadScreenshots(imageName string, imageBytes []byte) {
	var aliOss = appConf.OssConf
	aliOss.PutBytesOnOSS(imageName, imageBytes)
	fmt.Printf("Upload %s to oss, success !\n", imageName)
}

// getScreenshotsName is used to generate filename of tarantula
func getScreenshotsName(param capture.ScreenshotsParam) string {
	timeStr := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s_%s_%s.png", param.Channel, param.Country, param.Asin, timeStr)
}

//
func publishScreenshotsResult(msg string, price float32, cutName string, status string) {
	response := capture.ScreenshotsResult{}
	err := json.Unmarshal([]byte(msg), &response)
	if err != nil {
		log.Fatalf("ampq message.format_error: %v", err)
	}
	response.Status = status
	response.NewPrice = price
	response.Screenshot = cutName

	rsJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Resonse json.serialize_error: %v", err)
	}
	publishAmpqConf := appConf.AmpqConf
	publishAmpqConf.Queue = appConf.PublishQueue
	ampq.Publish(publishAmpqConf, string(rsJson))
}

// consumeCallback is the RabbitMQ consumer callback function
// @param msg string
func consumeCallback(msg string) {
	fmt.Println(msg)
	param := capture.ScreenshotsParam{}
	err := json.Unmarshal([]byte(msg), &param)
	if err != nil {
		log.Fatalf("ampq message.format_error: %v", err)
	} //json解析到结构体里面

	// get []byte of tarantula
	price, imageBytes, status := getEbayWebScreenshots(param.Asin)
	imageName := ""
	if len(imageBytes) > 0 {
		// upload tarantula
		imageName = getScreenshotsName(param)
		var aliOss = appConf.OssConf
		if !aliOss.PutBytesOnOSS(imageName, imageBytes) {
			status = string(capture.UPLOAD_TO_OSS_ERROR)
		}
	}

	// Publish tarantula result to RabbitMQ tarantula.callback
	publishScreenshotsResult(msg, price, imageName, status)
}

func main() {
	flag.Parse()
	// set config of app
	setAppConf()

	// set consume
	consumeAmpqConf := appConf.AmpqConf
	consumeAmpqConf.Queue = appConf.ConsumeQueue
	ampq.Consume(consumeAmpqConf, consumeCallback)
}
