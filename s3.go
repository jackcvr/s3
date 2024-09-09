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

func (c *Client) NewBucketManager(bucketName string, serializer Serializer) *BucketManager {
	if serializer == nil {
		serializer = c.Serializer
	}
	return &BucketManager{
		Client:     c,
		BucketName: bucketName,
		Serializer: serializer,
	}
}

type BucketManager struct {
	Client     *Client
	BucketName string
	Serializer Serializer
}

func (m *BucketManager) EnsureBucket(ctx context.Context, opts minio.MakeBucketOptions) error {
	exists, err := m.Client.BucketExists(ctx, m.BucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return m.Client.MakeBucket(ctx, m.BucketName, opts)
}

func (m *BucketManager) RemoveBucket(ctx context.Context) error {
	return m.Client.RemoveBucket(ctx, m.BucketName)
}

func (m *BucketManager) ListObjects(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return m.Client.ListObjects(ctx, m.BucketName, opts)
}

func (m *BucketManager) StatObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (minio.ObjectInfo, error) {
	return m.Client.StatObject(ctx, m.BucketName, objectName, opts)
}

func (m *BucketManager) GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return m.Client.GetObject(ctx, m.BucketName, objectName, opts)
}

func (m *BucketManager) RemoveObject(ctx context.Context, objectName string, opts minio.RemoveObjectOptions) error {
	return m.Client.RemoveObject(ctx, m.BucketName, objectName, opts)
}

func (m *BucketManager) PutBytes(ctx context.Context, objectName string, value []byte, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.Client.PutBytes(ctx, m.BucketName, objectName, value, opts)
}

func (m *BucketManager) Put(ctx context.Context, objectName string, value any, opts minio.PutObjectOptions, s ...Serializer) (minio.UploadInfo, error) {
	ser := m.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	return m.Client.Put(ctx, m.BucketName, objectName, value, opts, ser)
}

func (m *BucketManager) ReadBytes(ctx context.Context, objectName string, dst io.Writer, opts minio.GetObjectOptions) error {
	return m.Client.ReadBytes(ctx, m.BucketName, objectName, dst, opts)
}

func (m *BucketManager) Read(ctx context.Context, objectName string, dst any, opts minio.GetObjectOptions, s ...Serializer) error {
	ser := m.Serializer
	if len(s) > 0 {
		ser = s[0]
	}
	return m.Client.Read(ctx, m.BucketName, objectName, dst, opts, ser)
}
