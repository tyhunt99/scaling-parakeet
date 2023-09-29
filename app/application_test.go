package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockDynamoDB struct {
	dynamodbiface.DynamoDBAPI

	PutItemError  error
	ExistingUsers []*User
}

func (ddb *MockDynamoDB) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if ddb.PutItemError != nil {
		return nil, ddb.PutItemError
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (ddb *MockDynamoDB) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	// always assume the first user is the one to return
	if ddb.ExistingUsers != nil && len(ddb.ExistingUsers) > 0 {
		user := ddb.ExistingUsers[0]
		return &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"id": {S: aws.String(user.Id)},
			},
		}, nil
	}

	return nil, fmt.Errorf("no mocked get item")
}

func TestCreateUser(t *testing.T) {
	tcs := map[string]struct {
		requestData map[string]interface{}

		expectedUser *User
		expectError  bool
	}{
		"base case": {
			requestData: map[string]interface{}{
				"name": "Test Name",
				"dob":  "07-20-1975",
			},

			expectedUser: &User{
				Name: "Test Name",
				DOB:  "07-20-1975",
			},
		},
		"no dob": {
			requestData: map[string]interface{}{
				"name": "Test Name",
			},

			expectedUser: &User{
				Name: "Test Name",
				DOB:  "",
			},
		},
		"failed insert": {
			requestData: map[string]interface{}{
				"name": "Test Name",
			},

			expectedUser: &User{
				Name: "Test Name",
				DOB:  "",
			},
			expectError: true,
		},
	}

	for name, vals := range tcs {
		t.Run(name, func(t *testing.T) {
			mock_ddb := &MockDynamoDB{}

			if vals.expectError {
				mock_ddb.PutItemError = errors.New("failed to insert")
			}

			app := &Application{
				db: mock_ddb,
			}
			data, _ := json.Marshal(vals.requestData)
			u, err := app.createUser(data)

			if vals.expectError {
				assert.NotNil(t, err, "unexpected nil err in createUser")
				// expected err passed so further tests are not needed
				return
			}

			assert.Nil(t, err, "unexpected err in createUser")
			if _, err := uuid.Parse(u.Id); err != nil {
				t.Fatalf("invalid uuid generated")
			}
			assert.Equal(t, vals.expectedUser.Name, u.Name)
		})
	}

}

// TODO: determine how to actually test more than passthrough for mocked dynamodb
func TestGetUser(t *testing.T) {
	tcs := map[string]struct {
		userId string

		expectedUser *User
		expectError  bool
	}{
		"base case": {
			userId: "abcd-1234-wxyz",

			expectedUser: &User{
				Id:    "abcd-1234-wxyz",
				Name:  "Test User",
				Email: "test@test.com",
			},
		},
	}

	for name, vals := range tcs {
		t.Run(name, func(t *testing.T) {

			mock_ddb := &MockDynamoDB{}
			app := &Application{
				db: mock_ddb,
			}

			u, err := app.getUser(vals.userId)

			assert.Nil(t, err)
			assert.NotNil(t, u)

		})
	}
}

func TestHandleRequest(t *testing.T) {
	tcs := map[string]struct {
		request       events.APIGatewayProxyRequest
		existingUsers []*User

		expectedStatusCode int
	}{
		"base case": {
			request: events.APIGatewayProxyRequest{
				Path: "/user",
			},
			existingUsers: []*User{
				{
					Id:    "abcd-1234-wxyz",
					Name:  "Test User",
					Email: "test@test.com",
				},
			},

			expectedStatusCode: 200,
		},
		"invalid path": {
			request: events.APIGatewayProxyRequest{
				Path: "/badpath",
			},

			expectedStatusCode: 404,
		},
	}

	for name, vals := range tcs {
		t.Run(name, func(t *testing.T) {

			mock_ddb := &MockDynamoDB{}
			app := &Application{
				db: mock_ddb,
			}

			response, err := app.HandleRequest(vals.request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, vals.expectedStatusCode, response.StatusCode)

		})
	}
}
