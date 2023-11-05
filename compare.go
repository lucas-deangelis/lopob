///usr/bin/true; exec /usr/bin/env go run "$0" "$@"
package main

import (
	"fmt"
	"image/png"
	"os"
)

func compareImages(img1Path, img2Path string) (bool, error) {
	file1, err := os.Open(img1Path)
	if err != nil {
		fmt.Printf("Failed to open path %s: %s\n", img1Path, err)
		return false, err
	}
	defer file1.Close()

	file2, err := os.Open(img2Path)
	if err != nil {
		fmt.Printf("Failed to open path %s: %s\n", img2Path, err)
		return false, err
	}
	defer file2.Close()

	img1, err := png.Decode(file1)
	if err != nil {
		fmt.Printf("Failed to decode image %s: %s\n", file1.Name(), err)
		return false, err
	}

	img2, err := png.Decode(file2)
	if err != nil {
		fmt.Printf("Failed to decode image %s: %s\n", file2.Name(), err)
		return false, err
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()
	if bounds1 != bounds2 {
		fmt.Printf("Images have different bounds: %v != %v\n", bounds1, bounds2)
		return false, nil
	}

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				fmt.Printf("Pixel at (%d, %d) is different\n", x, y)
				// Print the pixels:
				fmt.Printf("Pixel at (%d, %d): %v\n", x, y, img1.At(x, y))
				fmt.Printf("Pixel at (%d, %d): %v\n", x, y, img2.At(x, y))
				return false, nil
			}
		}
	}

	return true, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run compare_images.go <image1.png> <image2.png>")
		os.Exit(1)
	}

	img1Path := os.Args[1]
	img2Path := os.Args[2]

	same, err := compareImages(img1Path, img2Path)
	if err != nil {
		fmt.Printf("Failed to compare images: %s\n", err)
		os.Exit(1)
	}
	if same {
		fmt.Println("The images are the same.")
	} else {
		fmt.Println("The images are different.")
	}
}
