package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("failed to load SDK configuration, %v", err)
	}
	client := s3.NewFromConfig(cfg)

	var token *string
	for {
		resp, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            aws.String("my-bucket"),
			Prefix:            aws.String("my-prefix"),
			ContinuationToken: token,
		})
		if err != nil {
			log.Fatal("failed to list objects, %v", err)
		}
		for _, c := range resp.Contents {
			fmt.Printf("Name:%s LastModified:%v\n", *c.Key, c.LastModified.Format(time.RFC3339))
		}
		if resp.ContinuationToken == nil {
			break
		}
		token = resp.ContinuationToken
	}
}
