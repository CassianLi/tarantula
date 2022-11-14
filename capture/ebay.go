package capture

import (
	"errors"
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
func getPriceExpr(text string) (float32, error) {
	reg := regexp.MustCompile(`\d+\.\d+`)
	s := reg.FindAllString(text, -1)
	fmt.Println("FindAllString", s)
	if len(s) > 0 {
		fmt.Println("price text: ", s[0])
		price, err := strconv.ParseFloat(s[0], 32)
		if err != nil {
			return 0.0, err
		}
		return float32(price), nil
	}
	return 0.0, nil
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
func elementScreenshots(wd selenium.WebDriver, eleId string) ([]byte, error) {
	ele, err := wd.FindElement(selenium.ByID, eleId)
	if err != nil {
		fmt.Println("By.ID ElementScreenshots error:", err)
		return nil, err
	}

	eleImage, err := ele.Screenshot(true)
	if err != nil {
		return nil, err
	}

	return eleImage, nil
}

func getDescriptionCutSize(wd selenium.WebDriver, eleId string, bottomId string) (int, int, error) {
	ele, err := wd.FindElement(selenium.ByID, eleId)
	if err != nil {
		return 0, 0, err
	}
	size, _ := ele.Size()

	bootomEle, err := wd.FindElement(selenium.ByID, bottomId)
	if err != nil {
		return 0, 0, err
	}

	bottomSize, _ := bootomEle.Size()

	return size.Width, size.Height - bottomSize.Height, nil
}

// getPrice Get the price by element id
func getPrice(wd selenium.WebDriver) (float32, error) {
	// Get price panel
	ebayPriceXpaths := []string{
		"//*[@id=\"prcIsum\"]",
		"//*[@id=\"mainContent\"]/form/div[2]/div/div[1]/div/div[2]/div[1]/span[1]",
	}

	for _, xpath := range ebayPriceXpaths {
		elem, err := wd.FindElement(selenium.ByXPATH, xpath)
		if err != nil {
			log.Printf("Price XPath:%s find element error ,price.error:%v \n", xpath, err)
		} else {
			priceText, err := elem.Text()
			if err != nil {
				log.Println("Get element text, price.error:", err)
				return 0.0, err
			}

			fmt.Println("price text: ", priceText)

			price, err := getPriceExpr(priceText)
			if err != nil {
				log.Println("Get element text expr, price.error:", err)
				return 0.0, err
			}
			return price, nil
		}
	}
	return 0.0, errors.New("the element id can not find the price element")
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

	// Get price
	price, err := getPrice(wd)
	if err != nil {
		log.Printf("Find price element error: %v \n", err)
		return price, nil, string(PRICE_ERROR)
	}

	// Screenshot
	// cut two image to one
	detailImgBytes, err := elementScreenshots(wd, EBAY_DETAIL_ELE_ID)
	if err != nil || len(detailImgBytes) == 0 {
		log.Printf("Cant find element by.ID: %s \n", EBAY_DETAIL_ELE_ID)
		return price, nil, string(SCREENSHOT_ERROR)
	}
	fmt.Println("len(detailImgBytes): ", len(detailImgBytes))

	descriptionImgBytes, err := elementScreenshots(wd, EBAY_DESRIPTION_ELE_ID)
	if err != nil || len(descriptionImgBytes) == 0 {
		log.Printf("Cant find element by.ID: %s \n", EBAY_DESRIPTION_ELE_ID)
		return price, nil, string(SCREENSHOT_ERROR)
	}

	// cut picture
	width, height, err := getDescriptionCutSize(wd, EBAY_DESRIPTION_ELE_ID, EBAY_DESRIPTION_WRAPPER_ID)
	if err == nil {
		descriptionImgBytes, err = tools.CutPicture(descriptionImgBytes, 0, 0, width, height)
		if err != nil {
			log.Printf("Resize(%d, %d) descriptionImgBytes.error: %v", width, height, err)
		}
	}

	if len(detailImgBytes) > 0 && len(descriptionImgBytes) > 0 {
		// splice
		screenshotBytes, err := tools.SplicePicsBytes(detailImgBytes, descriptionImgBytes, true, "png")
		if err != nil {
			log.Println("screenshot.error: ", err)
			return price, nil, string(SCREENSHOT_ERROR)
		}
		return price, screenshotBytes, string(SUCCESS)
	}

	return price, nil, string(SCREENSHOT_ERROR)

}
