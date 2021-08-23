package main

import (
	"flag"
	"log"
)

func main() {

	// Two ways to use this.
	// 1. Fetch a specific album/recording
	artistPtr := flag.String("a", "", "The artist.")
	recordingPtr := flag.String("r", "", "The recording.")
	destinationPtr := flag.String("o", ".", "The directory to write the output to.")

	// 2. Point it to the Navidrome DB & music library and let it find missing covers
	dbPtr := flag.String("d", "", "Path to the Navidrome database.")
	libraryPtr := flag.String("l", "", "Path to the music library.")
	maxCountPtr := flag.Int("c", 10, "Max number of covers to fetch")

	flag.Parse()

	if (len(*artistPtr) > 0) && (len(*recordingPtr) > 0) {
		FetchCover(*artistPtr, *recordingPtr, *destinationPtr)
	} else if (len(*dbPtr) > 0) && (len(*libraryPtr) > 0) {
		FetchRandomMissing(*dbPtr, *libraryPtr, *maxCountPtr)
	} else {
		log.Fatal("You must provide both an Artist and a Recording name, or a database and library path.")
	}
}
