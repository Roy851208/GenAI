// main.go
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
)

// Config 结构体用于解析 YAML 文件
type Config struct {
	APIKey string `yaml:"api_key"`
}

func main() {
	// 读取配置文件
	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	router := gin.Default()
	// 设置模板文件的路径
	router.LoadHTMLGlob("templates/*")
	// Define a route for the home page
	router.GET("/", func(c *gin.Context) {
		if c == nil {
			log.Println("Context is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		log.Println("Handling root path")

		// Check if the template exists
		_, err := c.Request.Cookie("Pycharm-4d8bf835") // Assuming this is your session cookie
		if err != nil {
			log.Println("Cookie not found:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Define a route to handle the AI generation request
	router.POST("/generate", func(c *gin.Context) {
		// Get user input from the request body
		var userInput struct {
			Text string `json:"text"`
		}
		if err := c.BindJSON(&userInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-pro")
		resp, err := model.GenerateContent(ctx, genai.Text(userInput.Text))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send the generated content as JSON response
		c.JSON(http.StatusOK, gin.H{"content": resp.Candidates[0].Content.Parts})
	})

	// Run the server on port 8080
	router.Run(":8080")
}

// 读取配置文件
func readConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
