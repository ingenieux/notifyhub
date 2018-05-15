package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/Jeffail/gabs"
	"gopkg.in/resty.v1"
	"strconv"
	"fmt"
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	c, err := gabs.ParseJSON([]byte(req.Body))

	bodyType, _ := strconv.Unquote(c.S("Type").String())

	log.Infof("Type: %s", bodyType)

	if "Notification" == bodyType {
		response, err = handleNotification(req, c)
	} else if "SubscriptionConfirmation" == bodyType {
		response, err = handleSubscriptionConfirmation(req, c)
	} else if "UnsubscribeConfirmation" == bodyType {
		response, err = handleUnsubscribeConfirmation(req, c)
	}

	return response, err
}

func handleNotification(req events.APIGatewayProxyRequest, c *gabs.Container) (response events.APIGatewayProxyResponse, err error) {
	channel := req.PathParameters["channel"]

	token0 := req.PathParameters["token0"]
	token1 := req.PathParameters["token1"]
	token2 := req.PathParameters["token2"]

	messageBody, _ := strconv.Unquote(c.S("Message").String())

	parsed, err := gabs.ParseJSON([]byte(messageBody))

	if nil != err {
		log.Warnf("Oops: %s", err)

		return
	}

	urlToCall := fmt.Sprintf("https://hooks.slack.com/services/%s/%s/%s", token0, token1, token2)

	log.Infof("Calling url: %s with args: %s", fmt.Sprintf("https://hooks.slack.com/services/%s/%s/%s", token0, token1, token2), parsed.String())

	resp, err := resty.R().
		SetHeader("Content-Type", "text/json").
		SetBody(map[string]interface{}{
		"username":    "aws",
		"icon_url":    "https://dl.dropboxusercontent.com/u/62469907/aws-icon.png",
		"channel":     channel,
		"attachments": []interface{}{parsed.Data()},
	}).
		Post(urlToCall)

	if nil != err {
		log.Warnf("Oops: %s", err)
		return
	}

	log.Infof("resp.Body: %s statusCode: %03d", resp.Body(), resp.StatusCode())

	response = events.APIGatewayProxyResponse{
		StatusCode: 200,
	}

	return
}

func handleSubscriptionConfirmation(req events.APIGatewayProxyRequest, c *gabs.Container) (response events.APIGatewayProxyResponse, err error) {
	subscribeURL, _ := strconv.Unquote(c.S("SubscribeURL").String())

	log.Infof("Retrieving from subscribeURL: %s", subscribeURL)

	resp, err := resty.R().
		Get(subscribeURL)

	if nil != err {
		log.Warnf("Oops: %s", err)

		return
	}

	log.Infof("Resp: %s (%s)", string(resp.Body()), resp.StatusCode())

	response = events.APIGatewayProxyResponse{
		StatusCode: 200,
	}

	return
}

func handleUnsubscribeConfirmation(req events.APIGatewayProxyRequest, c *gabs.Container) (response events.APIGatewayProxyResponse, err error) {
	return
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	log.SetLevel(log.DebugLevel)

	lambda.Start(Handler)
}
