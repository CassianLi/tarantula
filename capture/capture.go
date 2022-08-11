package capture

// Selenium is the selenium attr
type Selenium struct {
	DriverPath string
	Port       int
}

type ScreenshotsStatus string

// enum status of tarantula
const (
	SUCCESS             ScreenshotsStatus = "SUCCESS"
	PAGE_ERROR          ScreenshotsStatus = "PAGE_ERROR"
	SCREENSHOT_ERROR    ScreenshotsStatus = "SCREENSHOT_ERROR"
	PRICE_ERROR         ScreenshotsStatus = "PRICE_ERROR"
	UPLOAD_TO_OSS_ERROR ScreenshotsStatus = "UPLOAD_TO_OSS_ERROR"
)

// ScreenshotsParam
// @Description: Screenshots request params
type ScreenshotsParam struct {
	Channel string `json:"channel"`
	Country string `json:"country"`
	Asin    string `json:"asin"`
	Price   string `json:"price"`
	PriceNo string `json:"priceNo"`
}

// ScreenshotsResult
// @Description: The result of tarantula
type ScreenshotsResult struct {
	Channel    string  `json:"channel"`
	Country    string  `json:"country"`
	Asin       string  `json:"asin"`
	Price      string  `json:"price"`
	PriceNo    string  `json:"priceNo"`
	Status     string  `json:"status"`
	Screenshot string  `json:"screenshot"`
	NewPrice   float32 `json:"newPrice"`
}

type Screenshots interface {
	// Url is used to make url of webpage
	Url() string

	// WebScreenshots is used to capture web pictures
	WebScreenshots() (float32, []byte, string)
}
