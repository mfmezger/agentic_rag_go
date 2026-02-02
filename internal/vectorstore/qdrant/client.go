// Package qdrant provides a Qdrant vector store client with hybrid search support.
package qdrant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a Qdrant vector store client.
type Client struct {
	conn        *grpc.ClientConn
	points      pb.PointsClient
	collections pb.CollectionsClient
}

// Config holds Qdrant client configuration.
type Config struct {
	Host       string
	GRPCPort   int
	Collection string
	VectorSize uint64
}

// New creates a new Qdrant client.
func New(ctx context.Context, cfg Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPCPort)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	client := &Client{
		conn:        conn,
		points:      pb.NewPointsClient(conn),
		collections: pb.NewCollectionsClient(conn),
	}

	return client, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// EnsureCollection creates the collection if it doesn't exist.
// Sets up for hybrid search with dense and sparse vectors.
func (c *Client) EnsureCollection(ctx context.Context, name string, vectorSize uint64) error {
	// Check if collection exists
	exists, err := c.collections.CollectionExists(ctx, &pb.CollectionExistsRequest{
		CollectionName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if exists.GetResult().GetExists() {
		return nil
	}

	// Create collection with named vectors for hybrid search
	_, err = c.collections.Create(ctx, &pb.CreateCollection{
		CollectionName: name,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_ParamsMap{
				ParamsMap: &pb.VectorParamsMap{
					Map: map[string]*pb.VectorParams{
						"dense": {
							Size:     vectorSize,
							Distance: pb.Distance_Cosine,
						},
					},
				},
			},
		},
		SparseVectorsConfig: &pb.SparseVectorConfig{
			Map: map[string]*pb.SparseVectorParams{
				"sparse": {},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// Document represents a document to be stored.
type Document struct {
	ID       string
	Content  string
	Metadata map[string]string
	Dense    []float32
	Sparse   *SparseVector
}

// SparseVector represents a sparse vector for BM25-style search.
type SparseVector struct {
	Indices []uint32
	Values  []float32
}

// Upsert inserts or updates documents in the collection.
func (c *Client) Upsert(ctx context.Context, collection string, docs []Document) error {
	points := make([]*pb.PointStruct, len(docs))

	for i, doc := range docs {
		// Generate UUID if not provided
		id := doc.ID
		if id == "" {
			id = uuid.New().String()
		}

		// Build payload
		payload := make(map[string]*pb.Value)
		payload["content"] = &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: doc.Content},
		}
		for k, v := range doc.Metadata {
			payload[k] = &pb.Value{
				Kind: &pb.Value_StringValue{StringValue: v},
			}
		}

		// Build vectors
		vectors := &pb.Vectors{
			VectorsOptions: &pb.Vectors_Vectors{
				Vectors: &pb.NamedVectors{
					Vectors: map[string]*pb.Vector{
						"dense": {
							Data: doc.Dense,
						},
					},
				},
			},
		}

		point := &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: id},
			},
			Vectors: vectors,
			Payload: payload,
		}

		// Add sparse vector if provided
		if doc.Sparse != nil {
			point.Vectors = &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vectors{
					Vectors: &pb.NamedVectors{
						Vectors: map[string]*pb.Vector{
							"dense": {Data: doc.Dense},
							"sparse": {
								Data:    doc.Sparse.Values,
								Indices: &pb.SparseIndices{Data: doc.Sparse.Indices},
							},
						},
					},
				},
			}
		}

		points[i] = point
	}

	_, err := c.points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collection,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}

	return nil
}

// SearchResult represents a search result.
type SearchResult struct {
	ID      string
	Score   float32
	Content string
	Payload map[string]string
}

// HybridSearch performs hybrid search with dense and sparse vectors.
func (c *Client) HybridSearch(ctx context.Context, collection string, denseVector []float32, sparseVector *SparseVector, topK uint64) ([]SearchResult, error) {
	// Build prefetch queries
	prefetch := []*pb.PrefetchQuery{
		{
			Query: &pb.Query{
				Variant: &pb.Query_Nearest{
					Nearest: &pb.VectorInput{
						Variant: &pb.VectorInput_Dense{
							Dense: &pb.DenseVector{Data: denseVector},
						},
					},
				},
			},
			Using: strPtr("dense"),
			Limit: &topK,
		},
	}

	// Add sparse search if provided
	if sparseVector != nil {
		prefetch = append(prefetch, &pb.PrefetchQuery{
			Query: &pb.Query{
				Variant: &pb.Query_Nearest{
					Nearest: &pb.VectorInput{
						Variant: &pb.VectorInput_Sparse{
							Sparse: &pb.SparseVector{
								Indices: sparseVector.Indices,
								Values:  sparseVector.Values,
							},
						},
					},
				},
			},
			Using: strPtr("sparse"),
			Limit: &topK,
		})
	}

	// Fusion query using RRF (Reciprocal Rank Fusion)
	limit := topK
	resp, err := c.points.Query(ctx, &pb.QueryPoints{
		CollectionName: collection,
		Prefetch:       prefetch,
		Query: &pb.Query{
			Variant: &pb.Query_Fusion{
				Fusion: pb.Fusion_RRF,
			},
		},
		Limit:       &limit,
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	results := make([]SearchResult, len(resp.Result))
	for i, point := range resp.Result {
		result := SearchResult{
			Score:   point.Score,
			Payload: make(map[string]string),
		}

		// Extract ID
		if uuid := point.Id.GetUuid(); uuid != "" {
			result.ID = uuid
		}

		// Extract payload
		for k, v := range point.Payload {
			if sv := v.GetStringValue(); sv != "" {
				if k == "content" {
					result.Content = sv
				} else {
					result.Payload[k] = sv
				}
			}
		}

		results[i] = result
	}

	return results, nil
}

func strPtr(s string) *string {
	return &s
}
