package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"fmt"

	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var client = lambda.New(session.New())

func getSecret() (string, error) {
	secretName := "CastAI-API"
	region := os.Getenv("AWS_REGION")

	//Create a Secrets Manager client
	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				log.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				log.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				log.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				log.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				log.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return "", err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString string
	//decodedBinarySecret string
	if result.SecretString != nil {
		secretString = *result.SecretString
	}

	var secretStruct Secret
	err = json.Unmarshal([]byte(secretString), &secretStruct)
	if err != nil {
		// If unmarshalling fails, return the error message.
		return "", err
	}

	// return the secret to be used for the CastAI API Calls
	return secretStruct.CastAIKey, nil
}

func clusterAction(secret, clusterID, action string) error {
	// Define the CastAI API endpoint to be used for the calls.
	URL := fmt.Sprintf("https://api.cast.ai/v1/kubernetes/external-clusters/%s/%s", clusterID, action)
	httpClient := &http.Client{}

	// Create the Post request to CastAI to pause or resume the cluster.
	req, _ := http.NewRequest("POST", URL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", secret)

	// Submit the request and capture the response
	_, err := httpClient.Do(req)
	if err != nil {
		// If the API request fails, return the error message
		return err
	}

	return nil
}

// getClusterID will retrieve the list of clusters and create a struct to be used for gathering Cluster ID's.
func getClusterID(secret string) (ClusterItems, error) {

	var clusters ClusterItems
	clusterURL := "https://api.cast.ai/v1/kubernetes/external-clusters"
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", clusterURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", secret)

	// Submit the GET request to retrieve the list of clusters
	res, err := httpClient.Do(req)

	if err != nil {
		return clusters, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return clusters, err
	}
	err = json.Unmarshal([]byte(body), &clusters)
	if err != nil {
		return clusters, err
	}

	return clusters, nil
}

//func handleRequest(ctx context.Context, event events.CloudWatchEvent) (string, error) {
func handleRequest(ctx context.Context, event json.RawMessage) (string, error) {
	clustersToAction := ClustersToAction{}
	log.Printf("Getting Started")
	log.Printf(string(event))
	err := json.Unmarshal(event, &clustersToAction)
	if err != nil {
		return "ERROR", err
	}
	secret, err := getSecret()
	if err != nil {
		return "ERROR", err
	}
	clusters, err := getClusterID(secret)
	if err != nil {
		return "ERROR", err
	}

	// Iterate through clusters, match the names to the cluster names passed in the EventBridge action.
	for _, e := range clusters.Items {
		for _, c := range clustersToAction.ClusterNames {
			if c == e.Name {
				clusterAction(secret, e.ID, clustersToAction.Action)
			}
		}
	}

	return "", nil
}

func main() {
	runtime.Start(handleRequest)
}
