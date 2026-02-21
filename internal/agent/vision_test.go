package agent

import (
	"context"
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
