package s3

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"log"
	"os"
	"testing"
)

const (
	testBucketName = "test-bucket"
	testObjectName = "test-object"
)

var (
	serverEndpoint = os.Getenv("SERVER_ENDPOINT")
	accessKey      = os.Getenv("ACCESS_KEY")
	secretKey      = os.Getenv("SECRET_KEY")
)

func newClient() *Client {
	c, err := NewClient(serverEndpoint, NewOptions(accessKey, secretKey, "", false), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func setupTest(t *testing.T) *Client {
	c := newClient()
	if err := c.MakeBucket(context.Background(), testBucketName, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	return c
}

func cleanupServer(c *Client) {
	if c == nil {
		c = newClient()
	}
	if err := c.RemoveObject(context.Background(), testBucketName, testObjectName, minio.RemoveObjectOptions{}); err != nil {
		log.Println("cleanup:", err)
	}
	if err := c.RemoveBucket(context.Background(), testBucketName); err != nil {
		log.Println("cleanup:", err)
	}
}

func TestMain(m *testing.M) {
	if serverEndpoint == "" || accessKey == "" || secretKey == "" {
		panic("serverEndpoint or accessKey or secretKey is empty")
	}
	cleanupServer(nil)
	os.Exit(m.Run())
}

func TestPutReadBytes(t *testing.T) {
	c := setupTest(t)
	t.Cleanup(func() {
		cleanupServer(c)
	})

	testData := []byte{0xA, 0xB, 0xC, 0xD, 0xE, 0xF}

	t.Run("Test PutBytes", func(t *testing.T) {
		if _, err := c.PutBytes(context.Background(), testBucketName, testObjectName, testData, minio.PutObjectOptions{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test ReadBytes", func(t *testing.T) {
		var buf bytes.Buffer
		if err := c.ReadBytes(context.Background(), testBucketName, testObjectName, &buf, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), testData) {
			t.Fatal("data != testData")
		}
	})
}

func TestPutRead(t *testing.T) {
	c := setupTest(t)
	t.Cleanup(func() {
		cleanupServer(c)
	})

	type Test struct {
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	testObject := Test{
		Name:   "TestName",
		Amount: 12,
	}
	testData, err := defaultJSONSerializer.Serialize(testObject)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test Put", func(t *testing.T) {
		if _, err = c.Put(context.Background(), testBucketName, testObjectName, testObject, minio.PutObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err = c.ReadBytes(context.Background(), testBucketName, testObjectName, &buf, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), testData) {
			t.Fatal("data != testData")
		}
	})

	t.Run("Test Read", func(t *testing.T) {
		obj := Test{}
		if err = c.Read(context.Background(), testBucketName, testObjectName, &obj, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if obj != testObject {
			t.Fatal("obj != testObject")
		}
	})
}
