package msgpack

import "github.com/vmihailenco/msgpack/v5"

type MsgPackSerializer struct{}

func (_ MsgPackSerializer) Serialize(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (_ MsgPackSerializer) Deserialize(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
