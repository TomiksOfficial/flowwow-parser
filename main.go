package main

import (
	"encoding/json"
	"fmt"
	"github.com/gebleksengek/useragents"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"math/rand"
	"os"
	"time"
)

const PageCount = 5

type FlowerItem struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Image string `json:"image"`
}

func randomSleep(min, max int) {
	duration := time.Duration(rand.Intn(max-min)+min) * time.Millisecond
	time.Sleep(duration)
}

func main() {
	var parsedFlowers []FlowerItem

	service, err := selenium.NewChromeDriverService("./chrome/chromedriver", 4444)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--window-size=1920,1080",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--disable-extensions",
			"--headless",
			fmt.Sprintf("--user-agent=%s", useragents.ChromeLatest()),
		},
	})

	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	script := `
        Object.defineProperty(navigator, 'webdriver', {
            get: () => undefined,
        });
        window.navigator.chrome = {
            runtime: {},
        };
        Object.defineProperty(navigator, 'languages', {
            get: () => ['en-US', 'en'],
        });
        Object.defineProperty(navigator, 'plugins', {
            get: () => [1, 2, 3],
        });
    `
	_, err = driver.ExecuteScript(script, nil)
	if err != nil {
		fmt.Println("Error injecting JavaScript:", err)
		return
	}

	for i := 1; i <= PageCount; i++ {
		err = driver.Get(fmt.Sprintf("https://flowwow.com/vologda/page-%d/", i))
		if err != nil {
			fmt.Println("Error driver.Get():", err)
			return
		}

		basic, err := driver.FindElement(selenium.ByCSSSelector, ".tab-content-products")
		if err != nil {
			fmt.Println("Failed to get table of elements:", err)
			return
		}

		elements, err := basic.FindElements(selenium.ByCSSSelector, ".tab-content-products-item")

		for _, item := range elements {

			pc, _ := item.FindElement(selenium.ByCSSSelector, "a.product-card")
			dname, _ := item.FindElement(selenium.ByCSSSelector, "div.name")
			img, _ := item.FindElement(selenium.ByCSSSelector, "img")

			url, _ := pc.GetAttribute("href")
			name, _ := dname.Text()
			image, _ := img.GetAttribute("src")
			parsedFlowers = append(parsedFlowers, FlowerItem{name, url, image})
		}
	}

	f, _ := os.Create("output.json")
	defer f.Close()
	js, _ := json.MarshalIndent(parsedFlowers, "", "\t")

	f.Write(js)
}
