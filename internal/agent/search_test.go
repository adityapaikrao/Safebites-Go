package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchAgentSearch(t *testing.T) {
	fake := newFakeLLM(`{"List_of_ingredients":[{"name":"Water","description":"Solvent"}]}`)
	a, err := NewSearchAgent(fake)
	require.NoError(t, err)

	out, err := a.Search(context.Background(), "Sparkling Water")
	require.NoError(t, err)
	require.Len(t, out.ListOfIngredients, 1)
	require.Equal(t, "Water", out.ListOfIngredients[0].Name)
	require.Len(t, fake.requests, 1)
}

func TestSearchAgentEmptyProductName(t *testing.T) {
	fake := newFakeLLM(`{"List_of_ingredients":[]}`)
	a, err := NewSearchAgent(fake)
	require.NoError(t, err)

	_, err = a.Search(context.Background(), "   ")
	require.Error(t, err)
	require.ErrorContains(t, err, "product name is required")
}

func TestSearchAgentMalformedJSON(t *testing.T) {
	fake := newFakeLLM(`not-json`)
	a, err := NewSearchAgent(fake)
	require.NoError(t, err)

	_, err = a.Search(context.Background(), "Sparkling Water")
	require.Error(t, err)
	require.ErrorContains(t, err, "parse search result")
}

func TestSearchAgentCodeFenceJSON(t *testing.T) {
	fake := newFakeLLM("```json\n{\"List_of_ingredients\":[{\"name\":\"Water\",\"description\":\"Solvent\"}]}\n```")
	a, err := NewSearchAgent(fake)
	require.NoError(t, err)

	out, err := a.Search(context.Background(), "Sparkling Water")
	require.NoError(t, err)
	require.Len(t, out.ListOfIngredients, 1)
	require.Equal(t, "Water", out.ListOfIngredients[0].Name)
}
