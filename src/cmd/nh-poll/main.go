package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"services"
	log "github.com/sirupsen/logrus"
)

func Handler() (error) {
	f := services.NewFeedService()

	err := f.Update()

	return err
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	log.SetLevel(log.DebugLevel)

	lambda.Start(Handler)
}
