// Package customer provides methods to use customers via the dwolla api.
package customer

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/ahmedaabouzied/dwolla-go/dwolla/client"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/funding"
	"github.com/ahmedaabouzied/dwolla-go/dwolla/transfer"
	"github.com/pkg/errors"
)

// Customer represents an individual or business with whom you intend to transact with
type Customer struct {
	Client       client.DwollaClient
	ID           string                 `json:"id"`
	FirstName    string                 `json:"firstName"`
	LastName     string                 `json:"lastName"`
	Email        string                 `json:"email"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	BusinessName string                 `json:"businessName,omitempty"`
	IPAddress    string                 `json:"ipAddress"`
	CreatedAt    string                 `json:"created"`
	DateOfBirth  string                 `json:"dateOfBirth,omitempty"`
	SSN          string                 `json:"ssn,omitempty"`
	State        string                 `json:"state"`
	PostalCode   string                 `json:"postalCode"`
	City         string                 `json:"city"`
	Address      string                 `json:"address1,omitempty"`
	Passport     string                 `json:"passport,omitempty"`
	Links        map[string]client.Link `json:"_links"`
}

// Document is a file sumbitted to dwolla to be validated
type Document struct {
	Client    client.DwollaClient
	Links     map[string]client.Link `json:"_links"`
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Type      string                 `json:"passport"`
	CreatedAt string                 `json:"created"`
}

type listCustomersResponse struct {
	Links    map[string]client.Link `json:"_links"`
	Embedded map[string][]Customer  `json:"_embedded"`
}

type listDocumentsResponse struct {
	Links    map[string]client.Link `json:"_links"`
	Embedded map[string][]Document  `json:"_embedded"`
}

type createFudingSourceToken struct {
	Token string                       `json:"token"`
	Links map[string]map[string]string `json:"_links"`
}

// Create a new customer
func Create(c client.DwollaClient, cu *Customer) (string, error) {
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get auth token")
	}
	body, err := json.Marshal(cu)
	if err != nil {
		return "", errors.Wrap(err, "error marshalling customer into req body")
	}
	req, err := http.NewRequest("POST", c.RootURL()+"/customers", bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Content-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 201:
		return strings.TrimPrefix(res.Header.Get("Location"), c.RootURL()+"/customers/"), nil
	case 403:
		return "", errors.New("not authorized to create customers")
	case 400:
		io.Copy(os.Stdout, res.Body)
		return "", errors.New("duplicate customer or validation error")
	case 404:
		return "", errors.New("account not found")
	default:
		return "", errors.New(res.Status)
	}
}

// List retrieves a list of created customers
func List(c client.DwollaClient) ([]Customer, error) {
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", c.RootURL()+"/customers", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &listCustomersResponse{}
		err = d.Decode(body)
		for _, customer := range body.Embedded["customers"] {
			customer.Client = c
		}
		return body.Embedded["customers"], nil
	case 403:
		return nil, errors.New("not authorized to list customers")
	case 404:
		return nil, errors.New("account not found")
	default:
		return nil, errors.New(res.Status)
	}
}

// GetCustomer retrieves a customer belonging to the authorized Dwolla Master Account by it's ID
func GetCustomer(c client.DwollaClient, customerID string) (*Customer, error) {
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", c.RootURL()+"/customers/"+customerID, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &Customer{}
		err = d.Decode(body)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing JSON response")
		}
		body.Client = c
		return body, nil
	case 403:
		return nil, errors.New("not authorized to retrieve the customer")
	case 404:
		return nil, errors.New("account not found")
	default:
		return nil, errors.New(res.Status)
	}
}

// Update can be used to achieve the following :
// update Customer information,
// upgrade an unverified Customer to a verified Customer,
// suspend a Customer, deactivate a Customer,
// reactivate a Customer,
// and update a verified Customer’s information to retry verification.
func (cu *Customer) Update() error {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return errors.Wrap(err, "failed to get auth token")
	}
	body, err := json.Marshal(cu)
	if err != nil {
		return errors.Wrap(err, "error marshalling customer into req body")
	}
	req, err := http.NewRequest("POST", c.RootURL()+"/customers/"+cu.ID, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Content-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		return nil
	case 403:
		return errors.New("not authorized to update the customer")
	case 404:
		return errors.New("account not found")
	default:
		return errors.New(res.Status)
	}

}

// TODO : Add ListBusinessClassification Method

// TODO : Add RetrieveBusinessClassification Method

// AddDocument uploads a document to a customer for verification
func (cu *Customer) AddDocument(file *os.File, documentType string) error {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return errors.Wrap(err, "failed to get auth token")
	}
	if err != nil {
		return errors.Wrap(err, "error parsing file")
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	err = writer.WriteField("documentType", documentType)
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	writer.Close()
	req, err := http.NewRequest("POST", c.RootURL()+"/customers/+"+cu.ID+"/documents", body)
	if err != nil {
		return errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
	req.Header.Add("Cache-Control", "no-cache")
	res, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 201:
		return nil
	case 403:
		return errors.New("not authorized to uplaod document to customer")
	case 404:
		return errors.New("account not found")
	default:
		return errors.New(res.Status)
	}

}

// ListDocuments retrieves documents submitted to be validated for this customer
func (cu *Customer) ListDocuments() ([]Document, error) {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", c.RootURL()+"/customers/"+cu.ID+"/documents", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &listDocumentsResponse{}
		err = d.Decode(body)
		return body.Embedded["documents"], nil
	case 403:
		return nil, errors.New("not authorized to list customers")
	case 404:
		return nil, errors.New("account not found")
	default:
		return nil, errors.New(res.Status)
	}
}

// TODO : Add CreateDocumentForBenificialOwner method.

// TODO : Add ListDocumentsForBenificialOwner method.

// GetDocument retrieves a docuemnt by ID
func GetDocument(c client.DwollaClient, docuemntID string) (*Document, error) {
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", c.RootURL()+"/documents/"+docuemntID, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &Document{}
		err = d.Decode(body)
		return body, nil
	case 403:
		return nil, errors.New("not authorized to retrieve the customer")
	case 404:
		return nil, errors.New("account not found")
	default:
		return nil, errors.New(res.Status)
	}
}

// CreateFundingSource creates a funding source for a customer
func (cu *Customer) CreateFundingSource(f *funding.Resource) error {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return errors.Wrap(err, "failed to get auth token")
	}
	body, err := json.Marshal(f)
	if err != nil {
		return errors.Wrap(err, "error marshalling customer into req body")
	}
	req, err := http.NewRequest("POST", c.RootURL()+"/customers/"+cu.ID+"/funding-sources", bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Content-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 201:
		return nil
	case 403:
		return errors.New("not authorized to create funding source")
	case 400:
		return errors.New("duplicate funding source or validation error. Authorization already associated to a funding source")
	default:
		return errors.New(res.Status)
	}
}

// CreateFundingSourceToken creates a new funding source from a token via dwolla.js
func (cu *Customer) CreateFundingSourceToken() (string, error) {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get auth token")
	}

	req, err := http.NewRequest("POST", c.RootURL()+"/customers/"+cu.ID+"/funding-sources-token", nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Content-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &createFudingSourceToken{}
		err = d.Decode(body)
		return body.Token, nil
	case 404:
		return "", errors.New("customer not found")
	default:
		return "", errors.New(res.Status)
	}
}

// CreateIAVFundingSourceToken creates a token to add and verify
func (cu *Customer) CreateIAVFundingSourceToken() (string, error) {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("POST", c.RootURL()+"/customers/"+cu.ID+"/iav-token", nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Conetent-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &createFudingSourceToken{}
		err = d.Decode(body)
		return body.Token, nil
	case 404:
		return "", errors.New("customer not found")
	default:
		return "", errors.New(res.Status)
	}
}

// ListFundingSources retrieves funding sources that belong to the customer.
func (cu *Customer) ListFundingSources() ([]funding.Resource, error) {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", c.RootURL()+"/customers/"+cu.ID+"/funding-sources", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	req.Header.Add("Conetent-Type", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &funding.ListResourcesResponse{}
		err = d.Decode(body)
		var sources []funding.Resource
		for _, s := range body.Embedded["funding-sources"] {
			source := funding.Resource{
				Client:          cu.Client,
				ID:              s.ID,
				Status:          s.Status,
				AccountNumber:   s.AccountNumber,
				RoutingNumber:   s.RoutingNumber,
				BankAccountType: s.BankAccountType,
				Name:            s.Name,
				BankName:        s.BankName,
				Type:            s.Type,
				CreatedAt:       s.CreatedAt,
				Removed:         s.Removed,
				PlaidToken:      s.PlaidToken,
				Channels:        s.Channels,
				Links:           s.Links,
			}
			sources = append(sources, source)
		}
		return sources, nil
	case 403:
		return nil, errors.New("not authorized to list funding sources")
	case 404:
		return nil, errors.New("customer not found")
	default:
		return nil, errors.New(res.Status)
	}
}

// ListTransfers retrieves the customer's list of transfers.
func (cu *Customer) ListTransfers() ([]transfer.Transfer, error) {
	var c = cu.Client
	hc := &http.Client{}
	token, err := c.AuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auth token")
	}
	req, err := http.NewRequest("GET", cu.Links["self"].Href+"/transfers", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.dwolla.v1.hal+json")
	res, err := hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to dwolla api")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		d := json.NewDecoder(res.Body)
		body := &transfer.ListTransferResponse{}
		err = d.Decode(body)
		return body.Embedded["transfers"], nil
	case 403:
		return nil, errors.New("not authorized to list transfers")
	case 404:
		return nil, errors.New("customer not found")
	default:
		return nil, errors.New(res.Status)
	}
}
