//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/songomes/desafiocleanarchitecture/internal/entity"
	"github.com/songomes/desafiocleanarchitecture/internal/event"
	"github.com/songomes/desafiocleanarchitecture/internal/infra/database"
	"github.com/songomes/desafiocleanarchitecture/internal/infra/web"
	"github.com/songomes/desafiocleanarchitecture/internal/usecase"
	"github.com/songomes/desafiocleanarchitecture/pkg/events"
)

var setOrderRepositoryDependency = wire.NewSet(
	database.NewOrderRepository,
	wire.Bind(new(entity.OrderRepositoryInterface), new(*database.OrderRepository)),
)

var setEventDispatcherDependency = wire.NewSet(
	events.NewEventDispatcher,
	event.NewOrderCreated,
	wire.Bind(new(events.EventInterface), new(*event.OrderCreated)),
	wire.Bind(new(events.EventDispatcherInterface), new(*events.EventDispatcher)),
)

var setOrderCreatedEvent = wire.NewSet(
	event.NewOrderCreated,
	wire.Bind(new(events.EventInterface), new(*event.OrderCreated)),
)

func NewCreateOrderUseCase(db *sql.DB, eventDispatcher events.EventDispatcherInterface) *usecase.CreateOrderUseCase {
	wire.Build(
		setOrderRepositoryDependency,
		setOrderCreatedEvent,
		usecase.NewCreateOrderUseCase,
	)
	return &usecase.CreateOrderUseCase{}
}

func NewGetAllOrdersUseCase(db *sql.DB, eventDispatcher events.EventDispatcherInterface) *usecase.GetAllOrdersUseCase {
	wire.Build(
		setOrderRepositoryDependency,
		setGetAllOrdersEvent,
		usecase.NewGetAllOrdersUseCase,
	)
	return &usecase.GetAllOrdersUseCase{}
}

func NewWebOrderHandler(db *sql.DB, eventDispatcher events.EventDispatcherInterface) *web.WebOrderHandler {
	wire.Build(
		setOrderRepositoryDependency,
		setOrderCreatedEvent,
		web.NewWebOrderHandler,
	)
	return &web.WebOrderHandler{}
}
