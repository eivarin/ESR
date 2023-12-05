package shell

import (
	"fmt"
	"os"
	"strings"
	"flag"
)

func GetRealArgs(fileDescr string, nameDescr string) ([]string, string) {
	fileFlag := flag.String("f", "", fileDescr)
	nFlag := flag.String("n", "", nameDescr)

	// Parse the command line arguments
	flag.Parse()

	if len(flag.Args()) == 0 && *nFlag == "" && *fileFlag == "" {
		fmt.Println("No RP Address provided")
		os.Exit(1)
	}

	Neighbors := []string{}
	name := ""
	// If neither flag is provided, use args
	if !(*nFlag != "" || *fileFlag != "") {
		Neighbors = flag.Args()
		name = flag.Args()[0]
	} else if *nFlag != "" {
		Neighbors = readNeighborsFromFile("./static/configs/" + *nFlag + ".conf")
		name = *nFlag
	} else if *fileFlag != "" {
		Neighbors = readNeighborsFromFile(*fileFlag)
		name = *fileFlag
	}
	return Neighbors, name
}

func readNeighborsFromFile(filename string) []string {
	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Split the content into lines and populate the Neighbors slice
	Neighbors := strings.Split(string(content), "\n")

	// Remove empty strings from the Neighbors slice
	var cleanedNeighbors []string
	for _, neighbor := range Neighbors {
		if neighbor != "" {
			cleanedNeighbors = append(cleanedNeighbors, neighbor)
		}
	}

	// Update the Neighbors slice with cleaned data
	return cleanedNeighbors
}