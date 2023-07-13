package main

import (
	"assignment/constants"
	"assignment/dbstore"
	"assignment/schema"
	"assignment/server"
	"context"
	"encoding/json"
	"log"

	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/logger"
)

func init() {
	// Initialize the logger
	logger.Init("LoggerExample", true, false, io.Discard)
}

func main() {

	//TODO
	AccessKeyId, SecretAccessKey := getSecrets()

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(constants.Region),
		Credentials: credentials.NewStaticCredentials(AccessKeyId, SecretAccessKey, ""),
	})
	if err != nil {
		logger.Fatal("Failed to create AWS session:", err)
	}

	// Create a DynamoDB client
	dynamoDBClient := dynamodb.New(sess)

	// Populate DynamoDB with IPFS metadata
	dbstore.PopulateDynamoDB(dynamoDBClient)

	// Set up the RESTful API
	server.SetupAPI(dynamoDBClient)
}

func getSecrets() (string, string) {

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(constants.Region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(constants.SecretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {

		logger.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString
	logger.Info("Testing:", secretString)

	var secrets schema.Secrets
	err = json.Unmarshal([]byte(secretString), &secrets)
	if err != nil {
		logger.Fatal("Error:", err)
	}
	return secrets.NitinAccessKey, secrets.NitinSecret
}