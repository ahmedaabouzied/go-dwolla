// Copyrights 2019 Ahmed Abouzied.
// All rights reserved.

// Package dwolla is a go client library for dwolla v2 rest API.
package dwolla

import (
	"github.com/ahmedaabouzied/dwolla-go/dwolla/account"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/client"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/customer"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/funding"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/transfer"
)

const (
	// Production represents a dwolla production environemnt
	Production string = "production"
	// Sandbox represents a dwolla sandbox environment
	Sandbox string = "sandbox"
)

// Client wraps the client.Client to dwolla.Client
type Client struct {
	client *client.Client
}

// CreateClient creates a new dwolla client.
func CreateClient(env string, clientID string, clientSecret string) (*Client, error) {
	client, err := client.CreateClient(env, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

// RetrieveAccount returns the dwolla master account.
func (c *Client) RetrieveAccount() (*account.Account, error) {
	return account.RetrieveAccount(c.client)
}

// CreateCustomer creates a new customer.
func (c *Client) CreateCustomer(cu *customer.Customer) error {
	return customer.Create(c.client, cu)
}

// ListCustomers retrieves a list of created customers.
func (c *Client) ListCustomers() ([]customer.Customer, error) {
	return customer.List(c.client)
}

// GetCustomer retrieves a customer by ID.
func (c *Client) GetCustomer(customerID string) (*customer.Customer, error) {
	return customer.GetCustomer(c.client, customerID)
}

// GetDocument retrieves a document by ID.
func (c *Client) GetDocument(documentID string) (*customer.Document, error) {
	return customer.GetDocument(c.client, documentID)
}

// GetFundingSource retrieves a funding source by ID.
func (c *Client) GetFundingSource(sourceID string) (*funding.Resource, error) {
	return funding.GetFundingSource(c.client, sourceID)
}

// CreateTransfer creates a transfer between two funding sources
func (c *Client) CreateTransfer(t *transfer.Transfer) error {
	return transfer.CreateTransfer(c.client, t)
}

// GetTransfer retrieves a transfer by it's ID.
func (c *Client) GetTransfer(transferID string) (*transfer.Transfer, error) {
	return transfer.GetTransfer(c.client, transferID)
}

// CreateOnDemandAuth creates an on-demand token.
func (c *Client) CreateOnDemandAuth() (string, error) {
	return transfer.CreateOnDemandAuth(c.client)
}
