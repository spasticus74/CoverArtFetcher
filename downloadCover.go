package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	caa "github.com/mineo/gocaa"
)

type exclusion struct {
	artist string
	album  string
}

func (e *exclusion) ToString() []string {
	s := make([]string, 2)
	s[0] = e.artist
	s[1] = e.album
	return s
}

func downloadCover(releaseID, outputDestination string) error {
	c := caa.NewCAAClient("CoverArtFetcher")

	log.Printf("  Downloading '%s' to '%s'\n", releaseID, outputDestination)
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
		log.Println("  Done.")
	case "image/png":
		out, err := os.Create(outputDestination + ".png")
		if err != nil {
			return err
		}

		err = png.Encode(out, img)
		if err != nil {
			return err
		}
		log.Println("  Done.")
	default:
		return errors.New("Unhandled MIME type " + imgData.Mimetype)
	}
	return nil
}

func getReleaseMBID(artist, release string) (string, error) {
	log.Printf("Searching for '%s' - '%s' ...\n", artist, release)

	var mbArtist, artistMBID = SearchArtistMBID(artist)
	if mbArtist == "[no artist]" {
		log.Printf("  * No Artist matching '%s' was found.\n", artist)
		return "", errors.New("Artist not found")
	} else if mbArtist != artist {
		if strings.ToLower(mbArtist) == strings.ToLower(artist) {
			mbArtist, artistMBID = SearchArtistMBID(mbArtist)
		} else {
			log.Printf("  * No Artist matching '%s' was found. Did you mean '%s'?\n", artist, mbArtist)
			return "", errors.New("Artist not found")
		}
	}

	recordingMBID, err := SearchReleaseMBID(artistMBID, release)
	if err != nil {
		return "", err
	}
	log.Printf("  Found Release '%s' with MBID '%s'\n", release, recordingMBID)

	return recordingMBID, nil
}

// Download a single specified cover
func FetchCover(artist, release, outputDir string) {
	releaseMBID, err := getReleaseMBID(artist, release)
	if err != nil {
		log.Fatal("  * Release MBID not found:", err)
	}

	downloadErrors := downloadCover(releaseMBID, outputDir+"/"+release+"/"+release)
	if downloadErrors != nil {
		log.Fatal("  ! Failed to download cover:", err)
	}
}

var artist string
var album string

// Download a number of missing covers
func FetchRandomMissing(dbPath, albumPath, excludePath string, maxAlbums int) {
	var newExclusions = false

	var exclusions []exclusion
	exclusions, err := ReadExcludeFile(excludePath)
	if err != nil {
		log.Print("unable to read exclusion file (" + excludePath + ") " + err.Error())
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error opening"+dbPath+":", err)
	}

	rows, err := db.Query("SELECT name, artist FROM album WHERE cover_art_path = '' AND id IN (SELECT id FROM album ORDER BY RANDOM() LIMIT ?)", maxAlbums)
	if err != nil {
		log.Fatal("Error retrieving data from database:", err)
	}
	defer rows.Close()

	// pause a random few seconds between requests so we don't hammer the server
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for rows.Next() {

		pauseSeconds := r1.Intn(10) + 5 // 5 to 15 seconds
		time.Sleep(time.Duration(pauseSeconds) * time.Second)

		err = rows.Scan(&album, &artist)
		if err != nil {
			log.Fatal("Error processing data: ", err)
		}

		if Exclude(artist, album, exclusions) {
			log.Println("  - Skipping known missing file")
			continue
		}

		releaseMBID, err := getReleaseMBID(artist, album)
		if err != nil {
			newExclusions = true
			exclusions = append(exclusions, exclusion{artist: artist, album: album})
			log.Println("  * Release MBID not found:", err)
			continue
		}

		err = downloadCover(releaseMBID, albumPath+"/"+artist+"/"+album+"/Folder")
		if err != nil {
			newExclusions = true
			exclusions = append(exclusions, exclusion{artist: artist, album: album})
			log.Println("  ! Failed to download cover:", err)
			continue
		}
	}

	if newExclusions {
		_ = WriteExcludeFile(excludePath, exclusions)
	}
}

func ReadExcludeFile(exc string) ([]exclusion, error) {
	var ex []exclusion

	f, err := os.Open(exc)
	if err != nil {
		return ex, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ex, err
		}

		ex = append(ex, exclusion{artist: record[0], album: record[1]})
	}

	log.Printf("Read %d exclusions\n", len(ex))
	return ex, nil
}

func WriteExcludeFile(exc string, exclusions []exclusion) error {
	f, err := os.OpenFile(exc, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewWriter(f)

	for _, e := range exclusions {
		err := r.Write(e.ToString())
		if err != nil {
			log.Fatal("error updating exclude file: " + err.Error())
		}
	}
	r.Flush()

	log.Printf("\nWrote %d exclusions to %s\n", len(exclusions), exc)
	return nil
}

func Exclude(artist, album string, exclusions []exclusion) bool {
	var found = false

	for _, e := range exclusions {
		if artist == e.artist && album == e.album {
			found = true
			break
		}
	}
	return found
}
