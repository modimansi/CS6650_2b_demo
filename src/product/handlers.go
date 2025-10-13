package product

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// GET /products
func (h *Handlers) ListProducts(c *gin.Context) {
	name := c.Query("name")
	category := c.Query("category")
	const maxCheck = 100
	const maxReturn = 20

	start := time.Now()
	products, total := h.store.SearchLimited(name, category, maxCheck, maxReturn)
	elapsed := time.Since(start)

	resp := SearchResponse{
		Products:   products,
		TotalFound: total,
		SearchTime: elapsed.String(),
	}
	c.JSON(http.StatusOK, resp)
}

// POST /products
func (h *Handlers) CreateProduct(c *gin.Context) {
	var body Product
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid JSON body"})
		return
	}
	if body.Price < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "price must be non-negative"})
		return
	}
	if len(body.Name) == 0 || len(body.Name) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid name"})
		return
	}

	created := h.store.Create(body)
	c.JSON(http.StatusCreated, created)
}

// GET /products/{productId}
func (h *Handlers) GetProduct(c *gin.Context) {
	id, ok := parseProductID(c.Param("productId"))
	if !ok || id < 1 {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "product not found"})
		return
	}
	product, found := h.store.Get(id)
	if !found {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// POST /products/{productId}/details
func (h *Handlers) AddProductDetails(c *gin.Context) {
	id, ok := parseProductID(c.Param("productId"))
	if !ok || id < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid productId"})
		return
	}

	var body Product
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid JSON body"})
		return
	}
	if body.Price < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "price must be non-negative"})
		return
	}
	if len(body.Name) > 0 && len(body.Name) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "name is too long"})
		return
	}

	if _, exists := h.store.Get(id); !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "product not found"})
		return
	}

	if _, ok := h.store.UpdateDetails(id, body); !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseProductID(raw string) (int32, bool) {
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(v), true
}
