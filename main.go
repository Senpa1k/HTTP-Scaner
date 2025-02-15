package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type SiteScanner struct {
	Client         *http.Client
	Sites          []string
	StatusLog      *os.File
	ErrorLog       *os.File
	LogMutex       sync.Mutex
	AmountAttempts int8
}

func clean(arr []string) []string {
	newArr := []string{}
	for _, str := range arr {
		if len(str) > 2 {
			newArr = append(newArr, str)
		}
	}
	return newArr
}

func (s *SiteScanner) Check(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	var response *http.Response
	var err error

	var i int8
	for i = 0; i < s.AmountAttempts; i++ {
		response, err = s.Client.Get(url)
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}

	s.LogMutex.Lock()
	defer s.LogMutex.Unlock()

	if err != nil {
		fmt.Fprintf(s.ErrorLog, "Fail to access %s: %s\n", url, err)
		return
	}
	defer response.Body.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(s.StatusLog, "[%s] %s %s\n", timestamp, url, response.Status)
}

func (s *SiteScanner) Run() {
	wg := &sync.WaitGroup{}

	for _, str := range s.Sites {
		wg.Add(1)
		go s.Check(str, wg)
	}

	wg.Wait()
}

func openListOfSites() ([]string, error) {
	sites, err := os.Open("ListOfSites.txt")
	if err != nil {
		return nil, err
	}
	defer sites.Close()

	scanner := bufio.NewScanner(sites)
	arrStr := []string{}

	for scanner.Scan() {
		arrStr = append(arrStr, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return clean(arrStr), nil
}

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	statusLog, err := os.OpenFile("status.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Can not open status.log:", err)
	}
	defer statusLog.Close()

	errorLog, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Can not open errors.log:", err)
	}
	defer errorLog.Close()

	sites, err := openListOfSites()
	if err != nil {
		log.Fatal("Can not open list of sites:", err)
	}

	siteScanner := &SiteScanner{
		Client:         client,
		Sites:          sites,
		StatusLog:      statusLog,
		ErrorLog:       errorLog,
		AmountAttempts: 3,
	}

	siteScanner.Run()
}
