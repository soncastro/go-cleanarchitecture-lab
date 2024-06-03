package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/songomes/desafiocleanarchitecture/configs"
	"github.com/songomes/desafiocleanarchitecture/graph"
	"github.com/songomes/desafiocleanarchitecture/internal/event/handler"
	"github.com/songomes/desafiocleanarchitecture/internal/infra/grpc/service"
	"github.com/songomes/desafiocleanarchitecture/internal/pb"
	"github.com/songomes/desafiocleanarchitecture/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
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

	db, err := sql.Open(configs.DBDriver, configs.DBPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := os.Stat(configs.DBPath); os.IsNotExist(err) {
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS orders (
    			id varchar(255) NOT NULL, 
    			price float NOT NULL, 
    			tax float NOT NULL, 
    			final_price float NOT NULL, 
    			PRIMARY KEY (id)
    		)`); err != nil {
			panic(err)
		}
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		panic(err)
	}

	if count == 0 {
		_, err = db.Exec(`INSERT INTO orders (id, price, tax, final_price) VALUES
			('1', 100.0, 10.0, 110.0),
			('2', 200.0, 20.0, 220.0)`)
		if err != nil {
			panic(err)
		}
		fmt.Println("Registros iniciais inseridos na tabela orders.")
	}

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
	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				return ch
			}
		}
		fmt.Println("RabbitMQ não está pronto, tentando novamente em 5 segundos...")
		time.Sleep(5 * time.Second)
	}
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

//
//func waitForMySQLConnection(dsn string) *sql.DB {
//	var db *sql.DB
//	var err error
//
//	for {
//		db, err = sql.Open("mysql", dsn)
//		if err == nil {
//			err = db.Ping()
//			if err == nil {
//				log.SetOutput(os.Stdout)
//				log.Println("Successfully connected to MySQL")
//				fmt.Println("Successfully connected to MySQL")
//				return db
//			}
//		}
//
//		log.SetOutput(os.Stderr)
//		log.Println("Failed to connect to MySQL, retrying in 5 seconds...")
//		fmt.Println("Failed to connect to MySQL, retrying in 5 seconds...")
//		time.Sleep(5 * time.Second)
//	}
//}
