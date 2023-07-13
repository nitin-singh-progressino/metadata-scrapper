package dbstore

import (
	"assignment/constants"
	"assignment/schema"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/logger"
)

// PopulateDynamoDB populates DynamoDB with IPFS metadata
func PopulateDynamoDB(dynamoDBClient *dynamodb.DynamoDB) {

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed to get current directory: %v", err)
	}

	// Specify the filename
	filename := "ipfs_cids.csv"

	// Construct the file path
	filePath := dir + "/" + filename

	// Open the CSV file containing the CIDs
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Read the CIDs from the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logger.Fatalf("Error reading CSV file: %v", err)
	}

	// Populate DynamoDB with IPFS metadata
	for _, record := range records {
		if len(record) > 0 {
			cid := record[0]

			// Construct the URL for the CID
			ipfsURL := constants.Endpoint + cid

			// Make an HTTP GET request to the IPFS URL
			resp, err := http.Get(ipfsURL)
			if err != nil {
				logger.Errorf("Error fetching IPFS content for CID %s: %v", cid, err)
				continue
			}
			defer resp.Body.Close()

			// Check the response status code
			if resp.StatusCode != http.StatusOK {
				logger.Errorf("Error fetching IPFS content for CID %s. Response status: %s", cid, resp.Status)
				continue
			}

			// Parse the JSON response body
			var metadata schema.Metadata
			err = json.NewDecoder(resp.Body).Decode(&metadata)
			if err != nil {
				logger.Errorf("Error decoding JSON for CID %s: %v", cid, err)
				continue
			}

			// Save the metadata in DynamoDB
			err = saveMetadata(dynamoDBClient, metadata, cid)
			if err != nil {
				logger.Errorf("Error saving metadata for CID %s: %v", cid, err)
				continue
			}

			logger.Infof("Metadata saved for CID %s", cid)
		}
	}
}

// Save metadata in DynamoDB
func saveMetadata(dynamoDBClient *dynamodb.DynamoDB, metadata schema.Metadata, cid string) error {
	// Create an item input
	item := map[string]*dynamodb.AttributeValue{
		"cid": {
			S: aws.String(cid),
		},
		"image": {
			S: aws.String(metadata.Image),
		},
		"name": {
			S: aws.String(metadata.Name),
		},
		"description": {
			S: aws.String(metadata.Description),
		},
	}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(constants.TableName),
	}

	// Call the PutItem operation
	_, err := dynamoDBClient.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}
