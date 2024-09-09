package s3

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type Serializer interface {
	Serialize(any) ([]byte, error)
	Deserialize([]byte, any) error
}

type JSONSerializer struct{}

func (_ JSONSerializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (_ JSONSerializer) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewOptions(accessKeyID, secretAccessKey, sessionToken string, secure bool) *minio.Options {
	return &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, sessionToken),
		Secure: secure,
	}
}

type Client struct {
	*minio.Client
	Serializer Serializer
}

func NewClient(endpoint string, opts *minio.Options, serializer Serializer) (*Client, error) {
	mc, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}
	if serializer == nil {
		serializer = JSONSerializer{}
	}
	return &Client{
		Client:     mc,
		Serializer: serializer,
	}, nil
}

func (c *Client) PutBytes(ctx context.Context, bucketName, objectName string, value []byte, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	reader := bytes.NewReader(value)
	return c.PutObject(ctx, bucketName, objectName, reader, reader.Size(), opts)
}

func (c *Client) Put(ctx context.Context, bucketName, objectName string, value any, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	data, err := c.Serializer.Serialize(value)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	return c.PutBytes(ctx, bucketName, objectName, data, opts)
}

func (c *Client) ReadBytes(ctx context.Context, bucketName string, objectName string, dst io.Writer, opts minio.GetObjectOptions) error {
	obj, err := c.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return err
	}
	defer obj.Close()
	if _, err = io.Copy(dst, obj); err != nil {
		return err
	}
	return nil
}

func (c *Client) Read(ctx context.Context, bucketName string, objectName string, dst any, opts minio.GetObjectOptions) error {
	var buf bytes.Buffer
	if err := c.ReadBytes(ctx, bucketName, objectName, &buf, opts); err != nil {
		return err
	}
	return c.Serializer.Deserialize(buf.Bytes(), dst)
}
