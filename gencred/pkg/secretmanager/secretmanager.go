/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secretmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
)

// Client is a wrapper of secretmanager client
type Client struct {
	// ProjectID is GCP project in which to store secrets in Secret Manager.
	ProjectID string
	client    *secretmanager.Client
	dryrun    bool
}

// ClientInterface is the interface for manipulating secretmanager
type ClientInterface interface {
	CreateSecret(ctx context.Context, secretID string) (*secretmanagerpb.Secret, error)
	AddSecretVersion(ctx context.Context, secretName string, payload []byte) error
	ListSecrets(ctx context.Context) ([]*secretmanagerpb.Secret, error)
}

// NewClient creates a client for secretmanager, it would fail if not authenticated
func NewClient(projectID string, dryrun bool) (*Client, error) {
	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}
	return &Client{ProjectID: projectID, client: client, dryrun: dryrun}, nil
}

// CreateSecret creates a secret
func (c *Client) CreateSecret(ctx context.Context, secretID string) (*secretmanagerpb.Secret, error) {
	if c.dryrun {
		return nil, nil
	}
	// Create the request to create the secret.
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", c.ProjectID),
		SecretId: secretID,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	return c.client.CreateSecret(ctx, createSecretReq)
}

// AddSecretVersion adds a secret version, aka update the value of a secret
func (c *Client) AddSecretVersion(ctx context.Context, secretName string, payload []byte) error {
	if c.dryrun {
		return nil
	}
	// Build the request.
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", c.ProjectID, secretName),
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API.
	_, err := c.client.AddSecretVersion(ctx, addSecretVersionReq)
	return err
}

// ListSecrets lists all secrets under current project
func (c *Client) ListSecrets(ctx context.Context) ([]*secretmanagerpb.Secret, error) {
	var res []*secretmanagerpb.Secret
	// Build the request.
	listRequest := &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", c.ProjectID),
	}

	// Call the API.
	it := c.client.ListSecrets(ctx, listRequest)
	for {
		s, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}
