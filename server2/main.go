package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type NewsArticle struct {
	Header           string
	Subheader        string
	ImageLink        string
	ImageSubText     string
	ContentParagraph string
}

var newsArticles []NewsArticle

func fetchNews() {
	res, err := http.Get("https://www.moneycontrol.com/markets/fno-market-snapshot")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	baseURL := "https://www.moneycontrol.com/"

	var wg sync.WaitGroup

	for i := 1; i <= 11; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println("NEWS : ", i)
			selector := fmt.Sprintf("#news_tab > p:nth-child(%d)", i)
			element := doc.Find(selector).First()

			link, _ := element.Find("a").Attr("href")

			res, err := http.Get(baseURL + link)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}
			docPage, err := goquery.NewDocumentFromReader(res.Body)
			MainHeader := docPage.Find("#page1 > div > h1").Text()
			SubHeader := docPage.Find("#page1 > div > h2").Text()
			ImageLink, _ := docPage.Find("#contentdata > div.article_image_wrapper.article_image_main_wrapper > div > img").Attr("data-src")
			ImageSubText := docPage.Find("#contentdata > div.article_image_wrapper.article_image_main_wrapper > h2").Text()
			var ContentPara1 string
			docPage.Find("#contentdata > p").Each(func(i int, s *goquery.Selection) {

				text := s.Text()

				// List of strings to check
				var checkStrings = []string{
					"Follow our live blog for all the market action",
					"Also read",
					"ALSO READ",
					"Also Read",
					"Follow our market blog to catch all the live updates",
					"Disclaimer",
				}

				shouldAdd := true
				for _, str := range checkStrings {
					if strings.HasPrefix(text, str) {
						shouldAdd = false
						break
					}
				}

				if shouldAdd {
					ContentPara1 += text
				}

			})

			newsArticle := NewsArticle{
				Header:           MainHeader,
				Subheader:        SubHeader,
				ImageLink:        ImageLink,
				ImageSubText:     ImageSubText,
				ContentParagraph: ContentPara1,
			}

			mu.Lock()
			newsArticles = append(newsArticles, newsArticle)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
}

type FOStats struct {
	CompanyName   string
	CMP           float64
	Change        float64
	PercentChange float64
}

var StatsData map[string][]FOStats
var mu sync.Mutex

func fetchFOStats() {
	fmt.Println("Fetching Stats Data : ")
	res, err := http.Get("https://www.moneycontrol.com/markets/fno-market-snapshot")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var fOStatsGainers []FOStats
	var fOStatsLosers []FOStats

	for i := 2; i <= 8; i++ {
		// Fetch Gainers
		gainerSelector := fmt.Sprintf("#data_1_1 > div > table > tbody > tr:nth-child(%d) > td:nth-child(1)", i)
		CompanyName := doc.Find(gainerSelector).First().Text()
		var tempGainers = make([]float64, 3)
		for j := 2; j <= 4; j++ {
			selector := fmt.Sprintf("#data_1_1 > div > table > tbody > tr:nth-child(%d) > td:nth-child(%d)", i, j)
			element := doc.Find(selector).First().Text()

			// Remove commas and percentage sign from the string
			element = strings.ReplaceAll(element, ",", "")
			element = strings.ReplaceAll(element, "%", "")

			// Convert string to float64
			floatElement, err := strconv.ParseFloat(element, 64)
			if err != nil {
				fmt.Println(err)
			} else {
				tempGainers[j-2] = float64(floatElement)
			}
		}

		fOStatsGainer := FOStats{
			CompanyName:   CompanyName,
			CMP:           tempGainers[0],
			Change:        tempGainers[1],
			PercentChange: tempGainers[2],
		}
		fOStatsGainers = append(fOStatsGainers, fOStatsGainer)

		// Fetch Losers
		loserSelector := fmt.Sprintf("#data_1_2 > div > table > tbody > tr:nth-child(%d) > td:nth-child(1)", i)
		CompanyName = doc.Find(loserSelector).First().Text()
		var tempLosers = make([]float64, 3)
		for j := 2; j <= 4; j++ {
			selector := fmt.Sprintf("#data_1_2 > div > table > tbody > tr:nth-child(%d) > td:nth-child(%d)", i, j)
			element := doc.Find(selector).First().Text()

			// Remove commas and percentage sign from the string
			element = strings.ReplaceAll(element, ",", "")
			element = strings.ReplaceAll(element, "%", "")

			// Convert string to float64
			floatElement, err := strconv.ParseFloat(element, 64)
			if err != nil {
				fmt.Println(err)
			} else {
				tempLosers[j-2] = float64(floatElement)
			}
		}

		fOStatsLoser := FOStats{
			CompanyName:   CompanyName,
			CMP:           tempLosers[0],
			Change:        tempLosers[1],
			PercentChange: tempLosers[2],
		}
		fOStatsLosers = append(fOStatsLosers, fOStatsLoser)
	}

	mu.Lock()
	StatsData = map[string][]FOStats{
		"Gainers": fOStatsGainers,
		"Losers":  fOStatsLosers,
	}
	mu.Unlock()
}
func main() {
	r := gin.Default()

	// Fetch news immediately and then every 10 minutes
	go func() {
		fetchNews()
		for range time.Tick(10 * time.Minute) {
			fetchNews()
		}
	}()

	// Fetch F&O stats immediately and then every 10 minutes

	r.GET("/collect/news/moneycontrol", func(c *gin.Context) {
		mu.Lock()
		c.JSON(http.StatusOK, newsArticles)
		mu.Unlock()
	})

	r.GET("/collect/news/moneycontrol/F&O", func(c *gin.Context) {
		fetchFOStats()
		c.JSON(http.StatusOK, StatsData)
	})

	r.Run()
}
