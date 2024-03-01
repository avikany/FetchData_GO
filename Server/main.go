package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	ReadFiles "github.com/avikany/FetchData_GO/Functions"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Index struct {
	// Add fields here based on the data you're expecting
	IndexName string `json:"index_name"`
	// Add more fields as needed
}

func main() {
	// Open file
	// Read file
	filePath := "C:\\Users\\avika\\Downloads\\Compressed\\NFO_symbols.txt"
	lines, err := ReadFiles.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	//fmt.Println("File content:\n", lines)

	location, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(location)

	StTemp := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, location).Unix()
	EtTemp := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, location).Unix()
	st := strconv.FormatInt(StTemp, 10)
	et := strconv.FormatInt(EtTemp, 10)
	uid := "FA69146"
	intrv := "1"
	jKey := "f9f73acda69e5486c4e578f52f70736ab74defd9838784afd52c5bf6dc93ae4a"

	r := gin.Default()
	port := "8080"

	r.POST("/FetchAllOptions", func(c *gin.Context) {
		var Index Index

		if err := c.ShouldBindJSON(&Index); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		for _, line := range lines {
			//fmt.Println(line[0], line[3])
			if line[3] == Index.IndexName {
				//fmt.Println("TRUE")
				exch := line[0]
				token := line[1]
				data := fmt.Sprintf("jData={\"uid\": \"%s\",\"exch\": \"%s\",\"token\": \"%s\",\"st\":"+
					" \"%s\",\"et\": \"%s\",\"intrv\": \"%s\"}&jKey=%s",
					uid, exch, token, st, et, intrv, jKey)
				//fmt.Println(data)
				req, err := http.NewRequest("POST", "https://api.shoonya.com/NorenWClientTP/TPSeries",
					bytes.NewBuffer([]byte(data)))
				if err != nil {
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
				//fmt.Println(string(body))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				bodyStr := string(body) // Convert body to string
				//fmt.Println(bodyStr)

				// Unmarshal JSON body
				var result map[string]interface{}
				json.Unmarshal([]byte(bodyStr), &result)

				// Check if stat is not "Not_Ok"
				if result["stat"] != "Not_Ok" {
					//MAKE A folder and NEW FILE AND SAVE DATA
					today := time.Now().Format("2006-01-02") // Get today's date in YYYY-MM-DD format
					filePath1 := "C:\\Users\\avika\\Downloads\\Compressed\\DATA\\" + Index.IndexName + today
					// Create the new directory if it doesn't exist
					if _, err := os.Stat(filePath1); os.IsNotExist(err) {
						os.Mkdir(filePath1, 0755)
					}
					//dir := filepath.Dir(filePath1)
					outputFilePath := filepath.Join(filePath1, today+line[0]+line[1]+".json")
					err = os.WriteFile(outputFilePath, body, 0644)
					if err != nil {
						fmt.Println("Error writing file:", err)
					}
				}

				// Directly write the response body string as JSON to the response
				c.Data(http.StatusOK, "application/json", []byte(bodyStr))
			}

		}
	})

	err1 := r.Run(":" + port)
	if err1 != nil {
		fmt.Println("Error Running Server,Try Again")
	}
}
