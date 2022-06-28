package capture

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"log"
	"os"
	"regexp"
	"strconv"
	"y-clouds.com/tarantula/tools"
)

const (
	// EBAY_URL_PREFIX is the ebay url prefix
	EBAY_URL_PREFIX            = "https://www.ebay.com/itm/%v"
	EBAY_DETAIL_ELE_ID         = "CenterPanelInternal"
	EBAY_DESRIPTION_ELE_ID     = "vi-desc-maincntr"
	EBAY_DESRIPTION_WRAPPER_ID = "desc_wrapper_ctr"
)

// Ebay is the ebay params
type Ebay struct {
	Asin       string
	DriverPath string
	Port       int
}

// Url
//  @Description: Make url of ebay
//  @receiver ebay
//  @return string
func (ebay Ebay) url() string {
	return fmt.Sprintf(EBAY_URL_PREFIX, ebay.Asin)
}

// getPrice is a regular expression to get the price
func getPrice(text string) float32 {
	reg := regexp.MustCompile(`\d+\.\d+`)
	s := reg.FindAllString(text, -1)
	fmt.Println("FindAllString", s)
	if len(s) > 0 {
		fmt.Println("price text: ", s[0])
		price, err := strconv.ParseFloat(s[0], 32)
		if err != nil {
			fmt.Printf("Parse float32.error: %v", err)
			return 0.0
		}
		return float32(price)
	}
	return 0.0
}

// reSizeBrowserWindow Resize the window, or return the original WebDriver
func reSizeBrowserWindow(wd selenium.WebDriver) selenium.WebDriver {
	ele, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"viTabs_0_is\"]")
	if err != nil {
		log.Println("By.ID viTabs_0_is, ele.error:", err)
	}

	size, _ := ele.Size()
	location, _ := ele.LocationInView()
	if err != nil {
		log.Println("By.ID CenterPanelInternal, get_size.error:", err)
	} else {
		err = wd.ResizeWindow("", size.Width, size.Height+location.Y+200)
		if err != nil {
			log.Println("resize_window.error:", err)
		}
	}

	return wd
}

// elementScreenshots Take a screenshot of an element
func elementScreenshots(wd selenium.WebDriver, eleId string) []byte {
	ele, err := wd.FindElement(selenium.ByID, eleId)
	if err != nil {
		log.Printf("By.ID %s ele.error: %v", eleId, err)
	}

	eleImage, err := ele.Screenshot(true)

	if err != nil {
		log.Printf("By.ID %s, screenshot.error:%v", eleId, err)
	}

	return eleImage
}

func getDescriptionCutSize(wd selenium.WebDriver, eleId string, bottomId string) (int, int) {
	ele, err := wd.FindElement(selenium.ByID, eleId)
	if err != nil {
		log.Printf("By.ID %s ele.error: %v", eleId, err)
	}
	size, _ := ele.Size()

	bootomEle, err := wd.FindElement(selenium.ByID, bottomId)
	if err != nil {
		log.Printf("By.ID %s ele.error: %v", eleId, err)
	}

	bottomSize, _ := bootomEle.Size()

	return size.Width, size.Height - bottomSize.Height
}

// WebScreenshots
//  @return float32  the price in web page
//  @return []byte 	the tarantula of web page
//  @return string the status of tarantula
func (ebay Ebay) WebScreenshots() (float32, []byte, string) {
	// Start a Selenium WebDriver server instance (if one is not already running).
	var (
		geckoDriverPath = ebay.DriverPath
		port            = ebay.Port
	)
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.x
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			"--headless",
			"--start-maximized",
			//"--window-size=1200x600",
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
	fmt.Println(geckoDriverPath)
	service, err := selenium.NewGeckoDriverService(geckoDriverPath, port, opts[0])
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	// wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get(ebay.url()); err != nil {
		log.Println("web.open:", err)
		return 0.0, nil, string(PAGE_ERROR)
	}

	// Resize window
	//wd = reSizeBrowserWindow(wd)

	// Get price panel
	elem, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"prcIsum\"]")
	if err != nil {
		log.Println("price.error:", err)
		return 0.0, nil, string(PRICE_ERROR)
	}
	priceText, err := elem.Text()
	if err != nil {
		log.Println("price.error:", err)
		return 0.0, nil, string(PRICE_ERROR)
	}
	fmt.Println("price text: ", priceText)
	price := getPrice(priceText)

	fmt.Println("price: ", price)

	// Screenshot
	// cut two image to one
	detailImgBytes := elementScreenshots(wd, EBAY_DETAIL_ELE_ID)

	descriptionImgBytes := elementScreenshots(wd, EBAY_DESRIPTION_ELE_ID)

	// cut picture
	width, height := getDescriptionCutSize(wd, EBAY_DESRIPTION_ELE_ID, EBAY_DESRIPTION_WRAPPER_ID)
	descriptionImgBytes, err = tools.CutPicture(descriptionImgBytes, 0, 0, width, height)
	if err != nil {
		log.Printf("Resize(%d, %d) descriptionImgBytes.error: %v", width, height, err)
	}

	// splice
	screenshotBytes, err := tools.SplicePicsBytes(detailImgBytes, descriptionImgBytes, true, "png")

	if err != nil {
		log.Println("screenshot.error: ", err)
		return price, nil, string(SCREENSHOT_ERROR)
	}

	return price, screenshotBytes, string(SUCEESS)
}
