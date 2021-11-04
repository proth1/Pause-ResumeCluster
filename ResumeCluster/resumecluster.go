package resumecluster

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

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
	region := "us-west-2"

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
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
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
		return "", err
	}

	return secretStruct.CastAIKey, nil
}

func pauseCluster(secret, clusterID string) error {
	pauseURL := fmt.Sprintf("https://api.cast.ai/v1/kubernetes/external-clusters/%s/pause", clusterID)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", pauseURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", secret)
	res, err := httpClient.Do(req)
	if err != nil {
		print(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		print(err)
	}
	fmt.Println(string(body))
	return nil
}

func getClusterID(secret string) (ClusterItems, error) {
	clusterURL := fmt.Sprintf("https://api.cast.ai/v1/kubernetes/external-clusters")
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", clusterURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", secret)
	res, err := httpClient.Do(req)
	if err != nil {
		print(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		print(err)
	}
	var clusters ClusterItems
	err = json.Unmarshal([]byte(body), &clusters)
	if err != nil {
		return clusters, err
	}

	log.Printf(clusters.Items[0].ID)
	return clusters, nil
}

//func handleRequest(ctx context.Context, event events.CloudWatchEvent) (string, error) {
func handleRequest(ctx context.Context, event json.RawMessage) (string, error) {
	var toBePausedClusters ClustersToBePaused
	log.Printf("Getting Started")
	log.Printf(string(event))
	err := json.Unmarshal(event, &toBePausedClusters)
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

	for _, e := range clusters.Items {
		for _, c := range toBePausedClusters.ClusterNames {
			if c == e.Name {
				pauseCluster(secret, e.ID)
			}
		}
	}

	return secret, nil
}

func main() {
	runtime.Start(handleRequest)
}
