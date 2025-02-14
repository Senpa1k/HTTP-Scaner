package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func clean(arr []string) []string {
	newArr := []string{}

	for _, str := range arr {
		if len(str) > 2 {
			newArr = append(newArr, str)
		}
	}
	return newArr
}

func getClientResponse(url string, c *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.OpenFile("status.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	response, err := c.Get(url)
	if err != nil {
		stre := "The" + url + "in nit responding" + "\n"
		file.WriteString(stre)
		return
	}
	defer response.Body.Close()

	str := string(response.Status) + " " + url + "\n"

	file.WriteString(str)
}

func openListOfSites() []string {
	sites, err := os.Open("ListOfSites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer sites.Close()

	scanner := bufio.NewScanner(sites)

	arrStr := []string{}
	for scanner.Scan() {
		arrStr = append(arrStr, string(scanner.Text()))
	}

	return arrStr
}

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	wg := sync.WaitGroup{}
	file, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	data := clean(openListOfSites())

	for _, str := range data {
		wg.Add(1)
		go getClientResponse(str, client, &wg)
	}

	wg.Wait()

}
