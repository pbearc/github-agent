package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pbearc/github-agent/backend/internal/models"
	"github.com/pbearc/github-agent/backend/pkg/common"
)

// Neo4jClient handles interactions with the Neo4j database
type Neo4jClient struct {
	driver  neo4j.Driver
	uri     string
	logger  *common.Logger
}

// NewNeo4jClient creates a new Neo4j client
func NewNeo4jClient(uri, username, password string) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	err = driver.VerifyConnectivity()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Neo4jClient{
		driver: driver,
		uri:    uri,
		logger: common.NewLogger(),
	}, nil
}

// Close closes the Neo4j driver
func (c *Neo4jClient) Close() error {
	return c.driver.Close()
}

// StoreCodebaseStructure stores the codebase structure in Neo4j
func (c *Neo4jClient) StoreCodebaseStructure(ctx context.Context, owner, repo, branch string, 
                                             files []models.GitHubFile, importMap map[string][]string) error {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	// Clear existing data for this repository
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (r:Repository {owner: $owner, name: $repo})
			OPTIONAL MATCH (r)-[:HAS_BRANCH]->(b:Branch {name: $branch})
			OPTIONAL MATCH (b)-[*]->(n)
			DETACH DELETE n
			WITH r, b
			DETACH DELETE b
			RETURN r
		`
		_, err := tx.Run(query, map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"branch": branch,
		})
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Create repository and branch nodes
	_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MERGE (r:Repository {owner: $owner, name: $repo})
			MERGE (r)-[:HAS_BRANCH]->(b:Branch {name: $branch})
			RETURN r, b
		`
		_, err := tx.Run(query, map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"branch": branch,
		})
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create repository structure: %w", err)
	}

	// Create file nodes
	for _, file := range files {
		_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			query := `
				MATCH (b:Branch {name: $branch})-[:HAS_BRANCH]-(r:Repository {owner: $owner, name: $repo})
				MERGE (b)-[:CONTAINS]->(f:File {path: $path, name: $name, type: $type})
				RETURN f
			`
			_, err := tx.Run(query, map[string]interface{}{
				"owner":  owner,
				"repo":   repo,
				"branch": branch,
				"path":   file.Path,
				"name":   file.Name,
				"type":   file.Type,
			})
			return nil, err
		})

		if err != nil {
			c.logger.WithField("error", err).Warning("Failed to create file node: " + file.Path)
		}
	}

	// Create import relationships
	for source, targets := range importMap {
		for _, target := range targets {
			_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				query := `
					MATCH (b:Branch {name: $branch})-[:HAS_BRANCH]-(r:Repository {owner: $owner, name: $repo})
					MATCH (b)-[:CONTAINS]->(source:File {path: $source})
					MATCH (b)-[:CONTAINS]->(target:File {path: $target})
					MERGE (source)-[:IMPORTS]->(target)
					RETURN source, target
				`
				_, err := tx.Run(query, map[string]interface{}{
					"owner":  owner,
					"repo":   repo,
					"branch": branch,
					"source": source,
					"target": target,
				})
				return nil, err
			})

			if err != nil {
				c.logger.WithField("error", err).Warning(fmt.Sprintf("Failed to create import relationship: %s -> %s", source, target))
			}
		}
	}

	return nil
}

// GetCodebaseGraph retrieves the codebase graph from Neo4j
func (c *Neo4jClient) GetCodebaseGraph(ctx context.Context, owner, repo, branch string) (map[string]interface{}, error) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (b:Branch {name: $branch})-[:HAS_BRANCH]-(r:Repository {owner: $owner, name: $repo})
			MATCH (b)-[:CONTAINS]->(f:File)
			OPTIONAL MATCH (f)-[rel:IMPORTS]->(f2:File)
			WITH f, collect({target: f2.path, type: type(rel)}) as relationships
			RETURN {
				nodes: collect({id: f.path, label: f.name, type: f.type, path: f.path}),
				relationships: collect({source: f.path, targets: relationships})
			} as graph
		`
		result, err := tx.Run(query, map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"branch": branch,
		})

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().GetByIndex(0), nil
		}

		return map[string]interface{}{
			"nodes":         []interface{}{},
			"relationships": []interface{}{},
		}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get codebase graph: %w", err)
	}

	return result.(map[string]interface{}), nil
}