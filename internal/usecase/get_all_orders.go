package usecase

import (
	"github.com/songomes/desafiocleanarchitecture/internal/entity"
	"github.com/songomes/desafiocleanarchitecture/pkg/events"
)

type OrderListOutputDTO struct {
	ID         string  `json:"id"`
	Price      float64 `json:"price"`
	Tax        float64 `json:"tax"`
	FinalPrice float64 `json:"final_price"`
}

type GetAllOrdersUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
	OrderCreated    events.EventInterface
	EventDispatcher events.EventDispatcherInterface
}

func NewGetAllOrdersUseCase(
	OrderRepository entity.OrderRepositoryInterface,
	OrderCreated events.EventInterface,
	EventDispatcher events.EventDispatcherInterface,
) *GetAllOrdersUseCase {
	return &GetAllOrdersUseCase{
		OrderRepository: OrderRepository,
		OrderCreated:    OrderCreated,
		EventDispatcher: EventDispatcher,
	}
}

func (c *GetAllOrdersUseCase) Execute() ([]OrderListOutputDTO, error) {
	orders, err := c.OrderRepository.GetAllOrders()
	if err != nil {
		return nil, err
	}

	var dto []OrderListOutputDTO
	for _, order := range orders {
		dto = append(dto, OrderListOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		})
	}

	return dto, nil
}
