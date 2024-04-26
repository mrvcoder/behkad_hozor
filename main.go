package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jszwec/csvutil"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

var (
	user_latitude         = flag.Float64("latitude", 0, "Latitude")
	user_longitude        = flag.Float64("longitude", 0, "longitude")
	user_code_meli        = flag.String("code_meli", "", "code meli user")
	user_code_collage     = flag.String("code_collage", "", "code collage user")
	user_address          = flag.String("address", "", "hozor address")
	max_timeout_counts    = flag.Int("max_timeout_errors", 2, "Max timeout errors count to retry")
	timeout_errors_counts = 0
	lastIndexFile         = "last_index.txt"
)

type Reports struct {
	Dsc string `csv:"dsc"`
}

func main() {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)
	// check required flags
	flag.Parse()
	required := []string{"latitude", "longitude", "code_meli", "address"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			// or possibly use `log.Fatalf` instead of:
			gologger.Fatal().Msg(fmt.Sprintf("missing required -%s argument/flag\n", req))
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}

	url := "https://behkad.tvu.ac.ir/login.php"

	// read & parse csv
	currentContent, err := ioutil.ReadFile("./gozareshat.csv")
	if err != nil {
		gologger.Fatal().Msg("Error in Reading gozareshat.csv file : " + err.Error())
	}
	var reports []Reports
	if err := csvutil.Unmarshal(currentContent, &reports); err != nil {
		gologger.Fatal().Msg("Error in Parsing gozareshat.csv file : " + err.Error())
	}
	if len(reports) < 10 {
		gologger.Fatal().Msg("Please write at least 10 gozaresh !")
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// Get the last used index
	lastIndex := getLastIndex()
	// Generate a random index different from the last used one
	randomIndex := getRandomIndex(len(reports), lastIndex)
	// Retrieve the random report
	randomReport := reports[randomIndex]
	// Save the last used index
	err = saveLastIndex(randomIndex)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	browser, page := OpenBrowser()
	page = Brows(url, browser, page)
	w := page.MustWaitRequestIdle()
	gologger.Info().Msg("Loggin in behkad ... ")
	// login
	_ = page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
		document.querySelector("#user").value="%s"
		document.querySelector("#pass").value="%s"
		document.querySelector("#hero > div > div > div.col-md-5 > form > div:nth-child(11) > button.btn-get-started.login-user").click()
	}`, *user_code_collage, *user_code_meli)).ByUser())
	w()
	gologger.Info().Msg("Logged in behkad ... !")

	proto.BrowserGrantPermissions{
		Permissions: []proto.BrowserPermissionType{
			proto.BrowserPermissionTypeSensors,
			proto.BrowserPermissionTypeGeolocation,
			proto.BrowserPermissionTypeAccessibilityEvents},
	}.Call(page)

	// over write gps
	proto.EmulationSetGeolocationOverride{
		Latitude:  user_latitude,
		Longitude: user_longitude,
	}.Call(page)

	// hozor
	w = page.MustWaitRequestIdle()
	_ = page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
		document.querySelector("#adress").value="%s"
		document.querySelector("#lat").value="%f"
		document.querySelector("#len").value="%f"
		document.querySelector("#main-wrapper > div > div.container-fluid > div:nth-child(1) > div > div > form > div > div > div > fieldset > div > div:nth-child(1) > button").click()
		document.querySelector("#myModal4 > div > div > div.modal-footer > div:nth-child(1) > button").click()
	}`, *user_address, *user_latitude, *user_longitude)).ByUser())
	w()

	// check hozor
	d := page.MustEvaluate(rod.Eval(`() => {
		if(document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("موفقیت")>0){
			return true
		}else{
			return false
		}
	}`).ByUser())
	out := strings.TrimSpace(d.Value.String())
	status_hozor := string(out)
	if status_hozor == "true" {
		gologger.Info().Msg("hozor Done ... !")
		page.Timeout(time.Duration(3) * time.Second)
		Delay(time.Duration(3))

		// close popup hozor
		page.MustEvaluate(rod.Eval(`() => {
			document.querySelector("#myModal4 > div > div > div.modal-header > button").click()
		}`).ByUser())

		gologger.Info().Msg("Submiting Gozaresh ... !")
		// sabt gozaresh
		page.MustEvaluate(rod.Eval(fmt.Sprintf(`() => {
			document.querySelector("#res-hzor > div:nth-child(1) > div > div:nth-child(1) > div.col-md-4.btn-action-user > button").click()
			document.querySelector("#myModal1 > div > div > div.modal-body.h4.text-dark > form > textarea").value="%s"
		}`, randomReport.Dsc)).ByUser())
		w = page.MustWaitRequestIdle()
		page.MustEvaluate(rod.Eval(`() => {
			document.querySelector("#myModal1 > div > div > div.modal-body.h4.text-dark > form > div.modal-footer > button").click()
		}`).ByUser())
		w()

		// check gozarersh
		d := page.MustEvaluate(rod.Eval(`() => {
			if(document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("ثبت شد")>0 || document.querySelector("#myModal4 > div > div > div.modal-body.h4.text-dark > form > div:nth-child(2) > div > div").innerText.search("موفقیت")>0){
				return true
			}else{
				return false
			}
		}`).ByUser())

		out := strings.TrimSpace(d.Value.String())
		status_gozaresh := string(out)
		if status_gozaresh == "true" {
			gologger.Info().Msg("Gozaresh Done ... !")

		} else {
			gologger.Error().Msg("Error in gozaresh")
			exit_app()
		}
	} else {
		gologger.Error().Msg("Error in hozor")
		exit_app()
	}
}

// Function to get the last used index
func getLastIndex() int {
	currentContent, _ := ioutil.ReadFile(lastIndexFile)

	d := string(currentContent)
	if d == "" {
		return 0
	}

	i, _ := strconv.Atoi(d)
	return i
}

func exit_app() {
	reader := bufio.NewReader(os.Stdin)

	gologger.Warning().Msg("Press Enter to Exit ... !")

	// Loop indefinitely until the user presses Enter
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			gologger.Fatal().Msg("Error reading input: " + err.Error())
			os.Exit(1)
		}
		gologger.Info().Msg("Exiting ... ")
		// Add any cleanup tasks here if needed
		os.Exit(0)
	}
}

// Function to save the last used index
func saveLastIndex(index int) error {
	file, err := os.Create(lastIndexFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", index)
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
