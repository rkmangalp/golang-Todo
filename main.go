package main

import (
	"context"       // Provides context handling for request-scoped values and cancellation signals.
	"encoding/json" // For JSON encoding and decoding.
	"log"           // For logging errors and other information.
	"net/http"      // For HTTP client and server implementations.
	"os"            // For operating system functionalities like signals.
	"os/signal"     // For handling OS signals.
	"strings"       // For string manipulations.
	"time"          // For time-related functions.

	"github.com/go-chi/chi"                      // Lightweight, idiomatic router for building Go HTTP services.
	"github.com/go-chi/chi/middleware"           // Middleware for chi router.
	"github.com/thedevsaddam/renderer"           // For rendering JSON and HTML responses.
	"go.mongodb.org/mongo-driver/bson"           // For BSON handling in MongoDB.
	"go.mongodb.org/mongo-driver/bson/primitive" // For MongoDB ObjectID handling.
	"go.mongodb.org/mongo-driver/mongo"          // MongoDB driver.
	"go.mongodb.org/mongo-driver/mongo/options"  // For MongoDB client options.
)

var rnd *renderer.Render // Renderer for handling JSON and HTML responses.
var db *mongo.Database   // MongoDB database instance.

const (
	hostName       string = "mongodb://127.0.0.1:27017" // MongoDB connection URI.
	dbName         string = "demo_todo"                 // Database name.
	collectionName string = "todo"                      // Collection name.
	port           string = ":9000"                     // Server port.
)

type (
	todoModel struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"` // MongoDB ObjectID.
		Title     string             `bson:"title"`         // Title of the to-do item.
		Completed bool               `bson:"completed"`     // Completion status.
		CreatedAt time.Time          `bson:"createAt"`      // Creation timestamp.
	}
	todo struct {
		ID        string    `json:"id"`        // ID as a string for JSON responses.
		Title     string    `json:"title"`     // Title of the to-do item.
		Completed bool      `json:"completed"` // Completion status.
		CreatedAt time.Time `json:"createAt"`  // Creation timestamp.
	}
)

func init() {
	rnd = renderer.New()                                        // Initializes the renderer.
	clientOptions := options.Client().ApplyURI(hostName)        // Sets MongoDB client options.
	client, err := mongo.Connect(context.TODO(), clientOptions) // Connects to MongoDB.
	checkErr(err)                                               // Checks for connection errors.
	err = client.Ping(context.TODO(), nil)                      // Pings the MongoDB server.
	checkErr(err)                                               // Checks for ping errors.
	db = client.Database(dbName)                                // Selects the database.
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil) // Renders the home template.
	checkErr(err)                                                           // Checks for rendering errors.
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	collection := db.Collection(collectionName)              // Gets the collection.
	cursor, err := collection.Find(context.TODO(), bson.M{}) // Finds all documents.
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to fetch todos",
			"error":   err,
		}) // Responds with an error if the find operation fails.
		return
	}
	defer cursor.Close(context.TODO()) // Ensures the cursor is closed.

	var todos []todoModel // Slice to hold todos.
	if err := cursor.All(context.TODO(), &todos); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to decode todos",
			"error":   err,
		}) // Responds with an error if decoding fails.
		return
	}

	var todoList []todo // Slice to hold the response todos.
	for _, t := range todos {
		todoList = append(todoList, todo{
			ID:        t.ID.Hex(),
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		}) // Converts todos to the response format.
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	}) // Responds with the list of todos.
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err) // Decodes the request body and checks for errors.
		return
	}
	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title is required",
		}) // Checks for missing title.
		return
	}

	tm := todoModel{
		ID:        primitive.NewObjectID(), // Creates a new ObjectID.
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	collection := db.Collection(collectionName)        // Gets the collection.
	_, err := collection.InsertOne(context.TODO(), tm) // Inserts the new todo.
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to save todo",
			"error":   err,
		}) // Responds with an error if the insert operation fails.
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": tm.ID.Hex(),
	}) // Responds with success.
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id")) // Gets and trims the ID from the URL.
	objID, err := primitive.ObjectIDFromHex(id)    // Converts the ID to ObjectID.
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The Id is invalid",
		}) // Responds with an error if the ID is invalid.
		return
	}

	collection := db.Collection(collectionName)                         // Gets the collection.
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID}) // Deletes the document.
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to delete todo",
			"error":   err,
		}) // Responds with an error if the delete operation fails.
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "todo deleted successfully",
	}) // Responds with success.
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id")) // Gets and trims the ID from the URL.
	objID, err := primitive.ObjectIDFromHex(id)    // Converts the ID to ObjectID.
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		}) // Responds with an error if the ID is invalid.
		return
	}

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err) // Decodes the request body and checks for errors.
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is missing",
		}) // Checks for missing title.
		return
	}

	collection := db.Collection(collectionName) // Gets the collection.
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}},
	) // Updates the document.
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to update todo",
			"error":   err,
		}) // Responds with an error if the update operation fails.
		return
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	}) // Responds with success.
}

func main() {
	stopChan := make(chan os.Signal)      // Creates a channel to receive OS signals for graceful shutdown.
	signal.Notify(stopChan, os.Interrupt) // Notifies the channel on receiving an interrupt signal.

	r := chi.NewRouter()             // Creates a new router using chi.
	r.Use(middleware.Logger)         // Adds logging middleware to the router.
	r.Get("/", homeHandler)          // Sets the route for the home handler.
	r.Mount("/todo", todoHandlers()) // Mounts the todoHandlers under the "/todo" path.

	srv := &http.Server{ // Configures the HTTP server.
		Addr:         port,             // Sets the server address and port.
		Handler:      r,                // Sets the router as the request handler.
		ReadTimeout:  60 * time.Second, // Sets the maximum duration for reading the entire request.
		WriteTimeout: 60 * time.Second, // Sets the maximum duration before timing out writes of the response.
		IdleTimeout:  60 * time.Second, // Sets the maximum amount of time to wait for the next request when keep-alives are enabled.
	}

	go func() { // Starts the server in a new goroutine.
		log.Println("listening on port", port)       // Logs that the server is listening on the specified port.
		if err := srv.ListenAndServe(); err != nil { // Starts the HTTP server and logs errors if any.
			log.Printf("listen:%s\n", err)
		}
	}()

	<-stopChan                                                              // Blocks until an interrupt signal is received.
	log.Println("shutting down server...")                                  // Logs that the server is shutting down.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Creates a context with a 5-second timeout for shutdown.
	srv.Shutdown(ctx)                                                       // Shuts down the server gracefully.
	defer cancel()                                                          // Cancels the context to release resources.
	log.Println("server gracefully stopped!")                               // Logs that the server has been stopped gracefully.
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()         // Creates a new router group.
	rg.Group(func(r chi.Router) { // Groups routes related to todo operations.
		r.Get("/", fetchTodos)        // Route for fetching todos.
		r.Post("/", createTodo)       // Route for creating a new todo.
		r.Put("/{id}", updateTodo)    // Route for updating an existing todo.
		r.Delete("/{id}", deleteTodo) // Route for deleting a todo.
	})
	return rg // Returns the router group.
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err) // Logs a fatal error and exits the application.
	}
}
