package product

import "sync"

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
	s.products[1] = Product{ID: 1, Name: "Sample Product", Description: "Seeded item", Price: 9.99}
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
	if incoming.Description != "" {
		existing.Description = incoming.Description
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
		Description: incoming.Description,
		Price:       incoming.Price,
	}
	s.products[id] = created
	return created
}
