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

func (c *Client) EnsureBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	exists, err := c.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return c.MakeBucket(ctx, bucketName, opts)
}

func (c *Client) PutBytes(ctx context.Context, bucketName, objectName string, value []byte, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	reader := bytes.NewReader(value)
	return c.PutObject(ctx, bucketName, objectName, reader, reader.Size(), opts)
}

func (c *Client) Put(ctx context.Context, bucketName, objectName string, value any, opts minio.PutObjectOptions, s ...Serializer) (minio.UploadInfo, error) {
	ser := c.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	data, err := ser.Serialize(value)
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

func (c *Client) Read(ctx context.Context, bucketName string, objectName string, dst any, opts minio.GetObjectOptions, s ...Serializer) error {
	var buf bytes.Buffer
	if err := c.ReadBytes(ctx, bucketName, objectName, &buf, opts); err != nil {
		return err
	}
	ser := c.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	return ser.Deserialize(buf.Bytes(), dst)
}

func (c *Client) GetBucketClient(bucketName string, serializer Serializer) *BucketClient {
	if serializer == nil {
		serializer = c.Serializer
	}
	return &BucketClient{
		Client:     c,
		BucketName: bucketName,
		Serializer: serializer,
	}
}

func NewBucketClient(endpoint string, bucketName string, opts *minio.Options, serializer Serializer) (*BucketClient, error) {
	c, err := NewClient(endpoint, opts, serializer)
	if err != nil {
		return nil, err
	}
	return c.GetBucketClient(bucketName, serializer), nil
}

type BucketClient struct {
	Client     *Client
	BucketName string
	Serializer Serializer
}

func (c *BucketClient) EnsureBucket(ctx context.Context, opts minio.MakeBucketOptions) error {
	return c.Client.EnsureBucket(ctx, c.BucketName, opts)
}

func (c *BucketClient) RemoveBucket(ctx context.Context) error {
	return c.Client.RemoveBucket(ctx, c.BucketName)
}

func (c *BucketClient) ListObjects(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.Client.ListObjects(ctx, c.BucketName, opts)
}

func (c *BucketClient) StatObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (minio.ObjectInfo, error) {
	return c.Client.StatObject(ctx, c.BucketName, objectName, opts)
}

func (c *BucketClient) GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return c.Client.GetObject(ctx, c.BucketName, objectName, opts)
}

func (c *BucketClient) RemoveObject(ctx context.Context, objectName string, opts minio.RemoveObjectOptions) error {
	return c.Client.RemoveObject(ctx, c.BucketName, objectName, opts)
}

func (c *BucketClient) PutBytes(ctx context.Context, objectName string, value []byte, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return c.Client.PutBytes(ctx, c.BucketName, objectName, value, opts)
}

func (c *BucketClient) Put(ctx context.Context, objectName string, value any, opts minio.PutObjectOptions, s ...Serializer) (minio.UploadInfo, error) {
	ser := c.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	return c.Client.Put(ctx, c.BucketName, objectName, value, opts, ser)
}

func (c *BucketClient) ReadBytes(ctx context.Context, objectName string, dst io.Writer, opts minio.GetObjectOptions) error {
	return c.Client.ReadBytes(ctx, c.BucketName, objectName, dst, opts)
}

func (c *BucketClient) Read(ctx context.Context, objectName string, dst any, opts minio.GetObjectOptions, s ...Serializer) error {
	ser := c.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	return c.Client.Read(ctx, c.BucketName, objectName, dst, opts, ser)
}
