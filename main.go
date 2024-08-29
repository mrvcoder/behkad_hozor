package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

type TelegramConfig struct {
	UseBot         bool   `json:"useBot"`
	ApiKey         string `json:"ApiKey"`
	TelegramUserId int64  `json:"telegramUserId"`
}

type BehkadConfig struct {
	CodeMelli   string  `json:"codeMelli"`
	CodeCollege string  `json:"codeCollege"`
	Address     string  `json:"address"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type Config struct {
	Telegram            TelegramConfig `json:"telegram"`
	Behkad              BehkadConfig   `json:"behkad"`
	Gozareshat          []string       `json:"gozareshat"`
	LastIndex           int            `json:"lastIndex"`
	MaxTimeoutCounts    int            `json:"maxTimeoutCounts"`
	TimeoutErrorsCounts int            `json:"timeoutErrorsCounts"`
}

var config *Config

// LoadConfig load 'conf.json' file data
func LoadConfig() error {
	file, err := os.Open("conf.json")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	config = &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

func main() {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)

	// load config data
	if err := LoadConfig(); err != nil {
		gologger.Fatal().Msg("Error loading config: " + err.Error())
	}

	url := "https://behkad.tvu.ac.ir/login.php"

	// at least 10 reports required
	if len(config.Gozareshat) < 10 {
		gologger.Fatal().Msg("Please write at least 10 gozaresh !")
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// Generate a random index different from the last used one
	randomIndex := getRandomIndex(len(config.Gozareshat), config.LastIndex)
	// Retrieve the random report
	randomReport := config.Gozareshat[randomIndex]
	// Save the last used index
	err := updateLastIndex(randomIndex)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	browser, page := OpenBrowser()
	page = Brows(url, browser, page)

	proto.BrowserGrantPermissions{
		Permissions: []proto.BrowserPermissionType{
			proto.BrowserPermissionTypeSensors,
			proto.BrowserPermissionTypeGeolocation,
			proto.BrowserPermissionTypeAccessibilityEvents},
	}.Call(page)

	// overwrite gps
	proto.EmulationSetGeolocationOverride{
		Latitude:  &config.Behkad.Latitude,
		Longitude: &config.Behkad.Longitude,
	}.Call(page)

	w := page.MustWaitRequestIdle()
	gologger.Info().Msg("Logging in behkad ... ")
	// login
	_ = page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
		document.querySelector("#user").value="%s"
		document.querySelector("#pass").value="%s"
		document.querySelector("#hero > div > div > div.col-md-5 > form > div:nth-child(11) > button.btn-get-started.login-user").click()
		window.confirm=function(){ return true; };
		window.alert=function(){ return true; };
		window.prompt=function(){ return "textOfMyChoice"; };
		}`, config.Behkad.CodeCollege, config.Behkad.CodeMelli)).ByUser())
	w()

	// hijack the popups
	_ = page.MustEvaluate(rod.Eval(`() => {
		window.confirm=function(){ return true; };
		window.alert=function(){ return true; };
		window.prompt=function(){ return "textOfMyChoice"; };
	}`).ByUser())
	gologger.Info().Msg("Logged in behkad ... !")
	page.Timeout(time.Duration(10) * time.Second)
	Delay(time.Duration(10))
	// hozor
	w = page.MustWaitRequestIdle()
	_ = page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
		document.querySelector("#myModal4").click()
	}`)).ByUser())
	_ = page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
		document.querySelector("#adress").value="%s"
		document.querySelector("#lat").value="%f"
		document.querySelector("#len").value="%f"
		document.querySelector("#main-wrapper > div > div.container-fluid > div:nth-child(1) > div > div > form > div > div > div > fieldset > div > div:nth-child(1) > button").click()
		document.querySelector("#myModal4 > div > div > div.modal-footer > div:nth-child(1) > button").click()
	}`, config.Behkad.Address, config.Behkad.Latitude, config.Behkad.Longitude)).ByUser())
	w()

	page.Timeout(time.Duration(10) * time.Second)
	Delay(time.Duration(10))
	// check hozor
	d := page.MustEvaluate(rod.Eval(`() => {
		if(document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("موفقیت")>0 || document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("ثبت شد")>0){
			return true
		}else{
			return false
		}
	}`).ByUser())
	statusHozor := strings.TrimSpace(d.Value.String())
	if statusHozor == "true" {
		gologger.Info().Msg("hozor Done ... !")
		page.Timeout(time.Duration(3) * time.Second)
		Delay(time.Duration(3))

		// close popup hozor
		page.MustEvaluate(rod.Eval(`() => {
			document.querySelector("#myModal4 > div > div > div.modal-header > button").click()
		}`).ByUser())

		gologger.Info().Msg("Submitting Gozaresh ... !")
		// sabt gozaresh
		page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
			document.querySelector("#res-hzor > div:nth-child(1) > div > div:nth-child(1) > div.col-md-4.btn-action-user > button").click()
			document.querySelector("#myModal1 > div > div > div.modal-body.h4.text-dark > form > textarea").value="%s"
		}`, randomReport)).ByUser())
		w = page.MustWaitRequestIdle()
		page.MustEvaluate(rod.Eval(`() => {
			document.querySelector("#myModal1 > div > div > div.modal-body.h4.text-dark > form > div.modal-footer > button").click()
		}`).ByUser())
		w()
		page.Timeout(time.Duration(10) * time.Second)
		Delay(time.Duration(10))
		// check gozaresh
		d := page.MustEvaluate(rod.Eval(`() => {
			if(document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("ثبت شد")>0 || document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("موفقیت")>0){
				return true
			}else{
				return false
			}
		}`).ByUser())

		statusGozaresh := strings.TrimSpace(d.Value.String())
		if statusGozaresh == "true" {
			gologger.Info().Msg("Gozaresh Done ... !")
			sendMessageToTelegramBot(config.Telegram.TelegramUserId, "حضور برای کاربر با کد ملی "+config.Behkad.CodeMelli+" ثبت شد")
		} else {
			gologger.Error().Msg("Error in gozaresh")
			sendMessageToTelegramBot(config.Telegram.TelegramUserId, "گزارش شما ثبت نشد. لطفا بررسی نمایید")
			exitApp()
		}
	} else {
		gologger.Error().Msg("Error in hozor")
		sendMessageToTelegramBot(config.Telegram.TelegramUserId, "حضور شما ثبت نشد. لطفا بررسی نمایید")
		exitApp()
	}
}

func exitApp() {
	gologger.Info().Msg("Exiting ... ")
	os.Exit(0)
}

// Function to update the last used index
func updateLastIndex(newIndex int) error {
	config.LastIndex = newIndex

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("conf.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Function to generate a random index different from the last used one
func getRandomIndex(maxLength, lastIndex int) int {

	randomIndex := rand.Intn(maxLength)
	for randomIndex == lastIndex {
		if lastIndex+1 <= maxLength {
			return lastIndex + 1
		} else {
			return lastIndex - 1
		}
	}
	return randomIndex
}

func sendMessageToTelegramBot(telegramUserId int64, message string) {
	if config.Telegram.UseBot {
		gologger.Info().Msg("Sending message to telegram bot ... !")
		bot, err := tgbotapi.NewBotAPI(config.Telegram.ApiKey)
		if err != nil {
			gologger.Error().Msg("Error sending message in telegram: " + err.Error())
			exitApp()
		}

		bot.Debug = false

		// send message to user
		msg := tgbotapi.NewMessage(telegramUserId, message)
		_, err = bot.Send(msg)
		if err != nil {
			gologger.Error().Msg("Error sending message to telegram bot")
		}
		gologger.Info().Msg("Message sent to telegram bot ... !")
	}
}
