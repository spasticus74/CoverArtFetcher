package main

import (
	"errors"
	"strings"

	"github.com/michiwend/gomusicbrainz"
)

const artistSearchPrefix = "`artist:\""
const recordingSearchPrefix = "`recording:\""

func SearchArtistMBID(name string) (string, string) {
	client, _ := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"CoverArtFetcher",
		"0.0.1-alpha",
		"spasticus74@gmail.com")

	// Search for some artist(s)
	searchTerm := artistSearchPrefix + name + "\"`"
	resp, _ := client.SearchArtist(searchTerm, -1, -1)

	var aName string
	var mb string
	for _, artist := range resp.Artists {
		if resp.Scores[artist] == 100 {
			aName = string(artist.Name)
			mb = string(artist.Id())
			break
		}
	}
	return aName, mb
}

func SearchReleaseMBID(artistID, name string) (string, error) {
	client, _ := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"CoverArtFetcher",
		"0.0.1-alpha",
		"spasticus74@gmail.com")

	// Search for the recording
	searchTerm := recordingSearchPrefix + name + "\" AND arid:\"" + artistID + "\"`"
	//fmt.Println(searchTerm)
	resp, _ := client.SearchRelease(searchTerm, -1, -1)

	for _, rec := range resp.Releases {
		for _, a := range rec.ArtistCredit.NameCredits {
			//fmt.Println("T: ", rec.Title, ", A: ", a.Artist.Name, ", ID: ", string(rec.Id()))
			if (a.Artist.ID == gomusicbrainz.MBID(artistID)) && strings.EqualFold(rec.Title, name) {
				return string(rec.Id()), nil
			}
		}
	}
	return "", errors.New("not found")
}
