package entity

type OrderRepositoryInterface interface {
	Save(order *Order) error
	GetAllOrders() ([]*Order, error)
	// GetTotal() (int, error)
}
