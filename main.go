package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	application "github.com/tyhunt99/scaling-parakeet/app"
)

func main() {
	app := &application.Application{}
	lambda.Start(app.HandleRequest)
}
