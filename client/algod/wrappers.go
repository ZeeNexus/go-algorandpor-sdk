package algod

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ZeeNexus/go-algorandpor-sdk/client/algod/models"
)

// Status retrieves the StatusResponse from the running node
// the StatusResponse includes data like the consensus version and current round
func (client Client) Status(headers ...*Header) (response models.NodeStatus, err error) {
	err = client.get(&response, "/status", nil, headers)
	return
}

// HealthCheck does a health check on the the potentially running node,
// returning an error if the API is down
func (client Client) HealthCheck(headers ...*Header) error {
	return client.get(nil, "/health", nil, headers)
}

// StatusAfterBlock waits for a block to occur then returns the StatusResponse after that block
// blocks on the node end
func (client Client) StatusAfterBlock(blockNum uint64, headers ...*Header) (response models.NodeStatus, err error) {
	err = client.get(&response, fmt.Sprintf("/status/wait-for-block-after/%d", blockNum), nil, headers)
	return
}

type pendingTransactionsParams struct {
	Max uint64 `url:"max"`
}

// GetPendingTransactions asks algod for a snapshot of current pending txns on the node, bounded by maxTxns.
// If maxTxns = 0, fetches as many transactions as possible.
func (client Client) GetPendingTransactions(maxTxns uint64, headers ...*Header) (response models.PendingTransactions, err error) {
	err = client.get(&response, fmt.Sprintf("/transactions/pending"), pendingTransactionsParams{maxTxns}, headers)
	return
}

// Versions retrieves the VersionResponse from the running node
// the VersionResponse includes data like version number and genesis ID
func (client Client) Versions(headers ...*Header) (response models.Version, err error) {
	err = client.get(&response, "/versions", nil, headers)
	return
}

// LedgerSupply gets the supply details for the specified node's Ledger
func (client Client) LedgerSupply(headers ...*Header) (response models.Supply, err error) {
	err = client.get(&response, "/ledger/supply", nil, headers)
	return
}

type transactionsByAddrParams struct {
	FirstRound uint64 `url:"firstRound,omitempty"`
	LastRound  uint64 `url:"lastRound,omitempty"`
	FromDate   string `url:"fromDate,omitempty"`
	ToDate     string `url:"toDate,omitempty"`
	Max        uint64 `url:"max,omitempty"`
}

// TransactionsByAddr returns all transactions for a PK [addr] in the [first,
// last] rounds range.
func (client Client) TransactionsByAddr(addr string, first, last uint64, headers ...*Header) (response models.TransactionList, err error) {
	params := transactionsByAddrParams{FirstRound: first, LastRound: last}
	err = client.get(&response, fmt.Sprintf("/account/%s/transactions", addr), params, headers)
	return
}

// TransactionsByAddrLimit returns the last [limit] number of transaction for a PK [addr].
func (client Client) TransactionsByAddrLimit(addr string, limit uint64, headers ...*Header) (response models.TransactionList, err error) {
	params := transactionsByAddrParams{Max: limit}
	err = client.get(&response, fmt.Sprintf("/account/%s/transactions", addr), params, headers)
	return
}

// TransactionsByAddrForDate returns all transactions for a PK [addr] in the [first,
// last] date range. Dates are of the form "2006-01-02".
func (client Client) TransactionsByAddrForDate(addr string, first, last string, headers ...*Header) (response models.TransactionList, err error) {
	params := transactionsByAddrParams{FromDate: first, ToDate: last}
	err = client.get(&response, fmt.Sprintf("/account/%s/transactions", addr), params, headers)
	return
}

// AccountInformation also gets the AccountInformationResponse associated with the passed address
func (client Client) AccountInformation(address string, headers ...*Header) (response models.Account, err error) {
	err = client.get(&response, fmt.Sprintf("/account/%s", address), nil, headers)
	return
}

// Accounts gets the AccountList response
func (client Client) AccountList(headers ...*Header) (response models.AccountList, err error) {
	params := transactionsByAddrParams{Max: 10}
	err = client.get(&response, fmt.Sprintf("/account/accountlist"), params, headers)
	return
}

// AssetInformation also gets the AssetInformationResponse associated with the passed asset creator and index
func (client Client) AssetInformation(creator string, index uint64, headers ...*Header) (response models.AssetParams, err error) {
	err = client.get(&response, fmt.Sprintf("/account/%s/assets/%d", creator, index), nil, headers)
	return
}

// TransactionInformation gets information about a specific transaction involving a specific account
// it will only return information about transactions submitted to the node queried
func (client Client) TransactionInformation(accountAddress, transactionID string, headers ...*Header) (response models.Transaction, err error) {
	transactionID = stripTransaction(transactionID)
	err = client.get(&response, fmt.Sprintf("/account/%s/transaction/%s", accountAddress, transactionID), nil, headers)
	return
}

// PendingTransactionInformation gets information about a recently issued
// transaction.  There are several cases when this might succeed:
//
// - transaction committed (CommittedRound > 0)
// - transaction still in the pool (CommittedRound = 0, PoolError = "")
// - transaction removed from pool due to error (CommittedRound = 0, PoolError != "")
//
// Or the transaction may have happened sufficiently long ago that the
// node no longer remembers it, and this will return an error.
func (client Client) PendingTransactionInformation(transactionID string, headers ...*Header) (response models.Transaction, err error) {
	transactionID = stripTransaction(transactionID)
	err = client.get(&response, fmt.Sprintf("/transactions/pending/%s", transactionID), nil, headers)
	return
}

// TransactionByID gets a transaction by its ID. Works only if the indexer is enabled on the node
// being queried.
func (client Client) TransactionByID(transactionID string, headers ...*Header) (response models.Transaction, err error) {
	transactionID = stripTransaction(transactionID)
	err = client.get(&response, fmt.Sprintf("/transaction/%s", transactionID), nil, headers)
	return
}

// SuggestedFee gets the recommended transaction fee from the node
func (client Client) SuggestedFee(headers ...*Header) (response models.TransactionFee, err error) {
	err = client.get(&response, "/transactions/fee", nil, headers)
	return
}

// SuggestedParams gets the suggested transaction parameters
func (client Client) SuggestedParams(headers ...*Header) (response models.TransactionParams, err error) {
	err = client.get(&response, "/transactions/params", nil, headers)
	return
}

// SendRawTransaction gets the bytes of a SignedTxn and broadcasts it to the network
func (client Client) SendRawTransaction(stx []byte, headers ...*Header) (response models.TransactionID, err error) {
	err = client.post(&response, "/transactions", stx, headers)
	return
}

// Block gets the block info for the given round
func (client Client) Block(round uint64, headers ...*Header) (response models.Block, err error) {
	err = client.get(&response, fmt.Sprintf("/block/%d", round), nil, headers)
	return
}

func (client Client) doGetWithQuery(ctx context.Context, path string, queryArgs map[string]string) (result string, err error) {
	queryURL := client.serverURL
	queryURL.Path = path

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	for k, v := range queryArgs {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set(authHeader, client.apiToken)
	for _, header := range client.headers {
		req.Header.Add(header.Key, header.Value)
	}

	httpClient := http.Client{}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = extractError(resp)
	if err != nil {
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result = string(bytes)
	return
}
