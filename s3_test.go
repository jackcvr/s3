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

func setupTest(t *testing.T) *BucketManager {
	c := newClient()
	m := c.NewBucketManager(testBucketName, nil)
	if err := m.EnsureBucket(context.Background(), minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	return m
}

func cleanupServer(c *Client) {
	if c == nil {
		c = newClient()
	}
	ctx := context.Background()
	exists, err := c.BucketExists(ctx, testBucketName)
	if err != nil {
		panic(err)
	}
	if !exists {
		return
	}
	if err = c.RemoveObject(ctx, testBucketName, testObjectName, minio.RemoveObjectOptions{}); err != nil {
		log.Println("cleanup:", err)
	}
	if err = c.RemoveBucket(ctx, testBucketName); err != nil {
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
	m := setupTest(t)
	t.Cleanup(func() {
		cleanupServer(m.Client)
	})

	testData := []byte{0xA, 0xB, 0xC, 0xD, 0xE, 0xF}

	t.Run("Test PutBytes", func(t *testing.T) {
		if _, err := m.PutBytes(context.Background(), testObjectName, testData, minio.PutObjectOptions{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test ReadBytes", func(t *testing.T) {
		var buf bytes.Buffer
		if err := m.ReadBytes(context.Background(), testObjectName, &buf, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), testData) {
			t.Fatal("data != testData")
		}
	})
}

func TestPutRead(t *testing.T) {
	m := setupTest(t)
	t.Cleanup(func() {
		cleanupServer(m.Client)
	})

	type Test struct {
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	testObject := Test{
		Name:   "TestName",
		Amount: 12,
	}
	testData, err := m.Client.Serializer.Serialize(testObject)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test Put", func(t *testing.T) {
		if _, err = m.Put(context.Background(), testObjectName, testObject, minio.PutObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err = m.ReadBytes(context.Background(), testObjectName, &buf, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf.Bytes(), testData) {
			t.Fatal("data != testData")
		}
	})

	t.Run("Test Read", func(t *testing.T) {
		obj := Test{}
		if err = m.Read(context.Background(), testObjectName, &obj, minio.GetObjectOptions{}); err != nil {
			t.Fatal(err)
		}
		if obj != testObject {
			t.Fatal("obj != testObject")
		}
	})
}
