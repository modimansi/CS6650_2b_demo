package cart

// CartStore defines the interface for shopping cart storage
// Supports both PostgreSQL (int IDs) and DynamoDB (string UUIDs)
type CartStore interface {
	CreateCart(customerID int) (*ShoppingCart, error)
	GetCart(cartID interface{}) (*ShoppingCart, error)
	GetCartWithItems(cartID interface{}) (*CartWithItems, error)
	AddOrUpdateItem(cartID interface{}, productID int, quantity int) error
	CheckoutCart(cartID interface{}) (interface{}, error)
	Close() error
	InitSchema() error
}
