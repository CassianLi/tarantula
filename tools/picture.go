package tools

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

// SpliceImage Splice two image.Image picture
func SpliceImage(image1 image.Image, image2 image.Image, vertical bool) *image.RGBA {
	//starting position of the second image (bottom/left)
	sp2 := image.Point{X: image1.Bounds().Dx()}
	if vertical {
		sp2 = image.Point{Y: image1.Bounds().Dy()}
	}

	//new rectangle for the second image
	r2 := image.Rectangle{Min: sp2, Max: sp2.Add(image2.Bounds().Size())}

	//rectangle for the big image
	r := image.Rectangle{Min: image.Point{}, Max: r2.Max}

	rgba := image.NewRGBA(r)

	draw.Draw(rgba, image1.Bounds(), image1, image.Point{}, draw.Src)
	draw.Draw(rgba, r2, image2, image.Point{}, draw.Src)

	return rgba
}

// SplicePics Splice two pictures into one
// img1 is the first picture path
// img2 is the second picture path
// outImage is the mosaic path, only png and jpeg
// vertical indicates whether the splicing vertical
func SplicePics(img1 string, img2 string, outImage string, vertical bool) error {
	imgFile1, err := os.Open(img1)
	imgFile2, err := os.Open(img2)

	image1, _, err := image.Decode(imgFile1)
	image2, _, err := image.Decode(imgFile2)

	rgba := SpliceImage(image1, image2, vertical)

	out, err := os.Create(outImage)

	// png / jpeg
	if strings.HasSuffix(outImage, "png") {
		err = png.Encode(out, rgba)
	}

	if strings.HasSuffix(outImage, "jpeg") {
		var opt jpeg.Options
		opt.Quality = 80
		err = jpeg.Encode(out, rgba, &opt)
	}

	return err
}

// SplicePicsBytes Splice two pictures' byte array  into one byte array
// img1 is the first picture byte array
// img2 is the second picture byte array
// vertical indicates whether the splicing vertical
// imageFormat is the format of output image, png / jpeg
func SplicePicsBytes(img1 []byte, img2 []byte, vertical bool, imageFormat string) ([]byte, error) {
	image1, _, err := image.Decode(bytes.NewReader(img1))
	image2, _, err := image.Decode(bytes.NewReader(img2))

	// splice two pics
	rgba := SpliceImage(image1, image2, vertical)

	img, err := ImageToBytes(rgba, imageFormat)

	return img, err
}

// ImageReadToBytes is read image from local file, return bytes
func ImageReadToBytes(imagePath string) ([]byte, error) {
	imgFile, err := os.Open(imagePath)
	img, _, err := image.Decode(imgFile)

	var formatStr = string("")
	if strings.HasSuffix(imagePath, "png") {
		formatStr = "png"
	}

	if strings.HasSuffix(imagePath, "jpeg") {
		formatStr = "jpeg"
	}

	imgBytes, err := ImageToBytes(img, formatStr)

	return imgBytes, err
}

// ImageToBytes image.Image to []byte
func ImageToBytes(img image.Image, imageFormat string) ([]byte, error) {
	buf := new(bytes.Buffer)
	// png / jpeg
	var err = error(nil)
	if imageFormat == "png" {
		err = png.Encode(buf, img)
	}

	if imageFormat == "jpeg" {
		var opt jpeg.Options
		opt.Quality = 80
		err = jpeg.Encode(buf, img, &opt)
	}

	return buf.Bytes(), err
}

// BytesSaveToImageFile save []byte to local image file
func BytesSaveToImageFile(imgBytes []byte, outPath string) error {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))

	err = ImageSaveToImageFile(img, outPath)

	return err
}

// ImageSaveToImageFile Save Image object to local file
func ImageSaveToImageFile(img image.Image, outPath string) error {
	out, err := os.Create(outPath)

	// png / jpegs
	if strings.HasSuffix(outPath, "png") {
		err = png.Encode(out, img)
	}

	if strings.HasSuffix(outPath, "jpeg") {
		var opt jpeg.Options
		opt.Quality = 80
		err = jpeg.Encode(out, img, &opt)
	}

	return err
}

func CutPicture(imgBytes []byte, x0 int, y0 int, x1 int, y1 int) ([]byte, error) {
	rg := image.Rectangle{Min: image.Point{X: x0, Y: y0}, Max: image.Point{X: x1, Y: y1}}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Printf("image decode.error: %v, %d, %d", err, x1, y1)
	}
	rgba := image.NewRGBA(rg)
	draw.Draw(rgba, img.Bounds(), img, image.Point{}, draw.Src)

	return ImageToBytes(rgba, "png")
}
