package main

import (
	"crypto/subtle"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"sync"
	"unicode/utf8"
)

type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

// A pointer distinguishes a required zero weight from an omitted/null weight.
type productInput struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       *int32 `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type productStore struct {
	mu       sync.RWMutex
	products map[int32]Product
}

func fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, APIError{Error: code, Message: message})
}

func productID(c *gin.Context, invalidStatus int) (int32, bool) {
	id, err := strconv.ParseInt(c.Param("productId"), 10, 32)
	if err != nil || id < 1 {
		code := "INVALID_INPUT"
		if invalidStatus == http.StatusNotFound {
			code = "NOT_FOUND"
		}
		fail(c, invalidStatus, code, "Product ID must be a positive 32-bit integer")
		return 0, false
	}
	return int32(id), true
}

func (s *productStore) getProduct(c *gin.Context) {
	id, ok := productID(c, http.StatusNotFound)
	if !ok {
		return
	}
	s.mu.RLock()
	product, exists := s.products[id]
	s.mu.RUnlock()
	if !exists {
		fail(c, 404, "NOT_FOUND", "Product not found")
		return
	}
	c.JSON(200, product)
}

func (s *productStore) updateProduct(c *gin.Context) {
	id, ok := productID(c, 400)
	if !ok {
		return
	}
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		fail(c, 400, "INVALID_INPUT", "Content-Type must be application/json")
		return
	}
	var input productInput
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&input); err != nil {
		fail(c, 400, "INVALID_INPUT", "Body must be a JSON Product with 32-bit integer fields")
		return
	}
	// Reject additional JSON values or garbage, but accept trailing whitespace.
	if err := decoder.Decode(new(any)); err != io.EOF {
		fail(c, 400, "INVALID_INPUT", "Body must contain exactly one JSON object")
		return
	}
	if input.ProductID < 1 || input.CategoryID < 1 || input.SomeOtherID < 1 ||
		input.Weight == nil || *input.Weight < 0 ||
		utf8.RuneCountInString(input.SKU) < 1 || utf8.RuneCountInString(input.SKU) > 100 ||
		utf8.RuneCountInString(input.Manufacturer) < 1 || utf8.RuneCountInString(input.Manufacturer) > 200 {
		fail(c, 400, "INVALID_INPUT", "All six fields are required: IDs >= 1, weight >= 0, sku length 1-100, manufacturer length 1-200")
		return
	}
	if input.ProductID != id {
		fail(c, 400, "INVALID_INPUT", "Body product_id must match the URL productId")
		return
	}
	product := Product{input.ProductID, input.SKU, input.Manufacturer, input.CategoryID, *input.Weight, input.SomeOtherID}
	s.mu.Lock()
	_, exists := s.products[id]
	if exists {
		s.products[id] = product
	}
	s.mu.Unlock()
	if !exists {
		fail(c, 404, "NOT_FOUND", "Product not found")
		return
	}
	c.Status(http.StatusNoContent) // 204 must have no response body.
}

func newRouter(apiKey string, enableTestErrors bool) *gin.Engine {
	store := &productStore{products: map[int32]Product{
		1: {ProductID: 1, SKU: "ABC-123", Manufacturer: "Acme", CategoryID: 10, Weight: 1250, SomeOtherID: 20},
	}}
	router := gin.New()
	router.Use(gin.Logger())
	// Keep panic responses consistent with the Error schema.
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		fail(c, 500, "INTERNAL_ERROR", "Internal server error")
	}))
	api := router.Group("/v1")
	api.Use(func(c *gin.Context) {
		if apiKey == "" || subtle.ConstantTimeCompare([]byte(c.GetHeader("X-API-Key")), []byte(apiKey)) != 1 {
			fail(c, 401, "UNAUTHORIZED", "A valid X-API-Key is required")
			return
		}
		c.Next()
	})
	// Disabled by default; exercises actual panic recovery for local demonstrations.
	fault := func(c *gin.Context) {
		if enableTestErrors && c.GetHeader("X-Test-Error") == "500" {
			panic("simulated internal failure")
		}
		c.Next()
	}
	api.GET("/products/:productId", fault, store.getProduct)
	api.POST("/products/:productId/details", fault, store.updateProduct)
	return router
}

func main() {
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatal("Set API_KEY before starting the server")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(newRouter(key, os.Getenv("ENABLE_TEST_ERRORS") == "true").Run(":" + port))
}
