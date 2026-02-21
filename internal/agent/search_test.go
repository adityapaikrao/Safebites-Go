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
