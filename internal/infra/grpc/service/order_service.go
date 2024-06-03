package service

import (
	"context"
	"github.com/songomes/desafiocleanarchitecture/internal/pb"
	"github.com/songomes/desafiocleanarchitecture/internal/usecase"
	"strconv"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
	CreateOrderUseCase  usecase.CreateOrderUseCase
	GetAllOrdersUseCase usecase.GetAllOrdersUseCase
}

func NewOrderService(createOrderUseCase usecase.CreateOrderUseCase, getAllOrdersUseCase usecase.GetAllOrdersUseCase) *OrderService {
	return &OrderService{
		CreateOrderUseCase:  createOrderUseCase,
		GetAllOrdersUseCase: getAllOrdersUseCase,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	dto := usecase.OrderInputDTO{
		ID:    in.Id,
		Price: float64(in.Price),
		Tax:   float64(in.Tax),
	}
	output, err := s.CreateOrderUseCase.Execute(dto)
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrderResponse{
		Id:         output.ID,
		Price:      float32(output.Price),
		Tax:        float32(output.Tax),
		FinalPrice: float32(output.FinalPrice),
	}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, in *pb.Blank) (*pb.OrderList, error) {

	orders, err := s.GetAllOrdersUseCase.Execute()
	if err != nil {
		return nil, err
	}

	var ordersResponse []*pb.Order

	for _, order := range orders {
		orderResponse := &pb.Order{
			Id:         order.ID,
			Price:      strconv.FormatFloat(order.Price, 'f', -1, 64),
			Tax:        strconv.FormatFloat(order.Tax, 'f', -1, 64),
			FinalPrice: strconv.FormatFloat(order.FinalPrice, 'f', -1, 64),
		}

		ordersResponse = append(ordersResponse, orderResponse)
	}

	return &pb.OrderList{Orders: ordersResponse}, nil
}
