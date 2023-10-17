package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

const PATH_UPLOAD = "photo"

func main() {
	http.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir("views"))))
	http.Handle("/compress/", http.StripPrefix("/compress/", http.FileServer(http.Dir("compress"))))
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir("photo"))))

	http.HandleFunc("/compress", Compress)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/compress.html")
	})

	http.ListenAndServe(":9999", nil)
}

// Handler for the root URL ("/").
func Compress(w http.ResponseWriter, r *http.Request) {
	// Response for JS -> OK or NG
	response := "OK"

	// Parse the form data, specifying a maximum of 10 MB for the file(s)
	r.ParseMultipartForm(10 << 20)

	// Get a reference to the uploaded file(s)
	files := r.MultipartForm.File["file"]

	// Process each uploaded file
	for _, fileHeader := range files {

		file, err := fileHeader.Open() // Open the uploaded file
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		filePath := fmt.Sprintf(""+PATH_UPLOAD+"/original/%s", fileHeader.Filename) // Path save file
		destFile, err := os.Create(filePath)                                        // Create Destination path
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer destFile.Close()

		// Copy the uploaded file's contents to the destination file
		_, err = io.Copy(destFile, file) // Upload to path destination
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Open the input image file
		inputFile := PATH_UPLOAD + "/original/" + fileHeader.Filename // Replace with your input image file

		// Open the image file.
		img, err := imaging.Open(inputFile)
		if err != nil {
			fmt.Println("Open the image, error :", err.Error())
		}
		// Open the image file again to read EXIF data.
		files, err := os.Open(inputFile)
		if err != nil {
			fmt.Println("Open the image file again to read exif data, error :", err)
			return
		}
		defer files.Close()

		// Read the EXIF data.
		x, err := exif.Decode(files)
		if err != nil {
			fmt.Println("Read the exif data, error :", err)
		} else {
			// Get the EXIF orientation tag.
			orientationTag, err := x.Get(exif.Orientation)
			fmt.Println(orientationTag)
			if err == nil {
				orientation, _ := orientationTag.Int(0)
				switch orientation {
				case 2:
					img = imaging.FlipH(img)
				case 3:
					img = imaging.Rotate180(img)
				case 4:
					img = imaging.FlipV(img)
				case 5:
					img = imaging.Transpose(img)
				case 6:
					img = imaging.Rotate90(img)
				case 7:
					img = imaging.Transverse(img)
				case 8:
					img = imaging.Rotate90(img)
				}
			}

			// Save the image with the correct orientation.
			err = imaging.Save(img, inputFile)
			if err != nil {
				fmt.Println("Error 4:", err)
				// return
			}

			fmt.Printf("Image with correct orientation saved to %s\n", inputFile)
		}

		files1, err := os.Open(inputFile)
		if err != nil {
			fmt.Println("Error opening input file:", err)
			// return
		}
		defer files1.Close()

		// Decode the input image
		img1, _, err := image.Decode(files1)
		if err != nil {
			fmt.Println("Error decoding input image:", err.Error())
			// return
		}

		// Define percentage values for width and height for two different resizes
		widthPercent1 := 50  // 50% width for the first resize
		heightPercent1 := 50 // 75% height for the first resize

		// Perform the first resize
		resizedImg1 := resizeImage(img1, widthPercent1, heightPercent1)

		// Create output filenames for the resized images
		outputFile1 := generateResizedFileName(inputFile, widthPercent1, heightPercent1)

		// Replace path
		outputFile1 = strings.ReplaceAll(outputFile1, "original", "compress")

		// Save the resized images to their respective output files
		saveImageToFile(resizedImg1, outputFile1)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// resizeImage resizes an image based on width and height percentages
func resizeImage(img image.Image, widthPercent, heightPercent int) image.Image {
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	fmt.Println("originalWidth = ", originalWidth)
	fmt.Println("originalHeight = ", originalHeight)

	var newWidth, aspect_ratio, new_height int
	if widthPercent == 200 {
		newWidth = widthPercent

		if originalHeight > originalWidth {
			aspect_ratio = 200 / originalWidth
			new_height = (aspect_ratio * originalHeight) / 100
		} else if originalHeight == originalWidth {
			new_height = newWidth
		} else {
			aspect_ratio = (originalHeight / originalWidth)
			new_height = (originalWidth * aspect_ratio)
		}
	} else {
		newWidth = originalWidth / 2
		new_height = originalHeight / 2
	}

	newHeight := new_height
	// newHeight := (originalHeight * heightPercent) / 100

	return resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
}

// generateResizedFileName generates a new filename for the resized image
func generateResizedFileName(inputFile string, widthPercent, heightPercent int) string {
	parts := strings.Split(inputFile, ".")
	if len(parts) < 2 {
		return inputFile + fmt.Sprintf("_%dx%d.jpg", widthPercent, heightPercent)
	}
	name := strings.Join(parts[:len(parts)-1], ".")
	ext := parts[len(parts)-1]

	return fmt.Sprintf("%s_%d%%x%d%%.%s", name, widthPercent, heightPercent, ext)
}

// saveImageToFile encodes and saves an image to a file
func saveImageToFile(img image.Image, filename string) {
	filename = strings.ReplaceAll(filename, "%", "")
	outFile, err := os.Create(filename)

	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, img, nil)
	if err != nil {
		fmt.Println("Error encoding and saving resized image (JPEG) :", err.Error())
		errs := png.Encode(outFile, img)
		if errs != nil {
			fmt.Println("Error encoding and saving resized image (PNG) :", errs.Error())
			return
		}
	}
}
