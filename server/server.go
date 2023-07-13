package server

import (
	"assignment/constants"
	"assignment/schema"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/logger"
	"github.com/gorilla/mux"
)

//curl http://localhost:8080/tokens | jq

// FetchAllData fetches all data from DynamoDB
func FetchAllData(dynamoDBClient *dynamodb.DynamoDB) ([]*schema.Metadata, error) {

	// Create the Scan input
	input := &dynamodb.ScanInput{
		TableName: aws.String(constants.TableName),
	}

	// Call the Scan operation
	result, err := dynamoDBClient.Scan(input)
	if err != nil {
		return nil, err
	}

	// Convert the DynamoDB items to schema.Metadata structs
	var items []*schema.Metadata
	for _, item := range result.Items {
		metadata := &schema.Metadata{
			Image:       aws.StringValue(item["image"].S),
			Name:        aws.StringValue(item["name"].S),
			Description: aws.StringValue(item["description"].S),
		}
		items = append(items, metadata)
	}

	return items, nil
}

// FetchData fetches data for a specific CID from DynamoDB
func FetchData(dynamoDBClient *dynamodb.DynamoDB, cid string) (*schema.Metadata, error) {
	// Create the GetItem input
	input := &dynamodb.GetItemInput{
		TableName: aws.String(constants.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"cid": {
				S: aws.String(cid),
			},
		},
	}

	// Call the GetItem operation
	result, err := dynamoDBClient.GetItem(input)
	if err != nil {
		return nil, err
	}

	// If no item found, return nil
	if len(result.Item) == 0 {
		return nil, nil
	}

	// Convert the DynamoDB item to a Metadata struct
	item := &schema.Metadata{
		Image:       aws.StringValue(result.Item["image"].S),
		Name:        aws.StringValue(result.Item["name"].S),
		Description: aws.StringValue(result.Item["description"].S),
	}

	return item, nil
}

// SetupAPI sets up the RESTful API with endpoints
func SetupAPI(dynamoDBClient *dynamodb.DynamoDB) {
	// Create a new router using Gorilla Mux
	router := mux.NewRouter()

	// Define the GET /tokens endpoint
	router.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {

		// Get the security token from the request header
		token := extractTokenFromHeader(r.Header.Get("Authorization"))

		// Add the security token to the request header
		r.Header.Set("Authorization", "Bearer "+token)

		// Fetch all data from DynamoDB
		items, err := FetchAllData(dynamoDBClient)
		if err != nil {
			logger.Errorf("Error fetching data: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert the items to JSON
		jsonData, err := json.Marshal(items)
		if err != nil {
			logger.Errorf("Error encoding data to JSON: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the response content type
		w.Header().Set("Content-Type", "application/json")

		// Write the response
		w.Write(jsonData)
	}).Methods("GET")

	// Define the GET /tokens/{cid} endpoint
	router.HandleFunc("/tokens/{cid}", func(w http.ResponseWriter, r *http.Request) {
		// Get the CID from the request path parameters
		vars := mux.Vars(r)
		cid := vars["cid"]

		// Fetch the data for the CID from DynamoDB
		item, err := FetchData(dynamoDBClient, cid)
		if err != nil {
			logger.Errorf("Error fetching data: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// If no item found for the CID, return 404 Not Found
		if item == nil {
			http.NotFound(w, r)
			return
		}

		// Convert the item to JSON
		jsonData, err := json.Marshal(item)
		if err != nil {
			logger.Errorf("Error encoding data to JSON: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the response content type
		w.Header().Set("Content-Type", "application/json")

		// Write the response
		w.Write(jsonData)
	}).Methods("GET")

	// Start the HTTP server
	logger.Info("Starting server on port 8080")
	logger.Fatal(http.ListenAndServe(":8080", router))
}

// Function to extract the token from the Authorization header
func extractTokenFromHeader(headerValue string) string {
	// Check if the header value is empty or doesn't start with "Bearer "
	if headerValue == "" || !strings.HasPrefix(headerValue, "Bearer ") {
		return ""
	}

	// Extract the token from the header value
	token := strings.TrimPrefix(headerValue, "Bearer ")

	return token
}
