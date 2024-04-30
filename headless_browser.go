package main

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/projectdiscovery/gologger"
)

func OpenBrowser() (*rod.Browser, *rod.Page) {
	// Headless runs the browser on foreground, you can also use flag "-rod=show"
	// Devtools opens the tab in each new tab opened automatically
	l := launcher.New().
		Headless(false).
		Devtools(false).
		NoSandbox(true).
		Set("disable-gpu", "true").
		Set("ignore-certificate-errors", "true").
		Set("ignore-certificate-errors", "1").
		Set("disable-crash-reporter", "true").
		Set("disable-notifications", "true").
		Set("hide-scrollbars", "false").
		// Set("window-size", fmt.Sprintf("%d,%d", 2000, 1920)).
		Set("mute-audio", "true").
		Set("incognito", "true").
		Delete("use-mock-keychain")

	// defer l.Cleanup() // remove launcher.FlagUserDataDir

	u := l.MustLaunch()

	// Trace shows verbose debug information for each action executed
	// SlowMotion is a debug related function that waits 2 seconds between
	// each action, making it easier to inspect what your code is doing.
	browser := rod.New().
		ControlURL(u).
		MustConnect()
	page := browser.MustPage()

	// ServeMonitor plays screenshots of each tab. This feature is extremely
	// useful when debugging with headless mode.
	// You can also enable it with flag "-rod=monitor"
	// launcher.Open(browser.ServeMonitor(""))

	// defer browser.MustClose()
	gologger.Info().Msg("Broswer Opened Successfully !")
	return browser, page
}

func CloseBrowser(browser *rod.Browser) {
	browser.MustClose()
	gologger.Info().Msg("Broswer Closed Successfully !")
}

func Brows(myurl string, browser *rod.Browser, page *rod.Page) *rod.Page {

	page.MustSetExtraHeaders("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	// Eval js on the page

	// Evaluates given script in every frame upon creation
	// Disable all alerts by making window.alert no-op.
	page.MustEvalOnNewDocument(`window.alert = () => {}`)
	page.MustEvalOnNewDocument(`window.prompt = () => {}`)
	page.MustEvalOnNewDocument(`window.confirm = () => {}`)

	err := page.Timeout(time.Duration(10) * time.Second).Navigate(myurl)
	Delay(time.Duration(10))

	if err != nil && timeout_errors_counts < *max_timeout_counts {
		gologger.Warning().Msg("Error in loading Behkad website ! ")
		gologger.Info().Msg("Trying again ...")
		Delay(time.Duration(15))
		timeout_errors_counts += 1
		Brows(myurl, browser, page)
	} else if err != nil {
		gologger.Fatal().Msg("max retry exceed ! exiting...")
		fmt.Println(err.Error())
	}
	// Eval js on the page
	page.Timeout(time.Duration(10) * time.Second)
	Delay(time.Duration(10))

	return page
}

func Delay(Sec time.Duration) {
	time.Sleep(Sec * time.Second)
}
