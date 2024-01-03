package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

func main() {
	ctx := context.TODO()
	bucket, err := blob.OpenBucket(ctx, "file://./localbucket")
	if err != nil {
		log.Fatal(err)
	}
	defer bucket.Close()

	var token = blob.FirstPageToken
	for {
		opts := &blob.ListOptions{
			Prefix: "foo/",
		}
		objs, nextToken, err := bucket.ListPage(ctx, token, 10, opts)
		if err != nil {
			log.Fatal(err)
		}
		for _, obj := range objs {
			fmt.Printf("Name: %s LastModified: %v\n", obj.Key, obj.ModTime.Format(time.RFC3339))
		}
		if nextToken == nil {
			break
		}
		token = nextToken
	}
}
