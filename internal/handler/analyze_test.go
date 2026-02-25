package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/middleware"
	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockAnalyzeService struct {
	analyze func(ctx context.Context, imageBytes []byte, mimeType string, prefs *model.UserPreferences) (string, *model.ScorerResult, error)
}

func (m *mockAnalyzeService) Analyze(ctx context.Context, imageBytes []byte, mimeType string, prefs *model.UserPreferences) (string, *model.ScorerResult, error) {
	return m.analyze(ctx, imageBytes, mimeType, prefs)
}

type mockAnalyzeUserService struct {
	getByID           func(ctx context.Context, userID string) (*model.User, error)
	upsert            func(ctx context.Context, user *model.User) (*model.User, error)
	updatePreferences func(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error)
	applyTemplate     func(ctx context.Context, userID string, templateKey string) (*model.User, model.DietaryTemplate, error)
}

func (m *mockAnalyzeUserService) GetByID(ctx context.Context, userID string) (*model.User, error) {
	if m.getByID == nil {
		return nil, nil
	}
	return m.getByID(ctx, userID)
}

func (m *mockAnalyzeUserService) Upsert(ctx context.Context, user *model.User) (*model.User, error) {
	if m.upsert == nil {
		return nil, nil
	}
	return m.upsert(ctx, user)
}

func (m *mockAnalyzeUserService) UpdatePreferences(ctx context.Context, userID string, preferences model.UserPreferences) (*model.User, error) {
	if m.updatePreferences == nil {
		return nil, nil
	}
	return m.updatePreferences(ctx, userID, preferences)
}

func (m *mockAnalyzeUserService) ApplyTemplate(ctx context.Context, userID string, templateKey string) (*model.User, model.DietaryTemplate, error) {
	if m.applyTemplate == nil {
		return nil, model.DietaryTemplate{}, nil
	}
	return m.applyTemplate(ctx, userID, templateKey)
}

func makeAnalyzeMultipartRequest(t *testing.T, withImage bool) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if withImage {
		part, err := writer.CreateFormFile("image", "sample.png")
		require.NoError(t, err)
		_, err = part.Write([]byte("fake-image-bytes"))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestAnalyzeHandlerAnalyzeImageSuccessWithoutUser(t *testing.T) {
	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, imageBytes []byte, mimeType string, prefs *model.UserPreferences) (string, *model.ScorerResult, error) {
				require.NotEmpty(t, imageBytes)
				require.NotEmpty(t, mimeType)
				require.Nil(t, prefs)
				return "Product A", &model.ScorerResult{OverallScore: 7.8}, nil
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, true)
	rr := httptest.NewRecorder()

	h.AnalyzeImage(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"status":"success"`)
	require.Contains(t, rr.Body.String(), `"product_name":"Product A"`)
	require.Contains(t, rr.Body.String(), `"overall_score":7.8`)
}

func TestAnalyzeHandlerAnalyzeImageSuccessWithUserPreferences(t *testing.T) {
	userToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "auth0|user-1"})
	tokenString, err := userToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, _ []byte, _ string, prefs *model.UserPreferences) (string, *model.ScorerResult, error) {
				require.NotNil(t, prefs)
				require.Equal(t, []string{"vegan"}, prefs.DietGoals)
				return "Product B", &model.ScorerResult{OverallScore: 6.5}, nil
			},
		},
		Users: &mockAnalyzeUserService{
			getByID: func(_ context.Context, userID string) (*model.User, error) {
				require.Equal(t, "auth0|user-1", userID)
				return &model.User{ID: userID, DietGoals: []string{"vegan"}}, nil
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, true)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	middleware.OptionalAuth(&config.Config{})(http.HandlerFunc(h.AnalyzeImage)).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAnalyzeHandlerAnalyzeImageMissingImage(t *testing.T) {
	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, _ []byte, _ string, _ *model.UserPreferences) (string, *model.ScorerResult, error) {
				t.Fatal("analyze should not be called")
				return "", nil, nil
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, false)
	rr := httptest.NewRecorder()

	h.AnalyzeImage(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "image file is required")
}

func TestAnalyzeHandlerAnalyzeImageServiceError(t *testing.T) {
	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, _ []byte, _ string, _ *model.UserPreferences) (string, *model.ScorerResult, error) {
				return "", nil, errors.New("analyze failed")
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, true)
	rr := httptest.NewRecorder()

	h.AnalyzeImage(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAnalyzeHandlerAnalyzeImageUserServiceError(t *testing.T) {
	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, _ []byte, _ string, _ *model.UserPreferences) (string, *model.ScorerResult, error) {
				t.Fatal("analyze should not be called when user lookup fails")
				return "", nil, nil
			},
		},
		Users: &mockAnalyzeUserService{
			getByID: func(_ context.Context, _ string) (*model.User, error) {
				return nil, errors.New("db down")
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, true)
	userToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "user-1"})
	tokenString, err := userToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	middleware.OptionalAuth(&config.Config{})(http.HandlerFunc(h.AnalyzeImage)).ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAnalyzeHandlerAnalyzeImageUserNotFoundStillAnalyzes(t *testing.T) {
	h := &AnalyzeHandler{
		Analyze: &mockAnalyzeService{
			analyze: func(_ context.Context, _ []byte, _ string, prefs *model.UserPreferences) (string, *model.ScorerResult, error) {
				require.Nil(t, prefs)
				return "Product C", &model.ScorerResult{OverallScore: 5.1}, nil
			},
		},
		Users: &mockAnalyzeUserService{
			getByID: func(_ context.Context, _ string) (*model.User, error) {
				return nil, repository.ErrNotFound
			},
		},
	}

	req := makeAnalyzeMultipartRequest(t, true)
	userToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "user-1"})
	tokenString, err := userToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	middleware.OptionalAuth(&config.Config{})(http.HandlerFunc(h.AnalyzeImage)).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}
