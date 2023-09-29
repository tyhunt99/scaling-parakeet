package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
)

type Application struct {
	config AppConfig
	db     dynamodbiface.DynamoDBAPI
}

type AppConfig struct {
	tableName string
}

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	DOB   string `json:"dob"`
	City  string `json:"city,omitempty"`
}

func (app *Application) createUser(requestData []byte) (*User, error) {
	userId := uuid.New().String()
	fmt.Println("Generated new user id:", userId)

	user := &User{
		Id: userId,
	}
	json.Unmarshal(requestData, user)

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		fmt.Println("Error marshalling item: ", err.Error())
		return nil, fmt.Errorf("failed marshalling item: %s", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(app.config.tableName),
	}
	_, err = app.db.PutItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed calling PutItem: %s", err)
	}

	return user, nil
}

func (app *Application) deleteUser(userId string) error {
	if userId == "" {
		return fmt.Errorf("no valid user id provided")
	}

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(userId),
			},
		},
		TableName: aws.String(app.config.tableName),
	}

	_, err := app.db.DeleteItem(input)

	if err != nil {
		return fmt.Errorf("failed calling DeleteItem: %s", err)
	}

	return nil
}

func (app *Application) getUser(userId string) (*User, error) {
	result, err := app.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(app.config.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(userId),
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed calling GetItem")
	}

	// no user legitimately found
	if len(result.Item) == 0 {
		return nil, nil
	}

	user := &User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, user)

	if err != nil {
		return nil, fmt.Errorf("failed to UnmarshalMap result.Item: %s", err)
	}

	return user, nil
}

func (app *Application) listUsers() ([]*User, error) {
	var users []*User

	// scan table
	result, err := app.db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(app.config.tableName),
	})

	// Checking for errors, return error
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %s", err)
	}

	for _, i := range result.Items {
		user := &User{}

		// result is of type *dynamodb.GetItemOutput
		// result.Item is of type map[string]*dynamodb.AttributeValue
		// UnmarshallMap result.item to item
		err = dynamodbattribute.UnmarshalMap(i, user)

		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling: %s", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (app *Application) buildUserResponse(data any, code int) events.APIGatewayProxyResponse {
	// Marshal item to return
	jsonMarshalled, _ := json.Marshal(data)

	return events.APIGatewayProxyResponse{StatusCode: code, Body: string(jsonMarshalled)}
}

func (app *Application) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user *User
	var users []*User
	var err error
	var response events.APIGatewayProxyResponse
	// TODO: add better path checking
	if !strings.HasPrefix(request.Path, "/users") {
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "Invalid request"}, nil
	}

	// create DynamoDB client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	app.db = dynamodb.New(sess)

	// TODO: should the app functions return the responses for better control over status codes?
	switch request.HTTPMethod {
	case "GET":
		//TODO: should this be split into separate cases
		// var userData string

		if userId, ok := request.PathParameters["id"]; !ok {
			// listing all users
			users, err = app.listUsers()
			response = app.buildUserResponse(users, 200)
		} else {
			// lokking for a single user
			user, err = app.getUser(userId)
			response = app.buildUserResponse(user, 200)
			// special case of no user found
			if user == nil && err == nil {
				response.StatusCode = 404
			}
		}
	case "POST":
		user, err = app.createUser([]byte(request.Body))
		response = app.buildUserResponse(user, 201)
	case "DELETE":
		err = app.deleteUser(request.PathParameters["id"])
		response = events.APIGatewayProxyResponse{StatusCode: 204}
	default:
		response = events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}
	}

	if err != nil {
		errMessage := fmt.Sprintf("Request failed: %s", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: errMessage}, nil
	}

	return response, nil
}
