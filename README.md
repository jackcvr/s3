# Go S3 client

MinIO S3 client with (de)serialization.

Added extra methods:
 - PutBytes, ReadBytes: for more convenient storage of bytes
 - Put, Read: for storage of structs (serialized by JSON/MsgPack/etc...)

## Usage

```go
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jackcvr/s3"
	"github.com/jackcvr/s3/msgpack"
	"github.com/minio/minio-go/v7"
	"os"
)

const (
	bucketName = "test-bucket"
	objectName = "test-object"
)

func createBucket(ctx context.Context, c *s3.Client) {
	if exists, err := c.BucketExists(ctx, bucketName); err != nil {
		panic(err)
	} else if !exists {
		if err = c.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			panic(err)
		}
	}
}

func main() {
	var (
		accessKey = os.Getenv("ACCESS_KEY")
		secretKey = os.Getenv("SECRET_KEY")
	)

	c, err := s3.NewClient(
		"localhost:9000",
		s3.NewOptions(accessKey, secretKey, "", false),
		msgpack.MsgPackSerializer{}) // use MsgPack serializer instead of default one(JSON)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	createBucket(ctx, c)

	// Put bytes
	if _, err = c.PutBytes(ctx, bucketName, objectName, []byte("test"), minio.PutObjectOptions{}); err != nil {
		panic(err)
	}
	// Get bytes
	var buf bytes.Buffer
	if err = c.ReadBytes(ctx, bucketName, objectName, &buf, minio.GetObjectOptions{}); err != nil {
		panic(err)
	}
	fmt.Println(string(buf.Bytes())) // test

	type Test struct {
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	// Put serialized struct
	obj := Test{
		Name:   "Test",
		Amount: 12,
	}
	if _, err = c.Put(ctx, bucketName, objectName, obj, minio.PutObjectOptions{}); err != nil {
		panic(err)
	}

	// Get deserialized struct
	obj2 := Test{}
	if err = c.Read(ctx, bucketName, objectName, &obj2, minio.GetObjectOptions{}); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", obj2) // {Name:Test Amount:12}

	buf.Reset()
	if err = c.ReadBytes(ctx, bucketName, objectName, &buf, minio.GetObjectOptions{}); err != nil {
		panic(err)
	}
	fmt.Println(hex.Dump(buf.Bytes()))
	//00000000  82 a4 4e 61 6d 65 a4 54  65 73 74 a6 41 6d 6f 75  |..Name.Test.Amou|
	//00000010  6e 74 0c                                          |nt.|
}

```

## License

[MIT](https://spdx.org/licenses/MIT.html) 