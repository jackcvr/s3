# Go S3 client

Extended MinIO S3 client with (de)serialization.

Added extra functionality:
 - **PutBytes**, **ReadBytes** functions: for more convenient storage of bytes
 - **Put**, **Read** functions: for storage of structs (serialized in JSON/MsgPack/etc...)
 - **BucketClient**: for encapsulating functionality around specific bucket

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

func main() {
	var accessKey = os.Getenv("ACCESS_KEY")
	var secretKey = os.Getenv("SECRET_KEY")

	// create BucketClient to simplify working with a single bucket
	c, err := s3.NewBucketClient(
		"localhost:9000",
		bucketName,
		s3.NewOptions(accessKey, secretKey, "", false),
		msgpack.MsgPackSerializer{}) // use MsgPack serializer for this bucket
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// create bucket if it doesn't exist
	if err = c.EnsureBucket(ctx, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		panic(err)
	}

	// Put bytes
	if _, err = c.PutBytes(ctx, objectName, []byte("test"), minio.PutObjectOptions{}); err != nil {
		panic(err)
	}
	// Get bytes
	var buf bytes.Buffer
	if err = c.ReadBytes(ctx, objectName, &buf, minio.GetObjectOptions{}); err != nil {
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
	if _, err = c.Put(ctx, objectName, obj, minio.PutObjectOptions{}); err != nil {
		panic(err)
	}

	// Get deserialized struct
	obj2 := Test{}
	if err = c.Read(ctx, objectName, &obj2, minio.GetObjectOptions{}); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", obj2) // {Name:Test Amount:12}

	buf.Reset()
	if err = c.ReadBytes(ctx, objectName, &buf, minio.GetObjectOptions{}); err != nil {
		panic(err)
	}
	fmt.Println(hex.Dump(buf.Bytes()))
	//00000000  82 a4 4e 61 6d 65 a4 54  65 73 74 a6 41 6d 6f 75  |..Name.Test.Amou|
	//00000010  6e 74 0c                                          |nt.|
}
```

## License

[MIT](https://spdx.org/licenses/MIT.html) 