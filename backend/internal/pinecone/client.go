package pinecone

import (
	"context"
	"fmt"
	"strings"

	"github.com/pbearc/github-agent/backend/pkg/common"
	pc "github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client wraps the Pinecone client
type Client struct {
	client     *pc.Client
	indexName  string
	dimensions int
	logger     *common.Logger
}

// NewClient creates a new Pinecone client
func NewClient(apiKey, environment, indexName string, dimensions int) (*Client, error) {
	if apiKey == "" {
		return nil, common.NewError("Pinecone API key is required")
	}

	if indexName == "" {
		return nil, common.NewError("Pinecone index name is required")
	}

	// Create Pinecone client
	pcClient, err := pc.NewClient(pc.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, common.WrapError(err, "failed to create Pinecone client")
	}

	return &Client{
		client:     pcClient,
		indexName:  indexName,
		dimensions: dimensions,
		logger:     common.NewLogger(),
	}, nil
}

// EnsureIndex ensures that the index exists
func (c *Client) EnsureIndex(ctx context.Context) error {
	// Check if index exists
	_, err := c.client.DescribeIndex(ctx, c.indexName)
	
	// If the index exists, nothing to do
	if err == nil {
		c.logger.Info(fmt.Sprintf("Index %s already exists", c.indexName))
		return nil
	}
	
	// Check if it's a "not found" error, otherwise return the error
	if !strings.Contains(err.Error(), "not found") {
		return common.WrapError(err, "failed to check if index exists")
	}

	// If the index doesn't exist, create it
	vectorType := "dense"
	dimension := int32(c.dimensions)
	metric := pc.Cosine
	deletionProtection := pc.DeletionProtectionDisabled

	_, err = c.client.CreateServerlessIndex(ctx, &pc.CreateServerlessIndexRequest{
		Name:               c.indexName,
		VectorType:         &vectorType,
		Dimension:          &dimension,
		Metric:             &metric,
		Cloud:              pc.Aws,
		Region:             "us-east-1",
		DeletionProtection: &deletionProtection,
	})
	if err != nil {
		return common.WrapError(err, "failed to create index")
	}

	c.logger.Info(fmt.Sprintf("Created index %s, waiting for it to be ready...", c.indexName))
	
	return nil
}

// getIndexConnection gets a connection to the index for a specific namespace
func (c *Client) getIndexConnection(ctx context.Context, namespace string) (*pc.IndexConnection, error) {
	// Get the index model first
	indexModel, err := c.client.DescribeIndex(ctx, c.indexName)
	if err != nil {
		return nil, common.WrapError(err, "failed to describe index")
	}

	// Create a connection to the index with the specified namespace
	indexConn, err := c.client.Index(pc.NewIndexConnParams{
		Host:      indexModel.Host,
		Namespace: namespace,
	})
	if err != nil {
		return nil, common.WrapError(err, "failed to create index connection")
	}

	return indexConn, nil
}

// Vector represents a vector in Pinecone
type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Upsert inserts or updates vectors in the index
func (c *Client) Upsert(ctx context.Context, vectors []Vector, namespace string) (int, error) {
	// Get a connection to the index
	indexConn, err := c.getIndexConnection(ctx, namespace)
	if err != nil {
		return 0, err
	}

	// Convert our Vector type to Pinecone Vector type
	pcVectors := make([]*pc.Vector, len(vectors))
	for i, v := range vectors {
		// Convert metadata to structpb.Struct
		metadata, err := structpb.NewStruct(v.Metadata)
		if err != nil {
			return 0, common.WrapError(err, "failed to convert metadata for vector")
		}
		
		pcVectors[i] = &pc.Vector{
			Id:       v.ID,
			Values:   &v.Values,
			Metadata: metadata,
		}
	}

	// Upsert the vectors
	count, err := indexConn.UpsertVectors(ctx, pcVectors)
	if err != nil {
		return 0, common.WrapError(err, "failed to upsert vectors")
	}

	return int(count), nil
}

// QueryRequest is the request for querying vectors
type QueryRequest struct {
	Vector          []float32              `json:"vector"`
	TopK            int                    `json:"topK"`
	Namespace       string                 `json:"namespace,omitempty"`
	IncludeMetadata bool                   `json:"includeMetadata"`
	Filter          map[string]interface{} `json:"filter,omitempty"`
}

// QueryMatch represents a match in a query response
type QueryMatch struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// QueryResponse is the response for querying vectors
type QueryResponse struct {
	Matches   []QueryMatch `json:"matches"`
	Namespace string       `json:"namespace"`
}

// Query queries vectors in the index
func (c *Client) Query(ctx context.Context, queryReq QueryRequest) (*QueryResponse, error) {
	// Get a connection to the index
	indexConn, err := c.getIndexConnection(ctx, queryReq.Namespace)
	if err != nil {
		return nil, err
	}

	topK := uint32(queryReq.TopK)
	
	// Create Pinecone query request
	pcQueryReq := &pc.QueryByVectorValuesRequest{
		Vector:          queryReq.Vector,
		TopK:            topK,
		IncludeValues:   true,
		IncludeMetadata: queryReq.IncludeMetadata,
	}

	// Execute the query
	pcResponse, err := indexConn.QueryByVectorValues(ctx, pcQueryReq)
	if err != nil {
		return nil, common.WrapError(err, "failed to query vectors")
	}

	// Convert Pinecone response to our response type
	response := &QueryResponse{
		Namespace: queryReq.Namespace,
		Matches:   make([]QueryMatch, len(pcResponse.Matches)),
	}

	for i, match := range pcResponse.Matches {
		metadata := make(map[string]interface{})
		if match.Vector.Metadata != nil {
			metadata = match.Vector.Metadata.AsMap()
		}
		
		response.Matches[i] = QueryMatch{
			ID:       match.Vector.Id,
			Score:    match.Score,
			Metadata: metadata,
		}
	}

	return response, nil
}

// DeleteRequest is the request for deleting vectors
type DeleteRequest struct {
	IDs       []string `json:"ids,omitempty"`
	Namespace string   `json:"namespace,omitempty"`
	DeleteAll bool     `json:"deleteAll,omitempty"`
}

// Delete deletes vectors from the index
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
    // Get a connection to the index
    indexConn, err := c.getIndexConnection(ctx, req.Namespace)
    if err != nil {
        return err
    }

    if req.DeleteAll {
        // Delete all vectors in the namespace
        err = indexConn.DeleteAllVectorsInNamespace(ctx)
        if err != nil {
            return common.WrapError(err, "failed to delete all vectors in namespace")
        }
        return nil
    }

    if len(req.IDs) > 0 {
        // Delete specified IDs
        err = indexConn.DeleteVectorsById(ctx, req.IDs)
        if err != nil {
            return common.WrapError(err, "failed to delete vectors by ID")
        }
        return nil
    }

    return common.NewError("either deleteAll or IDs must be specified for deletion")
}
// NamespaceStats represents stats about a namespace
type NamespaceStats struct {
	VectorCount int `json:"vectorCount"`
}

// IndexStats represents stats about an index
type IndexStats struct {
	Dimension        int                       `json:"dimension"`
	TotalVectorCount int                       `json:"totalVectorCount"`
	Namespaces       map[string]NamespaceStats `json:"namespaces"`
}

// DescribeIndexStats gets statistics about the index
func (c *Client) DescribeIndexStats(ctx context.Context) (*IndexStats, error) {
	// Get a connection to the index
	indexConn, err := c.getIndexConnection(ctx, "")
	if err != nil {
		return nil, err
	}

	// Get stats
	pcStats, err := indexConn.DescribeIndexStats(ctx)
	if err != nil {
		return nil, common.WrapError(err, "failed to describe index stats")
	}

	// Convert to our stats structure
	stats := &IndexStats{
		Dimension:        c.dimensions, // Use the dimension we know, as pcStats might not have it
		TotalVectorCount: int(pcStats.TotalVectorCount),
		Namespaces:       make(map[string]NamespaceStats),
	}

	for ns, nsStats := range pcStats.Namespaces {
		stats.Namespaces[ns] = NamespaceStats{
			VectorCount: int(nsStats.VectorCount),
		}
	}

	return stats, nil
}