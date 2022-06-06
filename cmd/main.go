package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Property struct {
	Id            string
	StreetAddress string
	Town          string
	ValuationDate string
	Value         string
	Order         int
}

func main() {
	// Get all the properties from the given file
	properties, err := getProperties()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get properties %v", err))
	}
	fmt.Println(len(properties)) // Should be 292 lines including header

	// Test 1
	test1 := make(chan []Property)
	go removeDuplicates(properties, "last", test1)
	fmt.Println("Test1:", <-test1)

	// Test 2
	test2 := make(chan []Property)
	go removeDuplicates(properties, "", test2)
	fmt.Println("Test2:", <-test2)

	// Test 3
	test3 := make(chan []Property)
	go removeDuplicates(properties, "none", test3)
	fmt.Println("Test3:", <-test3)

	// Test 4
	test4 := make(chan []Property)
	go filterProperties(properties, test4)
	fmt.Println("Test4:", <-test4)

	// Test 4.5 (extra credit)
	chunks := chunkProperties(properties, 50)
	output := make(chan []Property)
	var wg sync.WaitGroup
	for _, chunk := range chunks {
		wg.Add(1)
		// Process each chuck with a goroutine and send output to sampleChan
		go filterProperties2(chunk, output, &wg)
	}
	go func() {
		wg.Wait()
		close(output)
	}()

	// Loop through output and put into chunkedfilteredProperties
	var chunkedfilteredProperties []Property
	for chunk := range output {
		for _, property := range chunk {
			chunkedfilteredProperties = append(chunkedfilteredProperties, property)
		}
	}
	fmt.Println("Test4.5:", chunkedfilteredProperties)
}

func getProperties() ([]Property, error) {
	file, err := os.Open("properties.txt")
	if err != nil {
		return []Property{}, fmt.Errorf("failed to open file %v", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			err = fmt.Errorf("failed to open file %v", err)
		}
	}()
	var properties []Property
	count := 0 // used to check original order of data
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// scan each line into Property struct
		splitLine := strings.Split(scanner.Text(), ",")
		if strings.TrimSpace(splitLine[0]) != "" {
			count++
			row := Property{
				Id:            strings.TrimSpace(splitLine[0]),
				StreetAddress: strings.TrimSpace(splitLine[1]),
				Town:          strings.TrimSpace(splitLine[2]),
				ValuationDate: strings.TrimSpace(splitLine[3]),
				Value:         strings.TrimSpace(splitLine[4]),
				Order:         count,
			}
			properties = append(properties, row)
		}
	}
	return properties, nil
}

func removeDuplicates(properties []Property, fill string, result chan []Property) {
	var sorted []Property
Loop:
	for _, p := range properties {
		for i, s := range sorted {
			// same comparison check allows use to reuse this.
			if strings.Join(strings.Fields(p.StreetAddress), "") == strings.Join(strings.Fields(s.StreetAddress), "") && strings.Join(strings.Fields(p.ValuationDate), "") == strings.Join(strings.Fields(s.ValuationDate), "") {
				// if we want the last encountered from comparison pass in last
				if fill == "last" {
					sorted[i] = p
				}
				// if don't want either from comparison pass in none
				if fill == "none" {
					sorted = append(sorted[:i], sorted[i+1:]...)
				}
				// if we want the first encountered from comparison we just contiune
				continue Loop
			}
		}
		sorted = append(sorted, p)
	}
	result <- sorted
}

func filterProperties(properties []Property, result chan []Property) {
	var filteredProperties []Property
	for _, property := range properties {
		convertedNumber, _ := strconv.Atoi(property.Value)
		// remove properties less then 400000
		if convertedNumber < 400000 {
			continue
		}
		// remove properties with AVE, CRES, PL in the address
		if strings.Contains(property.StreetAddress, "AVE") || strings.Contains(property.StreetAddress, "CRES") || strings.Contains(property.StreetAddress, "PL") {
			continue
		}
		// remove every 10 or divisable by 10
		correctOrder := property.Order - 1 // account for header line
		if correctOrder%10 == 0 {
			continue
		}
		filteredProperties = append(filteredProperties, property)
	}
	result <- filteredProperties
}

func chunkProperties(properties []Property, chunkSize int) [][]Property {
	var chunks [][]Property
	for {
		if len(properties) == 0 {
			break
		}
		// check properties capacity
		if len(properties) < chunkSize {
			chunkSize = len(properties)
		}
		chunks = append(chunks, properties[0:chunkSize])
		properties = properties[chunkSize:]
	}

	return chunks
}

func filterProperties2(properties []Property, result chan []Property, wg *sync.WaitGroup) {
	defer wg.Done()
	var filteredProperties []Property
	for _, property := range properties {
		convertedNumber, _ := strconv.Atoi(property.Value)
		// remove properties less then 400000
		if convertedNumber < 400000 {
			continue
		}
		// remove properties with AVE, CRES, PL in the address
		if strings.Contains(property.StreetAddress, "AVE") || strings.Contains(property.StreetAddress, "CRES") || strings.Contains(property.StreetAddress, "PL") {
			continue
		}
		// remove every 10 or divisable by 10
		correctOrder := property.Order - 1 // account for header line
		if correctOrder%10 == 0 {
			continue
		}
		filteredProperties = append(filteredProperties, property)
	}
	result <- filteredProperties
}
