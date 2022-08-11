package capture

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"log"
	"os"
	"time"
	"y-clouds.com/tarantula/tools"
)

// AMAZON_URL_PREFIX Amazon detail url
const AMAZON_URL_PREFIX = "http://www.amazon.%v/gp/product/%v"

var detailElementIDOfCountry = map[string]string{"fr": "productDetails_feature_div", "de": "detailBullets_feature_div", "it": "productDetailsWithModules_feature_div", "es": "productDetails2_feature_div"}

type Amazon struct {
	Asin       string
	Country    string
	DriverPath string
	Port       int
}

func (amazon *Amazon) Url() string {
	return fmt.Sprintf(AMAZON_URL_PREFIX, amazon.Country, amazon.Asin)
}

func elementScreenshotsByXpath(wd selenium.WebDriver, path string) ([]byte, error) {
	fmt.Println("elementScreenshotsByXpath: ", path)
	ele, err := wd.FindElement(selenium.ByXPATH, path)
	if err != nil {
		log.Println("findElement.error:", path)
		return nil, err
	}

	imgBytes, err := ele.Screenshot(true)
	if err != nil {
		log.Println("elementScreenshotsByXpath.error:", path)
		return nil, err
	}

	return imgBytes, nil
}

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

func (amazon Amazon) WebScreenshots() (float32, []byte, string) {
	var (
		driverPath = amazon.DriverPath
		port       = amazon.Port
	)
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.x
		selenium.GeckoDriver(driverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),       // Output debug information to STDERR.
	}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			"--headless",
			"--start-maximized",
			//"--window-size=1080x920",
			"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			"--disable-gpu",
			"--disable-impl-side-painting",
			"--disable-gpu-sandbox",
			"--disable-accelerated-2d-canvas",
			"--disable-accelerated-jpeg-decoding",
			"--test-type=ui",
		},
	}
	selenium.SetDebug(false)
	fmt.Println("NewGeckoDriverService:")
	fmt.Println(driverPath)
	_, err := selenium.NewGeckoDriverService(driverPath, port, opts[0])
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	//defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	err = firefoxCaps.SetProfile("/Users/jinrunli/Library/Application Support/Firefox/Profiles/koagcmlb.default-1630492705887")
	if err != nil {
		log.Fatalf("set profile.error: %v", err)
	}
	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	// wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Println("NewRemote failed to connect:", err)
	}
	//defer wd.Quit()

	//wd := webdriverInit(amazon.DriverPath, amazon.Port)

	fmt.Println("url: ", amazon.Url())
	// Navigate to the simple playground interface.
	if err := wd.Get(amazon.Url()); err != nil {
		log.Println("web.open:", err)
		return 0.0, nil, string(PAGE_ERROR)
	}

	// detail image
	//detailImage, err := elementScreenshotsByXpath(wd, "//*[@id=\"ppd\"]/div[2]")
	//if err != nil {
	//	panic(err)
	//}

	// description image
	//desImages, err := elementScreenshotsByXpath(wd, "//*[@id=\"detailBulletsReverseInterleaveContainer\"]")

	keepa, err := wd.FindElement(selenium.ByID, "keepa")
	if err != nil {
		panic(err)
	}

	fmt.Println(keepa.Size())

	time.Sleep(10 * time.Second)

	err = wd.SwitchFrame(keepa)
	if err != nil {
		log.Fatalf("SwitchFrame failed:%v", err)
	}
	//err = wd.Wait(Enabled(selenium.ByID, "priceHistory"))
	//if err != nil {
	//	log.Fatalf("wait.error: %v", err)
	//}

	//ele, err = wd.FindElement(selenium.ByID, "priceHistory")
	//if err != nil {
	//	log.Fatalf("iframe_find.error: %v", err)
	//}
	//fmt.Println(ele.Size())

	iframeScreenshots, err := wd.Screenshot()
	tools.BytesSaveToImageFile(iframeScreenshots, "test_iframe_img.png")
	//// keepa image
	//keepaImage, err := elementScreenshotsByXpath(wd, "//*[@id=\"detailProductPage\"]")
	//if err != nil {
	//	panic(err)
	//}

	//screenshotBytes, err := tools.SplicePicsBytes(detailImage, keepaImage, true, "png")
	//if err != nil {
	//	log.Println("screenshot splice.error: ", err)
	//	return 0.0, nil, string(SCREENSHOT_ERROR)
	//}

	//screenshotBytes, err = tools.SplicePicsBytes(screenshotBytes, desImages, true, "png")
	//
	//if err != nil {
	//	log.Println("screenshot splice.error: ", err)
	//	return 0.0, nil, string(SCREENSHOT_ERROR)
	//}

	return 0.0, nil, string(SUCCESS)
}
