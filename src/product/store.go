package product

import (
	"fmt"
	"strings"
	"sync"
)

// Store provides concurrent-safe in-memory storage for products.
type Store struct {
	mu       sync.RWMutex
	products map[int32]Product
	nextID   int32
}

func NewStore() *Store {
	return &Store{products: make(map[int32]Product), nextID: 1}
}

func (s *Store) SeedSample() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.products[1] = Product{ID: 1, Name: "Sample Product", Category: "Electronics", Description: "Seeded item", Brand: "Acme", Price: 9.99}
	if s.nextID <= 1 {
		s.nextID = 2
	}
}

func (s *Store) Get(id int32) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	return p, ok
}

// List returns all products filtered by optional name and category substrings (case-insensitive).
func (s *Store) List(nameFilter, categoryFilter string) []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Product
	name := strings.ToLower(nameFilter)
	category := strings.ToLower(categoryFilter)
	for _, p := range s.products {
		if name != "" {
			if !strings.Contains(strings.ToLower(p.Name), name) {
				continue
			}
		}
		if category != "" {
			if !strings.Contains(strings.ToLower(p.Category), category) {
				continue
			}
		}
		results = append(results, p)
	}
	return results
}

func (s *Store) UpdateDetails(id int32, incoming Product) (Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.products[id]
	if !ok {
		return Product{}, false
	}
	incoming.ID = id
	if incoming.Name != "" {
		existing.Name = incoming.Name
	}
	if incoming.Category != "" {
		existing.Category = incoming.Category
	}
	if incoming.Description != "" {
		existing.Description = incoming.Description
	}
	if incoming.Brand != "" {
		existing.Brand = incoming.Brand
	}
	if incoming.Price != 0 {
		existing.Price = incoming.Price
	}
	s.products[id] = existing
	return existing, true
}

func (s *Store) Create(incoming Product) Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nextID == 0 {
		s.nextID = 1
	}
	id := s.nextID
	s.nextID++
	created := Product{
		ID:          id,
		Name:        incoming.Name,
		Category:    incoming.Category,
		Description: incoming.Description,
		Brand:       incoming.Brand,
		Price:       incoming.Price,
	}
	s.products[id] = created
	return created
}

// SeedBulk deterministically generates N products with rotating brands and categories.
// Names follow the pattern "Product [Brand] [ID]" to ensure consistent search behavior.
func (s *Store) SeedBulk(n int) {
	if n <= 0 {
		return
	}

	categories := []string{"Electronics", "Books", "Home", "Toys", "Clothing", "Sports", "Garden", "Beauty", "Automotive", "Grocery"}
	brands := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Omega"}

	s.mu.Lock()
	// Recreate map with a capacity hint for performance during bulk load
	s.products = make(map[int32]Product, n)
	for i := 1; i <= n; i++ {
		id := int32(i)
		brand := brands[(i-1)%len(brands)]
		category := categories[(i-1)%len(categories)]
		name := fmt.Sprintf("Product %s %d", brand, i)
		description := fmt.Sprintf("Description for %s", name)
		// Deterministic price pattern in range ~1.00 - 110.99
		price := float64((i%110)+1) + float64(i%100)/100.0

		s.products[id] = Product{
			ID:          id,
			Name:        name,
			Category:    category,
			Description: description,
			Brand:       brand,
			Price:       price,
		}
	}
	s.nextID = int32(n) + 1
	s.mu.Unlock()
}

// SearchLimited scans up to maxCheck products and returns up to maxReturn matches,
// along with the total number of matches found among the scanned products.
// Matching is case-insensitive on name and category substrings. Empty filters match all.
func (s *Store) SearchLimited(nameFilter, categoryFilter string, maxCheck, maxReturn int) ([]Product, int) {
	if maxCheck <= 0 {
		return nil, 0
	}
	if maxReturn < 0 {
		maxReturn = 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	lowerName := strings.ToLower(nameFilter)
	lowerCategory := strings.ToLower(categoryFilter)

	results := make([]Product, 0, maxReturn)
	checked := 0
	totalFound := 0

	for _, p := range s.products {
		if checked >= maxCheck {
			break
		}
		checked++ // increment for EVERY product checked

		matches := true
		if lowerName != "" {
			if !strings.Contains(strings.ToLower(p.Name), lowerName) {
				matches = false
			}
		}
		if lowerCategory != "" {
			if !strings.Contains(strings.ToLower(p.Category), lowerCategory) {
				matches = false
			}
		}
		if matches {
			totalFound++
			if len(results) < maxReturn {
				results = append(results, p)
			}
		}
	}

	return results, totalFound
}
