package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/songomes/desafiocleanarchitecture/configs"
	"github.com/songomes/desafiocleanarchitecture/graph"
	"github.com/songomes/desafiocleanarchitecture/internal/event/handler"
	"github.com/songomes/desafiocleanarchitecture/internal/infra/grpc/service"
	"github.com/songomes/desafiocleanarchitecture/internal/pb"
	"github.com/songomes/desafiocleanarchitecture/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

type HandlerMain struct {
	DB              *sql.DB
	EventDispatcher *events.EventDispatcher
}

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rabbitMQChannel := getRabbitMQChannel()

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	eventDispatcher2 := events.NewEventDispatcher()
	eventDispatcher2.Register("GetAllOrdersFetched", &handler.GetAllOrdersFetchedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	getAllOrdersUseCase := NewGetAllOrdersUseCase(db, eventDispatcher2)

	hdlMain := &HandlerMain{
		DB:              db,
		EventDispatcher: eventDispatcher2,
	}

	fmt.Println("Starting web server on port", configs.WebServerPort)
	r := mux.NewRouter()
	r.HandleFunc("/order", hdlMain.ListOrdersREST).Methods("GET")

	go func() {
		if err := http.ListenAndServe(configs.WebServerPort, r); err != nil {
			panic(err)
		}
	}()

	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(*createOrderUseCase, *getAllOrdersUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", configs.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase:  *createOrderUseCase,
		GetAllOrdersUseCase: *getAllOrdersUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", configs.GraphQLServerPort)
	http.ListenAndServe(":"+configs.GraphQLServerPort, nil)
}

func getRabbitMQChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}

func (h *HandlerMain) ListOrdersREST(w http.ResponseWriter, r *http.Request) {
	getAllOrdersUseCase := NewGetAllOrdersUseCase(h.DB, h.EventDispatcher)
	orders, err := getAllOrdersUseCase.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ordersResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
