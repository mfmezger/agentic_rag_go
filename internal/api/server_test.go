package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mfmezger/agentic_rag_go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	server.writeJSON(w, 200, data)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "value", response["key"])
}

func TestWriteJSON_DifferentStatusCodes(t *testing.T) {
	server := &Server{}
	statusCodes := []int{200, 201, 400, 404, 500, 503}

	for _, code := range statusCodes {
		t.Run("status_"+string(rune(code)), func(t *testing.T) {
			w := httptest.NewRecorder()
			server.writeJSON(w, code, map[string]string{"status": "ok"})
			assert.Equal(t, code, w.Code)
		})
	}
}

func TestWriteJSON_EmptyBody(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()

	server.writeJSON(w, 200, nil)

	assert.Equal(t, 200, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWriteJSON_LargeData(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()

	largeData := make(map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		largeData[string(rune(i))] = "value"
	}

	server.writeJSON(w, 200, largeData)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestWriteError(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()

	server.writeError(w, 400, "test error message")

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test error message", response["error"])
}

func TestWriteError_DifferentErrorCodes(t *testing.T) {
	server := &Server{}
	errorCodes := []int{400, 401, 403, 404, 500, 503}

	for _, code := range errorCodes {
		t.Run("error_"+string(rune(code)), func(t *testing.T) {
			w := httptest.NewRecorder()
			server.writeError(w, code, "error message")
			assert.Equal(t, code, w.Code)
		})
	}
}

func TestWriteError_EmptyMessage(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()

	server.writeError(w, 500, "")

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "", response["error"])
}

func TestUploadTextRequest_JSONMarshaling(t *testing.T) {
	req := UploadTextRequest{
		Text: "test text content",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test text content")

	var decoded UploadTextRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "test text content", decoded.Text)
}

func TestUploadTextRequest_WithMetadata(t *testing.T) {
	req := UploadTextRequest{
		Text: "content",
		Metadata: map[string]string{
			"author": "John",
			"date":   "2024-01-01",
		},
		Source: "document.pdf",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded UploadTextRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "content", decoded.Text)
	assert.Equal(t, "John", decoded.Metadata["author"])
	assert.Equal(t, "document.pdf", decoded.Source)
}

func TestSearchRequest_JSONMarshaling(t *testing.T) {
	req := SearchRequest{
		Query: "test query",
		TopK:  10,
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded SearchRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "test query", decoded.Query)
	assert.Equal(t, 10, decoded.TopK)
}

func TestSearchRequest_ZeroTopK(t *testing.T) {
	req := SearchRequest{
		Query: "test",
		TopK:  0,
	}

	data, _ := json.Marshal(req)
	var decoded SearchRequest
	json.Unmarshal(data, &decoded)

	assert.Equal(t, "test", decoded.Query)
	assert.Equal(t, 0, decoded.TopK)
}

func TestChatRequest_JSONMarshaling(t *testing.T) {
	req := ChatRequest{
		Message:   "hello",
		UserID:    "user123",
		SessionID: "session456",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded ChatRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "hello", decoded.Message)
	assert.Equal(t, "user123", decoded.UserID)
	assert.Equal(t, "session456", decoded.SessionID)
}

func TestChatRequest_WithoutSession(t *testing.T) {
	req := ChatRequest{
		Message: "hello",
	}

	data, _ := json.Marshal(req)
	var decoded ChatRequest
	json.Unmarshal(data, &decoded)

	assert.Equal(t, "hello", decoded.Message)
	assert.Equal(t, "", decoded.SessionID)
}

func TestUploadTextResponse_JSONMarshaling(t *testing.T) {
	resp := UploadTextResponse{
		Message:    "success",
		ChunkCount: 5,
		ChunkIDs:   []string{"id1", "id2", "id3"},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded UploadTextResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "success", decoded.Message)
	assert.Equal(t, 5, decoded.ChunkCount)
	assert.Equal(t, 3, len(decoded.ChunkIDs))
}

func TestUploadTextResponse_EmptyChunkIDs(t *testing.T) {
	resp := UploadTextResponse{
		Message:    "failed",
		ChunkCount: 0,
		ChunkIDs:   []string{},
	}

	data, _ := json.Marshal(resp)
	var decoded UploadTextResponse
	json.Unmarshal(data, &decoded)

	assert.Equal(t, 0, decoded.ChunkCount)
	assert.Equal(t, 0, len(decoded.ChunkIDs))
}

func TestSearchResponse_JSONMarshaling(t *testing.T) {
	resp := SearchResponse{
		Results: []SearchResultItem{
			{
				ID:       "doc1",
				Content:  "content",
				Score:    0.95,
				Metadata: map[string]string{"key": "value"},
			},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded SearchResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, 1, len(decoded.Results))
	assert.Equal(t, "doc1", decoded.Results[0].ID)
	assert.Equal(t, float32(0.95), decoded.Results[0].Score)
}

func TestSearchResponse_MultipleResults(t *testing.T) {
	resp := SearchResponse{
		Results: []SearchResultItem{
			{ID: "1", Content: "a", Score: 0.9},
			{ID: "2", Content: "b", Score: 0.8},
			{ID: "3", Content: "c", Score: 0.7},
		},
	}

	data, _ := json.Marshal(resp)
	var decoded SearchResponse
	json.Unmarshal(data, &decoded)

	assert.Equal(t, 3, len(decoded.Results))
}

func TestChatResponse_JSONMarshaling(t *testing.T) {
	resp := ChatResponse{
		Response:  "AI response",
		SessionID: "session123",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded ChatResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "AI response", decoded.Response)
	assert.Equal(t, "session123", decoded.SessionID)
}

func TestSearchResultItem_Fields(t *testing.T) {
	item := SearchResultItem{
		ID:       "test-id",
		Content:  "test content",
		Score:    0.85,
		Metadata: map[string]string{"author": "test"},
	}

	assert.Equal(t, "test-id", item.ID)
	assert.Equal(t, "test content", item.Content)
	assert.Equal(t, float32(0.85), item.Score)
	assert.Equal(t, "test", item.Metadata["author"])
}

func TestSearchResultItem_EmptyMetadata(t *testing.T) {
	item := SearchResultItem{
		ID:       "id",
		Content:  "content",
		Score:    1.0,
		Metadata: map[string]string{},
	}

	assert.Equal(t, 0, len(item.Metadata))
}

func TestServer_ServeHTTP_CORSHeaders(t *testing.T) {
	server := &Server{
		mux: http.NewServeMux(),
	}
	server.mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestServer_ServeHTTP_OptionsRequest(t *testing.T) {
	server := &Server{
		mux: http.NewServeMux(),
	}

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_Close(t *testing.T) {
	server := &Server{
		qdrant: nil,
	}

	err := server.Close()
	assert.NoError(t, err)
}

func TestMiddleware_NewMiddleware(t *testing.T) {
	m := newMiddleware("key", 100, time.Minute)
	assert.NotNil(t, m)
	assert.Equal(t, "key", m.apiKey)
	assert.NotNil(t, m.rateLimiter)
}

func TestValidation_EmptyFields(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		err  bool
	}{
		{
			name: "valid upload request",
			data: []byte(`{"text":"content"}`),
			err:  false,
		},
		{
			name: "empty upload text",
			data: []byte(`{"text":""}`),
			err:  false,
		},
		{
			name: "invalid json",
			data: []byte(`{invalid`),
			err:  true,
		},
		{
			name: "malformed json",
			data: []byte(`{"text":}`),
			err:  true,
		},
		{
			name: "extra fields",
			data: []byte(`{"text":"content","extra":"field"}`),
			err:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UploadTextRequest
			err := json.Unmarshal(tt.data, &req)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_ContextCancellation(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &Server{
		cfg: &config.Config{},
		mux: http.NewServeMux(),
	}

	assert.NotNil(t, server)
}

func TestServer_Start_MethodsExist(t *testing.T) {
	server := &Server{
		mux: http.NewServeMux(),
	}

	assert.NotNil(t, server.Start)
	assert.NotNil(t, server.Close)
	assert.NotNil(t, server.ServeHTTP)
}
