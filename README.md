# notifyhub

A multi-slack serverless approach to AWS Outage Notifications. Inspired by [aws-status-in-slack](https://github.com/cloudwalkio/aws-status-in-slack), but serverless.

## How does it work?

There are two lambdas using [serverless](https://serverless.com/). One polls AWS and keeps an history on DynamoDB (nh-poll), forwarding to SNS. 

The other (nh-slack) is a SNS Handler via HTTPS which pushes to the right place.

## Install

```
npm install -g serverless
npm install
```

```
go get github.com/constabulary/gb/...
gb vendor restore
bash ./build.sh
```

```
serverless deploy
```

## Usage

 * Create the SNS Topic `nh-dev`
 * Create an Incoming Webhook on Slack, and take note of the three last parameters on its URL (we'll call it TOKENA TOKENB and TOKENC)
 * Create subscriptions to your lambda by means of HTTP, in the format:

```
https://YOURAPIGATEWAYURLCHECKSTACKJSON/dev/slack/CHANNEL/TOKENA/TOKENB/TOKENC
```

  * Hint: check stack.json once you deployed it.
  * ...
  * Profit!
