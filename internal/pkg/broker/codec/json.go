package codec

import "encoding/json"

type jsonCodec[T any] struct{}

func NewCodecJson[T any]() Codec[T] {
	return jsonCodec[T]{}
}

func (jsonCodec[T]) Encode(v *T) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (jsonCodec[T]) Decode(data []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
