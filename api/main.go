package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const apiKey = "c8fd43eb56e81c44f40e95ad7923f7e1"
const database = "TMDBDump"
const configCollection = "ConfigCollection"

type ConfigurationStruct struct {
	Images struct {
		BaseURL       string   `json:"base_url"`
		SecureBaseURL string   `json:"secure_base_url"`
		BackdropSizes []string `json:"backdrop_sizes"`
		LogoSizes     []string `json:"logo_sizes"`
		PosterSizes   []string `json:"poster_sizes"`
		ProfileSizes  []string `json:"profile_sizes"`
		StillSizes    []string `json:"still_sizes"`
	} `json:"images"`
	ChangeKeys []string `json:"change_keys"`
}

var mongoClient *mongo.Client

func main() {
	//configureMogoClient()
	setConfigurationSettings()
	configureRoute()
}

func configureRoute() {
	r := mux.NewRouter()
	r.HandleFunc("/config", handleConfigSetting).Methods("GET")
	http.ListenAndServe(":8081", r)
}

func configureMogoClient() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().SetHosts([]string{"mongo"})
	clientNew, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("1")
	if err != nil {
		fmt.Println("mongo err", err)
		return
	}
	err = clientNew.Ping(ctx, nil)
	fmt.Println("2")
	if err != nil {
		fmt.Println("mongo err2", err)
		return
	}
	fmt.Println("Connected to MongoDB!")
	mongoClient = clientNew
}

func setConfigurationSettings() {
	configurationURL := "https://api.themoviedb.org/3/configuration?api_key=" + apiKey
	client := &http.Client{}
	req, _ := http.NewRequest("GET", configurationURL, nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request")
		return
	}
	defer resp.Body.Close()

	var configuration ConfigurationStruct
	_ = json.NewDecoder(resp.Body).Decode(&configuration)
	collection := mongoClient.Database(database).Collection(configCollection)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, configuration)
	fmt.Println(result)
}

func getConfigurationSettings() (ConfigurationStruct, bool) {
	var configs []ConfigurationStruct
	collection := mongoClient.Database(database).Collection(configCollection)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return ConfigurationStruct{}, false
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var config ConfigurationStruct
		cursor.Decode(&config)
		configs = append(configs, config)
	}

	if err := cursor.Err(); err != nil {
		return ConfigurationStruct{}, false
	}
	return configs[0], true
}
func handleConfigSetting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	config, succ := getConfigurationSettings()
	if succ {
		json.NewEncoder(w).Encode(config)
	} else {
		w.Write([]byte(`{"message": "` + "error Occured" + `" }`))
	}
}
