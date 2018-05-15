package services

import (
	"time"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"strings"
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"bytes"
	"encoding/json"
	"fmt"
)

var (
	sess       = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	ddb        = dynamodb.New(sess)
	snsService = sns.New(sess)
)

type FeedService struct {
	GUID          string
	URL           string
	Interval      time.Duration
	DynamoDBTable string
	SNSTopic      string
}

type FeedItem struct {
	GUID  string `json:"guid"`
	Color string `json:"color"`
	Title string `json:"title"`
	Link  string `json:"title_link"`
	Text  string `json:"summary"`
}

func NewFeedService() *FeedService {
	return &FeedService{
		URL:           "http://status.aws.amazon.com/rss/all.rss",
		Interval:      30 * time.Minute,
		DynamoDBTable: os.Getenv("DYNAMODB_TABLE"),
		SNSTopic:      os.Getenv("SNS_TOPIC"),
	}
}

func (f *FeedService) FetchNewItems() (items []*FeedItem, err error) {
	log.Infof("Scanning %s for items less than %s", f.URL, f.Interval.String())

	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(f.URL)

	if nil != err {
		log.Warnf("Oops: %s", err)

		return
	}

	log.Infof("Total: %d items", len(feed.Items))

	items = []*FeedItem{}

	for _, v := range feed.Items {
		itemTime := *v.PublishedParsed

		elapsed := time.Now().Sub(itemTime)

		color := "danger"

		if strings.Contains(v.Title, "operating normally") {
			color = "good"
		} else if strings.Contains(v.Title, "Informational message:") {
			color = "warning"
		}

		if elapsed < f.Interval {
			log.Infof("Testing: %+v", v)

			newItem := &FeedItem{
				Color: color,
				Title: v.Title,
				Text:  v.Description,
				Link:  v.Link,
				GUID:  v.GUID,
			}

			log.Infof("Testing guid %s (event: %+v)", v.GUID, newItem)

			getItemResponse, err := ddb.GetItem(&dynamodb.GetItemInput{
				TableName: aws.String(f.DynamoDBTable),
				Key: map[string]*dynamodb.AttributeValue{
					"guid": &dynamodb.AttributeValue{S: aws.String(v.GUID)},
				},
			})

			log.Debugf("getItemResponse, err: %+v, %s", getItemResponse, err)

			if 0 == len(getItemResponse.Item) {
				items = append(items, newItem)
			} else if nil != err {
				log.Warnf("Oops: %s", err)

				continue
			}
		}
	}

	log.Infof("Returning %d items", len(items))

	return
}

func (f *FeedService) Update() (err error) {
	items, err := f.FetchNewItems()

	if nil != err {
		log.Warnf("Oops: %s", err)

		return err
	}

	for _, v := range items {
		log.Infof("Publishing: %+v", v)

		buf := bytes.NewBuffer([]byte{})

		err = json.NewEncoder(buf).Encode(v)

		if nil != err {
			log.Warnf("Oops: %s", err)

			return err
		}

		_, err = ddb.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(f.DynamoDBTable),
			Item: map[string]*dynamodb.AttributeValue{
				"guid":      &dynamodb.AttributeValue{S: aws.String(v.GUID)},
				"date_time": &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", time.Now().Add(24 * time.Hour).Unix()))},
			},
		})

		if nil != err {
			log.Warnf("Oops: %s", err)

			return err
		}

		out, err := snsService.Publish(&sns.PublishInput{
			TopicArn: aws.String(f.SNSTopic),
			Message:  aws.String(buf.String()),
		})

		if nil != err {
			log.Warnf("Oops: %s", err)

			return err
		}

		log.Infof("Out: %s", out.String())
	}

	return
}
