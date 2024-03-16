package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	ReadFiles "github.com/avikany/FetchData_GO/Functions"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Index struct {
	// Add fields here based on the data you're expecting
	IndexName string `json:"index_name"`
	DATE      string `json:"date"`
	// Add more fields as needed
}

func main() {
	r := gin.Default()
	port := "8080"

	// Open file and Read File
	filePath := "C:\\Users\\avika\\Downloads\\Compressed\\NFO_symbols.txt\\NFO_symbols.txt"

	lines, err := ReadFiles.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	uid := "FA69146"
	intrv := "1"
	jKey := "41522a23e255d4aeed9c9b6c10daa8fbf4e63540b6c30341be0fc8429b7b356b"

	r.POST("/FetchAllOptions", func(c *gin.Context) {
		var Index Index

		if err := c.ShouldBindJSON(&Index); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Reading Every line of the Opened Document
		for _, line := range lines {
			//Since at index is the name of the Option
			if line[3] == Index.IndexName {

				//Parsing Time and converting it into string due to the requirement of Shoonya API
				// date should be in the format mm-dd-year
				DATE := Index.DATE
				location, _ := time.LoadLocation("Asia/Kolkata")
				temp, _ := time.ParseInLocation("01-02-2006", DATE, location)

				StTemp := time.Date(temp.Year(), temp.Month(), temp.Day(), 9, 15, 0, 0, location).Unix()
				EtTemp := time.Date(temp.Year(), temp.Month(), temp.Day(), 15, 30, 0, 0, location).Unix()
				st := strconv.FormatInt(StTemp, 10)
				et := strconv.FormatInt(EtTemp, 10)

				exch := line[0]
				token := line[1]
				//Preparing data to send to API
				data := fmt.Sprintf("jData={\"uid\": \"%s\",\"exch\": \"%s\",\"token\": \"%s\",\"st\":"+
					" \"%s\",\"et\": \"%s\",\"intrv\": \"%s\"}&jKey=%s",
					uid, exch, token, st, et, intrv, jKey)

				req, err := http.NewRequest("POST", "https://api.shoonya.com/NorenWClientTP/TPSeries",
					bytes.NewBuffer([]byte(data)))
				if err != nil {
					fmt.Println(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				bodyStr := string(body) // Convert body to string

				// Unmarshal JSON body
				var result map[string]interface{}
				json.Unmarshal([]byte(bodyStr), &result)

				//MAKE A folder and NEW FILE AND SAVE DATA
				filePathfile := "C:\\Users\\avika\\Downloads\\Compressed\\DATA\\" + Index.IndexName + "\\" + line[4] +
					"\\" + Index.DATE

				// If correct data has returned

				if result["stat"] != "Not_Ok" {
					fmt.Println("Received Correct Data")
					// Create the new directory if it doesn't exist
					var result1 []map[string]interface{}
					err := json.Unmarshal([]byte(bodyStr), &result1)
					if err != nil {
						log.Fatalln("Error In Unmarshalling Data")
					}
					if _, err := os.Stat(filePathfile); os.IsNotExist(err) {
						err := os.MkdirAll(filePathfile, 0755)
						if err != nil {
							log.Fatalf("Error Creating Directory")
						}
					}

					// Create a new CSV file and make a new writer

					file, err := os.Create(filepath.Join(filePathfile, line[4]+".csv"))
					if err != nil {
						log.Fatalf("Failed to create file: %s", err)
					}
					defer file.Close()

					writer := csv.NewWriter(file)
					defer writer.Flush()

					// Write header to CSV file
					header := []string{"intc", "inth", "intl", "into", "intoi", "intv", "intvwap", "oi", "ssboe", "stat", "time", "v"}
					if err := writer.Write(header); err != nil {
						log.Fatalf("Failed to write to file: %s", err)
					}

					//Write data to CSV file
					for _, result := range result1 {
						var row []string
						for _, h := range header {
							// Convert Time to Unix that was received from API
							if h == "time" {
								t, err := time.ParseInLocation("02-01-2006 15:04:05", fmt.Sprintf("%v", result[h]), location)
								if err != nil {
									log.Fatalf("Failed to parse time: %s", err)
								}

								unixTime := t.Unix()
								row = append(row, fmt.Sprintf("%v", unixTime))
							} else {
								row = append(row, fmt.Sprintf("%v", result[h]))
							}
						}
						if err := writer.Write(row); err != nil {
							log.Fatalf("Failed to write to file: %s", err)
						}

					}

					//Create a log file in the same directory
					logFilePath := filepath.Join(filePathfile, "log.txt")
					logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						log.Fatalf("error opening file: %v", err)
					}
					defer logFile.Close()

					// Set the output of the logger to the file
					log.SetOutput(logFile)

					//Now use log.Println to write to the log file
					log.Println("File Written SuccessFully!")

				}
				//If the data returned is not ok

				if result["stat"] == "Not_Ok" {
					fmt.Println("Failed Fetching DATA")
					// Create the new directory if it doesn't exist

					if _, err := os.Stat(filePathfile); os.IsNotExist(err) {
						err := os.MkdirAll(filePathfile, 0755)
						if err != nil {
							log.Fatalf("Error forming directory")
						}
					}

					outputFilePath := filepath.Join(filePathfile, line[4]+".json")

					err = os.WriteFile(outputFilePath, body, 0644)
					if err != nil {
						fmt.Println("Error writing file:", err)
					}

					//Create a log file in the same directory
					logFilePath := filepath.Join(filePathfile, "log.txt")
					logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						log.Fatalf("error opening file: %v", err)
					}
					defer logFile.Close()

					// Set the output of the logger to the file
					log.SetOutput(logFile)

					//Now use log.Println to write to the log file
					log.Println(string(body))

					log.Println(line)
					log.Println(st, et)
				}
				//Directly write the response body string as JSON to the response
				c.Data(http.StatusOK, "application/json", []byte(bodyStr))
			}

		}
	})

	err1 := r.Run(":" + port)
	if err1 != nil {
		fmt.Println("Error Running Server,Try Again")
	}
}
