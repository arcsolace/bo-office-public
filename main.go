package main

import (
	"fmt"
  "context"
  "log"
  "net/http"
  "encoding/json"
  "github.com/gorilla/mux"
  "os"
	"github.com/joho/godotenv"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ConnectDB() *mongo.Collection {
  var config = GetConfiguration()
	clientOptions := options.Client().ApplyURI(config.ConnectionString)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	collection := client.Database("db_intern").Collection("user_data")

	return collection
}
//////////////////////////////////////////////////////////////////////

// Error response structure
type ErrorResponse struct {
	StatusCode   int    `json:"status"`
	ErrorMessage string `json:"message"`
}

// Error model function
func GetError(err error, w http.ResponseWriter) {

	log.Fatal(err.Error())
	var response = ErrorResponse {
		ErrorMessage: err.Error(),
		StatusCode:   http.StatusInternalServerError,
	}

	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
	w.Write(message)
}

// Configuration model
type Configuration struct {
	Port             string
	ConnectionString string
}

// Get configuration from .env
func GetConfiguration() Configuration {
	err := godotenv.Load("./.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	configuration := Configuration{
		os.Getenv("PORT"),
		os.Getenv("CONNECTION_STRING"),
	}

	return configuration
}

var collection = ConnectDB()

type User struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	BNID   string             `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Fname  string             `json:"fullname" bson:"fullname,omitempty"`
	Nname  string             `json:"nickname" bson:"nickname,omitempty"`
}

func getIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	// array of all users
	var users []User
	// get all data from collection first
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		GetError(err, w)
		return
	}
  // Defers cursor closure until surrounding function is finished.
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var user User // get one user
		err := cur.Decode(&user) // deserialize
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(users) // reserialize
}

func getId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// set header.from request
	w.Header().Set("Content-Type", "application/json")
	var user User
	// get parameters from mux
	var params = mux.Vars(r)
	// string to primitive.ObjectID
  user_id := params["user_id"]
	err := collection.FindOne(context.TODO(), bson.M{"user_id": user_id}).Decode(&user)
	if err != nil {
		GetError(err, w)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func main() {
	//Init mux router
	var r = mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte("{\"hello\": \"world\"}"))
    })
  // routes
	r.HandleFunc("/bo", getIds).Methods("GET", "OPTIONS")
	r.HandleFunc("/bo/{user_id}", getId).Methods("GET", "OPTIONS")

  // get port address from helper
  var config = GetConfiguration()
	log.Fatal(http.ListenAndServe(config.Port, r))
}
