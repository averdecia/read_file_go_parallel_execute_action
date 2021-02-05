package command

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

var elements = make(chan []string)
var start = time.Now()
var filesCount = 0
var successUsersCount = 0
var successFilesLastCheckBySeconds = 0
var failedUsersCount = 0

// ICommand is used for definning the action to execute
type ICommand interface {
	ExecuteAction(element []string) (string, error)
}

func waitToFinish(outputPointer *csv.Writer, outputFile *os.File, seconds int) {

	var wg sync.WaitGroup
	wg.Add(1)

	ticker := time.NewTicker(time.Duration(seconds) * time.Second)
	quit := make(chan struct{})
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				if filesCount == successFilesLastCheckBySeconds {
					printProgress()
					close(quit)
				} else {
					successFilesLastCheckBySeconds = filesCount
				}
			case <-quit:
				ticker.Stop()
				closeOutputWriter(outputPointer, outputFile)
				return
			}
		}
	}()
	fmt.Println("Waiting To Finish")
	wg.Wait()

	fmt.Println("\nTerminating Program")
}

func closeOutputWriter(pointer *csv.Writer, file *os.File) {
	pointer.Flush()
	file.Close()
}

func printProgress() {
	seconds := time.Since(start).Seconds()
	fmt.Printf("Mean velocity: %v f/s -- Index: %v files -- Success: %v files -- Failed: %v files \n",
		math.Round(float64(filesCount)/seconds),
		filesCount,
		successUsersCount,
		failedUsersCount,
	)
}

func definedOrEmpty(arr []string, pos int) string {
	if len(arr) >= pos+1 {
		return arr[pos]
	}
	return ""
}

func readFile(filepath string, separator string, elements chan []string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Println("Reading file")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), separator)
		if len(split) < 1 {
			fmt.Printf("Ignored small Element: %v \n", scanner.Text())
			continue
		}
		elements <- split
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ending reading file")
}

// createRoutines is used to create de amount of needed routines
func createRoutines(command ICommand, count int, outputPointer *csv.Writer) {
	for i := 0; i < count; i++ {
		time.Sleep(1 * time.Second)
		go func() {
			for {
				select {
				case element := <-elements:
					executeAction(command, element, outputPointer)
					filesCount++
					printProgress()
				}
			}
		}()
	}
}

func executeAction(command ICommand, element []string, outputPointer *csv.Writer) {
	res, err := command.ExecuteAction(element)

	if err != nil {
		fmt.Printf("Server error: %v \n", err)
		failedUsersCount++
		element = append(element, err.Error())
		outputPointer.Write(element)
		outputPointer.Flush()
		return
	}
	fmt.Printf("Response: %v \n", res)
	successUsersCount++
}

func getOutputWriter(outputPath string) (*csv.Writer, *os.File) {
	csvfile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		fmt.Printf("Failed creating file: %s \n", err)
		os.Exit(0)
	}
	pointer := csv.NewWriter(csvfile)

	return pointer, csvfile
}

// RunProcess is the function to trigger the full process
func RunProcess(command ICommand, routines int, inFile string, outFile string, seconds int) {

	outputPointer, outputFile := getOutputWriter(outFile)
	createRoutines(command, routines, outputPointer)
	readFile(inFile, ",", elements)

	waitToFinish(outputPointer, outputFile, seconds)
}
