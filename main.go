package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Announcement struct {
	IsActive bool               `json:"isActive" bson:"isActive"`
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title" bson:"title"`
	Content  string             `json:"content" bson:"content"`
	Date     time.Time          `json:"date" bson:"date"`
}

// middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://admin:pcqfugdm1cms@cluster0.kzftowi.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"))
	if err != nil {
		panic(err)
	}

	var dbName = "test"
	var collectionName = "announcements"
	collection := client.Database(dbName).Collection(collectionName)

	//Get Single Announcement
	getSingleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the ID from the URL path
		id, err := primitive.ObjectIDFromHex(strings.TrimPrefix(r.URL.Path, "/api/announcements/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Find the document in the collection
		var announcement Announcement
		err = collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&announcement)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode the announcement as JSON and send it as the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(announcement)
	})

	// Get All Announcements
	getAllHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := collection.Find(context.TODO(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var announcements []Announcement
		err = data.All(context.TODO(), &announcements)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(announcements)
	})

	// Create Announcement
	createHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var announcement Announcement

		err := json.NewDecoder(r.Body).Decode(&announcement)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := collection.InsertOne(context.TODO(), announcement)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Inserted document with ID: %v", result.InsertedID)
	})

	// Update Announcement
	updateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(strings.TrimPrefix(r.URL.Path, "/api/announcements/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var announcement Announcement
		err = json.NewDecoder(r.Body).Decode(&announcement)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": announcement})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Updated document with ID: %v", id)
	})

	// Delete Announcement
	deleteHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(strings.TrimPrefix(r.URL.Path, "/api/announcements/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Deleted document with ID: %v", id)
	})

	http.Handle("GET /api/announcements/", corsMiddleware(getSingleHandler))
	http.Handle("GET /api/announcements", corsMiddleware(getAllHandler))
	http.Handle("POST /api/announcements", corsMiddleware(createHandler))
	http.Handle("PUT /api/announcements/", corsMiddleware(updateHandler))
	http.Handle("DELETE /api/announcements/", corsMiddleware(deleteHandler))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
