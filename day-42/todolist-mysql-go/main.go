package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type task struct {
	Taskname string
	Complete bool
}
type Person struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`

	Name        string
	Age         int
	Description string
	Work        []task
	CreatedAT   time.Time
	UpdatedAt   time.Time
}

var client *mongo.Client
var err error
var person Person
var episodes []bson.M

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	dec := json.NewDecoder(r.Body).Decode(&person)
	if dec == nil {
		w.WriteHeader(500)
		return
	} else if len(person.Name) <= 0 || len(person.Description) <= 0 || person.Age == 0 || person.Age < 0 {
		fmt.Println("Please Provide specified field")
		w.WriteHeader(400)
		return
	}
	collection := client.Database("<dbname>").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.Find(ctx, bson.M{"name": person.Name})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if err = result.All(ctx, &episodes); err != nil {
		w.WriteHeader(400)
		return
	}
	if episodes == nil {
		res, _ := collection.InsertOne(ctx, person)
		json.NewEncoder(w).Encode(res)
		return
	}

}

func fetchall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	collection := client.Database("<dbname>").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Find(ctx, bson.M{})
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}

	if err = result.All(ctx, &episodes); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
	}
	if episodes == nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("No Data Found")
		return
	}
	json.NewEncoder(w).Encode(episodes)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(405)
		return
	}
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println(id)
	collection := client.Database("<dbname>").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	fmt.Println(person)
	fmt.Println(err)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(result)
}

func getUserTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println(id)
	collection := client.Database("<dbname>").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Find(ctx, bson.M{"_id": id})
	fmt.Println(result)
	fmt.Println(err)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if err = result.All(ctx, &episodes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(episodes)
	json.NewEncoder(w).Encode(episodes)

}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println(id)
	collection := client.Database("<dbname>").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.D{{"$set", bson.D{{"name", "SHIZUKA"}}}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	json.NewEncoder(w).Encode(result)
}

func main() {
	log.Info("starting server")
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://Ankita:Ank%401297@cluster0-evm4q.mongodb.net/<dbname>?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)
	router := mux.NewRouter()
	router.HandleFunc("/add", createUser).Methods("POST")
	router.HandleFunc("/allUser", fetchall).Methods("GET")
	router.HandleFunc("/delete/{id}", deleteUser).Methods("DELETE")
	router.HandleFunc("/fetchUser/{id}", getUserTask).Methods("GET")
	router.HandleFunc("/updateUser/{id}", updateUser).Methods("PUT")
	http.ListenAndServe(":8080", router)
}
