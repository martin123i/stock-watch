package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"stocktracker.com/app/internal/auth"
	"stocktracker.com/app/internal/db"
	"stocktracker.com/app/internal/middleware"
	"stocktracker.com/app/internal/portfolio"
)

// StockDetails holds additional stock details
type StockDetails struct {
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Open   float64 `json:"o"`
	Volume int     `json:"v"`
}

const API_KEY = "csido9hr01qt46e7sfegcsido9hr01qt46e7sff0" // Replace with your actual Finnhub API key

// Define the symbols you want to fetch
var STOCK_SYMBOLS = []string{"AAPL", "GOOGL", "MSFT", "AMZN", "FB", "TSLA", "NVDA", "INTC", "ADBE", "PYPL"}

type StockPrice struct {
	C float64 `json:"c"`  // Current price
	P float64 `json:"pc"` // Previous close price
}

func fetchStockPrices(symbol string) (string, string, error) {
	// Finnhub endpoint for getting the current stock price
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", symbol, API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("error fetching data from Finnhub: %s", resp.Status)
	}

	var stockPrice StockPrice
	if err := json.NewDecoder(resp.Body).Decode(&stockPrice); err != nil {
		return "", "", err
	}

	// Format current and previous prices as strings
	currentPrice := fmt.Sprintf("%.2f", stockPrice.C)
	previousPrice := fmt.Sprintf("%.2f", stockPrice.P)

	return currentPrice, previousPrice, nil
}

// Fetches detailed stock information for a given symbol
func fetchStockDetails(symbol string) (StockDetails, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", symbol, API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		return StockDetails{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StockDetails{}, fmt.Errorf("error fetching data from Finnhub: %s", resp.Status)
	}

	var stockDetails StockDetails
	if err := json.NewDecoder(resp.Body).Decode(&stockDetails); err != nil {
		return StockDetails{}, err
	}

	return stockDetails, nil
}

// New endpoint to get additional stock details for a specific symbol
func stockDetailsHandler(c *gin.Context) {
	symbol := c.Param("symbol")
	details, err := fetchStockDetails(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch stock details"})
		return
	}
	c.JSON(http.StatusOK, details)
}

func stockPriceHandler(c *gin.Context) {
	results := make(map[string]map[string]string)
	for _, symbol := range STOCK_SYMBOLS {
		currentPrice, previousPrice, err := fetchStockPrices(symbol)
		if err != nil {
			results[symbol] = map[string]string{"error": "Unable to fetch data"}
		} else {
			results[symbol] = map[string]string{"current": currentPrice, "previous": previousPrice}
		}
	}
	c.JSON(http.StatusOK, results)
}

func main() {
	db.Init()
	r := gin.Default()

	// Enable CORS for your frontend origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Update if your frontend is on a different port
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/api/stocks", stockPriceHandler)
	r.GET("/api/stock/:symbol", stockDetailsHandler)
	r.POST("/register", auth.Register)
	r.POST("/login", auth.Login)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/add-favorite", portfolio.AddFavorite)
		protected.GET("/favorites", portfolio.GetFavorites)
	}

	fmt.Println("Server started at http://localhost:8080")
	r.Run(":8080")
}
