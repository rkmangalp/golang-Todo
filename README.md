# TODO List API

A simple TODO list API built with Go and MongoDB. This application provides basic CRUD operations for managing TODO tasks and uses the `go-chi` router for HTTP handling.

## Features

- **Create** a new TODO item
- **Read** all TODO items
- **Update** an existing TODO item
- **Delete** a TODO item
- Uses MongoDB for data storage

## Getting Started

### Prerequisites

Make sure you have the following installed on your system:

- [Go](https://golang.org/dl/) (version 1.20 or higher)
- [MongoDB](https://www.mongodb.com/try/download/community) (version 7.0 or higher)
- [Mongo Shell](https://www.mongodb.com/try/download/shell) (mongosh) for database management

### Installation

1. **Clone the Repository**

   ```sh
   git clone https://github.com/rkmangalp/golang-Todo
   cd your-repository

### Install Dependencies

Ensure you have Go modules enabled and run:

```sh
go mod tidy
```

This will download all the required dependencies specified in go.mod.

# Update MongoDB Connection

Make sure MongoDB is running on your local machine. You can check the connection in the `init` function of `main.go`:

```go
const (
    hostName       string = "mongodb://127.0.0.1:27017"
    dbName         string = "demo_todo"
    collectionName string = "todo"
    port           string = ":9000"
)
```
Update hostName if your MongoDB server is hosted on a different address.

# Run the Application

Build and start the application:

```sh
go run main.go
```

The application will start listening on http://localhost:9000.

## API Endpoints

Here is a list of available API endpoints:

| Method | Endpoint     | Description                  |
|--------|--------------|------------------------------|
| GET    | /            | Render the home page        |
| GET    | /todo         | Fetch all TODO items        |
| POST   | /todo         | Create a new TODO item      |
| PUT    | /todo/{id}    | Update a TODO item          |
| DELETE | /todo/{id}    | Delete a TODO item          |

### Example Requests

- **Fetch Todos**

  ```sh
  curl -X GET http://localhost:9000/todo

# Example Requests

Here are some example requests for the TODO List API:

### Fetch Todos

Retrieve all TODO items:

  ```sh
  curl -X GET http://localhost:9000/todo
  ```

### Create a Todo

  ```sh
  curl -X POST http://localhost:9000/todo -H "Content-Type: application/json" -d '{"title": "New Todo"}'
  ```
### Update a Todo
### Update an existing TODO item by ID:

  ```sh
  curl -X PUT http://localhost:9000/todo/{id} -H "Content-Type: application/json" -d '{"title": "Updated Todo", "completed": true}'
  ```
### Delete a Todo
### Delete a TODO item by ID:

  ```sh
  curl -X DELETE http://localhost:9000/todo/{id}
  ```
## Project Structure

- **`main.go`** - Main application file containing server setup and route handling.
- **`go.mod`** - Go module file with dependencies.
- **`static/home.tpl`** - HTML template for the home page (create this file for the home page rendering).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
