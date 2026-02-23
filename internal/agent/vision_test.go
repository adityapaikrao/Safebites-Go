package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVisionOCRExtractProductName(t *testing.T) {
	v := NewVisionOCR(&fakeVisionClient{text: "  Oatly Oat Milk  "})

	name, err := v.ExtractProductName(context.Background(), []byte("img"), "image/heic")
	require.NoError(t, err)
	require.Equal(t, "Oatly Oat Milk", name)
}

func TestVisionOCREmptyImage(t *testing.T) {
	v := NewVisionOCR(&fakeVisionClient{text: "ignored"})
	_, err := v.ExtractProductName(context.Background(), nil, "image/jpeg")
	require.Error(t, err)
}

func TestVisionOCRClientError(t *testing.T) {
	v := NewVisionOCR(&fakeVisionClient{err: errors.New("gemini unavailable")})

	_, err := v.ExtractProductName(context.Background(), []byte("img"), "image/jpeg")
	require.Error(t, err)
	require.ErrorContains(t, err, "gemini unavailable")
}

func TestVisionOCREmptyResponse(t *testing.T) {
	v := NewVisionOCR(&fakeVisionClient{text: "  "})

	_, err := v.ExtractProductName(context.Background(), []byte("img"), "image/jpeg")
	require.Error(t, err)
	require.ErrorContains(t, err, "vision response is empty")
}

func TestVisionOCRUnsupportedMimeType(t *testing.T) {
	v := NewVisionOCR(&fakeVisionClient{text: "ignored"})

	_, err := v.ExtractProductName(context.Background(), []byte("img"), "image/bmp")
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported image mime type")
}
