package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/codegangsta/negroni"
	//"github.com/streadway/amqp"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	//"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// MongoDB Config
var mongodb_server = "localhost:27015"
var mongodb_database = "test"
var mongodb_collection = "payments"
// RabbitMQ Config
/*var rabbitmq_server = "rabbitmq"
var rabbitmq_port = "5672"
var rabbitmq_queue = "gumball"
var rabbitmq_user = "guest"
var rabbitmq_pass = "guest"*/

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	n.UseHandler(mx)
	return n
}

func init(){
	//Code to create workers 
	fmt.Println("Started server...")
	Payment_channel=make(chan payment,10)
	go writerWorker()
	//for i:=0; i<4; i++{
		//go workerHandler()
	//} 

	//Create updater worker
}

// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/payment/{order_id}", paymentHandler(formatter)).Methods("GET")
	mx.HandleFunc("/payment/", newPaymentHandler(formatter)).Methods("POST")
}
/*
// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}*/


// API Ping Handler
func paymentHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//formatter.JSON(w, http.StatusOK, struct{ Test string }{"API version 1.0 alive!"})
		// Connects to MongoDB
		session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)
        params := mux.Vars(req)
        var order_id string = params["order_id"]
        fmt.Println("Fetching order#: ",order_id)
        var result bson.M
		err = c.Find(bson.M{"order_id" : order_id}).One(&result)
		if err != nil {
                log.Fatal(mux.Vars(req))
        }
        fmt.Println("Payment made is:", result)
		formatter.JSON(w, http.StatusOK, result)

	}
}

func newPaymentHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var data payment
		err := json.NewDecoder(req.Body).Decode(&data)
		if err!=nil{
			panic(err)
		}
		/*session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)
		c.Insert(data)*/
		go workerHandler(data,Payment_channel)
		formatter.JSON(w, http.StatusOK, data)
	}
}

func workerHandler(data payment,Payment_channel chan payment) {
	Payment_channel<-data
	//Worker tasks
	//Write to channel
}
func writerWorker(){

	for i:=0;;i++{
		payment_value:=<-Payment_channel
		session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)
		c.Insert(payment_value)

	}
}
/*
// API Gumball Machine Handler
func gumballHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)
        var result bson.M
        err = c.Find(bson.M{"SerialNumber" : "1234998871109"}).One(&result)
        if err != nil {
                log.Fatal(err)
        }
        fmt.Println("Gumball Machine:", result )
		formatter.JSON(w, http.StatusOK, result)
	}
}

// API Update Gumball Inventory
func gumballUpdateHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
    	var m gumballMachine
    	_ = json.NewDecoder(req.Body).Decode(&m)		
    	fmt.Println("Update Gumball Inventory To: ", m.CountGumballs)
		session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)
        query := bson.M{"SerialNumber" : "1234998871109"}
        change := bson.M{"$set": bson.M{ "CountGumballs" : m.CountGumballs}}
        err = c.Update(query, change)
        if err != nil {
                log.Fatal(err)
        }
       	var result bson.M
        err = c.Find(bson.M{"SerialNumber" : "1234998871109"}).One(&result)
        if err != nil {
                log.Fatal(err)
        }        
        fmt.Println("Gumball Machine:", result )
		formatter.JSON(w, http.StatusOK, result)
	}
}

// API Create New Gumball Order
func gumballNewOrderHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		uuid := uuid.NewV4()
    	var ord = order {
					Id: uuid.String(),            		
					OrderStatus: "Order Placed",
		}
		if orders == nil {
			orders = make(map[string]order)
		}
		orders[uuid.String()] = ord
		queue_send(uuid.String())
		fmt.Println( "Orders: ", orders )
		formatter.JSON(w, http.StatusOK, ord)
	}
}

// API Get Order Status
func gumballOrderStatusHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		params := mux.Vars(req)
		var uuid string = params["id"]
		fmt.Println( "Order ID: ", uuid )
		if uuid == ""  {
			fmt.Println( "Orders:", orders )
			var orders_array [] order
			for key, value := range orders {
    			fmt.Println("Key:", key, "Value:", value)
    			orders_array = append(orders_array, value)
			}
			formatter.JSON(w, http.StatusOK, orders_array)
		} else {
			var ord = orders[uuid]
			fmt.Println( "Order: ", ord )
			formatter.JSON(w, http.StatusOK, ord)
		}
	}
}

// API Process Orders 
func gumballProcessOrdersHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// Open MongoDB Session
		session, err := mgo.Dial(mongodb_server)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
        c := session.DB(mongodb_database).C(mongodb_collection)

       	// Get Gumball Inventory 
        var result bson.M
        err = c.Find(bson.M{"SerialNumber" : "1234998871109"}).One(&result)
        if err != nil {
                log.Fatal(err)
        }

 		var count int = result["CountGumballs"].(int)
        fmt.Println("Current Inventory:", count )

		// Process Order IDs from Queue
		var order_ids []string = queue_receive()
		for i := 0; i < len(order_ids); i++ {
			var order_id = order_ids[i]
			fmt.Println("Order ID:", order_id)
			var ord = orders[order_id] 
			ord.OrderStatus = "Order Processed"
			orders[order_id] = ord
			count -= 1
		}
		fmt.Println( "Orders: ", orders , "New Inventory: ", count)

		// Update Gumball Inventory
		query := bson.M{"SerialNumber" : "1234998871109"}
        change := bson.M{"$set": bson.M{ "CountGumballs" : count}}
        err = c.Update(query, change)
        if err != nil {
                log.Fatal(err)
        }

		// Return Order Status
		formatter.JSON(w, http.StatusOK, orders)
	}
}

// Send Order to Queue for Processing
func queue_send(message string) {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		rabbitmq_queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := message
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
}

// Receive Order from Queue to Process
func queue_receive() []string {
	conn, err := amqp.Dial("amqp://"+rabbitmq_user+":"+rabbitmq_pass+"@"+rabbitmq_server+":"+rabbitmq_port+"/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		rabbitmq_queue, // name
		false,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"orders",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	order_ids := make(chan string)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			order_ids <- string(d.Body)
		}
		close(order_ids)
	}()

	err = ch.Cancel("orders", false)
	if err != nil {
	    log.Fatalf("basic.cancel: %v", err)
	}

	var order_ids_array []string
	for n := range order_ids {
    	order_ids_array = append(order_ids_array, n)
    }	

    return order_ids_array
}


/*

  	-- Gumball MongoDB Collection (Create Document) --

    db.gumball.insert(
	    { 
	      Id: 1,
	      CountGumballs: NumberInt(202),
	      ModelNumber: 'M102988',
	      SerialNumber: '1234998871109' 
	    }
	) ;

    -- Gumball MongoDB Collection - Find Gumball Document --

    db.gumball.find( { Id: 1 } ) ;

    {
        "_id" : ObjectId("54741c01fa0bd1f1cdf71312"),
        "Id" : 1,
        "CountGumballs" : 202,
        "ModelNumber" : "M102988",
        "SerialNumber" : "1234998871109"
    }

    -- Gumball MongoDB Collection - Update Gumball Document --

    db.gumball.update( 
        { Dd: 1 }, 
        { $set : { CountGumballs : NumberInt(10) } },
        { multi : false } 
    )

    -- Gumball Delete Documents

    db.gumball.remove({})

 */
