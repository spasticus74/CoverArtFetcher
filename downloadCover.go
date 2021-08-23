package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	caa "github.com/mineo/gocaa"
)

func downloadCover(releaseID, outputDestination string) error {
	c := caa.NewCAAClient("CoverArtFetcher")

	fmt.Printf("Downloading '%s' to '%s'\n", releaseID, outputDestination)
	imgData, err := c.GetReleaseFront(caa.StringToUUID(releaseID), 0)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(imgData.Data))
	if err != nil {
		return err
	}

	switch imgData.Mimetype {
	case "image/jpeg":
		out, err := os.Create(outputDestination + ".jpg")
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
		out, err := os.Create(outputDestination + ".png")
		if err != nil {
			return err
		}

		err = png.Encode(out, img)
		if err != nil {
			return err
		}
	default:
		return errors.New("Unhandled MIME type " + imgData.Mimetype)
	}
	return nil
}

func getReleaseMBID(artist, release string) (string, error) {
	fmt.Printf("Searcing for '%s' - '%s' ...\n", artist, release)

	var artistName, artistMBID = SearchArtistMBID(artist)
	if artistName == "[no artist]" {
		fmt.Printf("  No Artist matching '%s' was found.\n", artist)
		return "", errors.New("Artist not found")
	} else if artistName != artist {
		fmt.Printf("  No Artist matching '%s' was found. Did you mean '%s'?\n", artist, artistName)
		return "", errors.New("Artist not found")
	}

	//fmt.Printf("Found Artist '%s' with MBID %s\n", artist, artistMBID)
	recordingMBID, err := SearchReleaseMBID(artistMBID, release)
	if err != nil {
		return "", err
	}
	fmt.Printf("  Found Release '%s' with MBID %s\n", release, recordingMBID)

	return recordingMBID, nil
}

// Download a single specified cover
func FetchCover(artist, release, outputDir string) {
	relMBID, err := getReleaseMBID(artist, release)
	if err != nil {
		log.Fatal("Release MBID not found:", err)
	}

	cover := make(map[string]string)
	cover[relMBID] = outputDir + "/" + release
	downloadErrors := downloadCover(relMBID, outputDir+"/"+release+"/"+release)
	if downloadErrors != nil {
		log.Fatal("Failed to download cover:", err)
	}
}

var artist string
var album string

// Download a number of missing covers
func FetchRandomMissing(dbPath, albumPath string, maxAlbums int) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error opening"+dbPath+":", err)
	}

	rows, err := db.Query("SELECT name, artist FROM album WHERE cover_art_path = '' AND id IN (SELECT id FROM album ORDER BY RANDOM() LIMIT ?)", maxAlbums)
	if err != nil {
		log.Fatal("Error retrieving data from database:", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&album, &artist)
		if err != nil {
			log.Fatal("Error processing data: ", err)
		}

		relMBID, err := getReleaseMBID(artist, album)
		if err != nil {
			fmt.Println("Release MBID not found:", err)
			continue
		}

		err = downloadCover(relMBID, albumPath+"/"+artist+"/"+album+"/Folder")
		if err != nil {
			fmt.Println("Failed to download cover:", err)
			continue
		}
	}
}
