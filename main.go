package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	bc "github.com/chtison/baseconverter"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var coll *mongo.Collection

func GetPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "9000"
	}
	return ":" + port
}

func init() {
	uri := "mongodb+srv://arpit:8445007708arpit@cluster0.zy1b3.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	//then := time.Now().AddDate(0, -6, 0)
	//collection.DeleteMany(context.TODO(), bson.M{"date": bson.M{"$lt": then}})
	if err != nil {
		log.Fatal(err)
	}
	collection = (*mongo.Collection)(client.Database("URLs").Collection("col"))
	coll = (*mongo.Collection)(client.Database("Last").Collection("row"))
	fmt.Println("Collection instance is ready")
}

func main() {
	router := mux.NewRouter()
	//deleteendpoint()

	router.HandleFunc("/shorten", CreateEndpoint).Methods("POST")
	router.HandleFunc("/{id}", RedirectEndpoint).Methods("GET")
	router.HandleFunc("/", Home).Methods("GET")
	log.Fatal(http.ListenAndServe(GetPort(), router))
}

type URL struct {
	LongURL  string    `json:"longURL,omitempty" bson:"longURL,omitempty"`
	ShortURL string    `json:"shortURL,omitempty" bson:"shortURL,omitempty"`
	Date     time.Time `json:"date,omitempty" bson:"date,omitempty"`
}

type LastURL struct {
	Last int64 `json:"lastURL,omitempty" bson:"lastURL, omitempty"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	w.Write(([]byte("<h1>We offer Url shortening service<h1><h1>Looks like original URL is unvalid or expired</h1>")))
}

func CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var last LastURL
	err := coll.FindOne(context.TODO(), bson.M{}).Decode(&last)
	if err != nil {
		fmt.Print(err)
		var new LastURL
		new.Last = 0
		inserted, err := coll.InsertOne(context.Background(), new)
		last = new
		fmt.Println(inserted)
		if err != nil {
			print(err)
		}
	}
	coll.UpdateOne(context.TODO(), bson.M{"lastURL": last.Last}, bson.M{"$set": bson.M{"lastURL": last.Last + 1}})
	//fmt.Println(result)
	var url URL
	var exist URL
	_ = json.NewDecoder(r.Body).Decode(&url)
	collection.FindOne(context.TODO(), bson.M{"longURL": url.LongURL}).Decode(&exist)
	if exist.LongURL != "" {
		json.NewEncoder(w).Encode(exist)
		return
	}
	var number string = strconv.FormatInt(last.Last, 10)
	var inBase string = "0123456789"
	var toBase string = "!#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	converted, _, _ := bc.BaseToBase(number, inBase, toBase)
	if converted == "~~~~~~" {
		coll.UpdateOne(context.TODO(), bson.M{"lastURL": last.Last}, bson.M{"$set": bson.M{"lastURL": 0}})
	}

	url.ShortURL = converted
	url.Date = time.Now()
	//fmt.Println(url.LongURL)
	//fmt.Println(url.ShortURL)
	//fmt.Println(url.Date)
	insertdata(url)
	json.NewEncoder(w).Encode(url)
}

//this is done
func RedirectEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	long := findurl(params["id"])
	//json.NewEncoder(w).Encode(long)
	http.Redirect(w, r, long, http.StatusPermanentRedirect)
}

func insertdata(new URL) {
	inserted, err := collection.InsertOne(context.Background(), new)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(inserted.InsertedID)
}

func findurl(short string) string {
	var long URL
	err := collection.FindOne(context.TODO(), bson.M{"shortURL": short}).Decode(&long)

	if err != nil {
		fmt.Println(err)
	}
	return long.LongURL
}

func deleteendpoint() {
	for range time.Tick(time.Hour * 4380) {
		then := time.Now().AddDate(0, -6, 0)
		//fmt.Println(then)
		//fmt.Println(time.Now())
		result, _ := collection.DeleteMany(context.TODO(), bson.M{"date": bson.M{"$lt": then}})
		fmt.Println(result)
	}
}
