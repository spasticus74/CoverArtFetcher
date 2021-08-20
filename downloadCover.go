package main

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	caa "github.com/spasticus74/gocaa"
)

func DownloadCover(mbid, outputDir, name string) error {
	c := caa.NewCAAClient("CoverArtFetcher")

	imgData, err := c.GetReleaseFront(caa.StringToUUID(mbid), 0)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(imgData.Data))
	if err != nil {
		log.Fatal("Error converting image data:", err)
	}

	switch imgData.Mimetype {
	case "image/jpeg":
		out, err := os.Create(outputDir + "/" + name + ".jpg")
		if err != nil {
			return err
		}

		var opt jpeg.Options
		opt.Quality = 100

		err = jpeg.Encode(out, img, &opt)
		if err != nil {
			return err
		}
	case "image/png":
		out, err := os.Create(outputDir + "/" + name + ".png")
		if err != nil {
			return err
		}

		err = png.Encode(out, img)
		if err != nil {
			return err
		}
	default:
		return errors.New("Unhandled Mimetype: " + imgData.Mimetype)
	}
	return nil
}
