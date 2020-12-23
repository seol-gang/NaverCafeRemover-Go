package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// Account struct
type Account struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

//Enabled function
func Enabled(by, elementName string) func(selenium.WebDriver) (bool, error) {
	return func(wd selenium.WebDriver) (bool, error) {
		el, err := wd.FindElement(by, elementName)
		if err != nil {
			return false, nil
		}
		enabled, err := el.IsEnabled()
		if err != nil {
			return false, nil
		}

		if !enabled {
			return false, nil
		}

		return true, nil
	}
}

// GetAccountInfo fuction
func GetAccountInfo() (string, string) {
	jsonFile, err := os.Open("./accountConfig.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonValue, _ := ioutil.ReadAll(jsonFile)

	var accountInfo Account
	json.Unmarshal(jsonValue, &accountInfo)
	return accountInfo.ID, accountInfo.PW
}

// RunSeleniumClient function
func RunSeleniumClient() (selenium.WebDriver, *selenium.Service) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--headless",
		},
	}
	caps.AddChrome(chromeCaps)

	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)

	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		fmt.Println(err)
	}

	return wd, service
}

// LoginNaver function
func LoginNaver(driver selenium.WebDriver, id, pw string) error {
	script := `
	(function execute(){
		document.querySelector('#id').value = "` + id + `";
		document.querySelector('#pw').value = "` + pw + `";
	})();
	`
	driver.ExecuteScript(script, nil)
	if err := driver.Wait(Enabled(selenium.ByCSSSelector, "input.btn_global")); err != nil {
		return err
	}
	element, _ := driver.FindElement(selenium.ByCSSSelector, "input.btn_global")
	element.Click()
	return nil
}

// RunDelete function
func RunDelete(driver selenium.WebDriver, id, pw string, typeCheck int) {
	var typeString = [2]string{
		`#ia-action-data > div.ia-info-data3 > ul > li.info2 > span > strong > a`,
		`#ia-action-data > div.ia-info-data3 > ul > li.info3 > span > strong > a`,
	}

	driver.Get("https://nid.naver.com/nidlogin.login")
	if err := driver.Wait(Enabled(selenium.ByCSSSelector, `#log\.login`)); err != nil {
		fmt.Println(err)
		return
	}
	if err := LoginNaver(driver, id, pw); err != nil {
		fmt.Println(err)
		return
	}
	if err := driver.Wait(Enabled(selenium.ByCSSSelector, `#footer > div > div.corp_area > address > a`)); err != nil {
		fmt.Println(err)
		return
	}
	driver.Get("https://cafe.naver.com/<your cafe url>")

	if err := driver.Wait(Enabled(selenium.ByCSSSelector, `#cafe-info-data > ul > li.tit-action > p > a`)); err != nil {
		fmt.Println(err)
		return
	}
	element, _ := driver.FindElement(selenium.ByCSSSelector, `#cafe-info-data > ul > li.tit-action > p > a`)
	element.Click()

	if err := driver.Wait(Enabled(selenium.ByCSSSelector, typeString[typeCheck])); err != nil {
		fmt.Println(err)
		return
	}
	element, _ = driver.FindElement(selenium.ByCSSSelector, typeString[typeCheck])
	element.Click()

	//Start Switch Frame
	if err := driver.Wait(Enabled(selenium.ByCSSSelector, "#cafe_main")); err != nil {
		fmt.Println(err)
		return
	}

	element, _ = driver.FindElement(selenium.ByCSSSelector, "#cafe_main")
	element.Click()
	if err := driver.SwitchFrame(element); err != nil {
		fmt.Println(err)
		return
	}

	element, _ = driver.FindElement(selenium.ByCSSSelector, "#innerNetwork")
	element.Click()
	if err := driver.SwitchFrame(element); err != nil {
		fmt.Println(err)
		return
	}

	for {
		if err := driver.WaitWithTimeout(Enabled(selenium.ByCSSSelector, `#main-area > div.article-board.article_profile.m-tcol-c > table > tbody > tr > td.td_article > div.check_box.only_box`), time.Second*10); err != nil {
			break
		}

		element, _ = driver.FindElement(selenium.ByCSSSelector, `#main-area > div.post_btns > div.fl > div > label`)
		element.Click()

		for {
			element, _ = driver.FindElement(selenium.ByXPATH, `//*[@id="selectAll"]`)
			if check, _ := element.IsEnabled(); check {
				break
			}
			element, _ = driver.FindElement(selenium.ByCSSSelector, `#main-area > div.post_btns > div.fl > div > label`)
			element.Click()
		}

		element, _ = driver.FindElement(selenium.ByCSSSelector, "#a_remove")
		element.Click()
		time.Sleep(time.Millisecond * 500)

		driver.AcceptAlert()
	}
}

// DeleteCafeNotice function
func DeleteCafeNotice(done chan bool) {
	naverID, naverPW := GetAccountInfo()
	driver, service := RunSeleniumClient()
	defer driver.Quit()
	defer service.Stop()
	RunDelete(driver, naverID, naverPW, 0)
	done <- true
}

// DeleteCafeComment function
func DeleteCafeComment(done chan bool) {
	naverID, naverPW := GetAccountInfo()
	driver, service := RunSeleniumClient()
	defer driver.Quit()
	defer service.Stop()
	RunDelete(driver, naverID, naverPW, 1)
	done <- true
}

func main() {
	done := make(chan bool)
	go DeleteCafeNotice(done)
	go DeleteCafeComment(done)

	for i := 0; i < 2; i++ {
		<-done
	}
}
