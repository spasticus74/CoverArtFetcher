package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {

	artistPtr := flag.String("a", "", "The artist")
	recordingPtr := flag.String("r", "", "The recording")
	destinationPtr := flag.String("d", ".", "THe directory to write the output to.")

	flag.Parse()

	if (len(*artistPtr) == 0) || (len(*recordingPtr) == 0) {
		log.Fatal("You must provide both an Artist and a Recording name")
	}

	var artistName, artistMBID = SearchArtistMBID(*artistPtr)
	if artistName == "[no artist]" {
		fmt.Println("No Artist matching '", *artistPtr, "' was found.")
		os.Exit(1)
	} else if artistName != *artistPtr {
		fmt.Println("No Artist matching '", *artistPtr, "' was found. Did you mean '", artistName, "?")
		os.Exit(1)
	}
	fmt.Printf("Found Artist '%s' with MBID %s\n", *artistPtr, artistMBID)

	recordingMBID, err := SearchReleaseMBID(artistMBID, *recordingPtr)
	if err != nil {
		log.Fatal("Error finding release ID: ", err)
	} else {
		fmt.Printf("  Found Release '%s' with MBID %s.\n  Downloading ...\n", *recordingPtr, recordingMBID)
		err = DownloadCover(recordingMBID, *destinationPtr, *recordingPtr)
		if err != nil {
			log.Fatal("Error downloading cover image: ", err)
		}
		fmt.Println("  Done.")
	}
}
