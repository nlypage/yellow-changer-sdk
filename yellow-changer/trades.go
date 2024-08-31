package yellowChanger

import (
	"context"
	"encoding/json"
	"github.com/nlypage/yellow-changer-sdk/yellow-changer/common/errorz"
	"net/http"
)

type withdrawNetwork struct {
	Network     string  `json:"network"`
	Fee         float64 `json:"fee"`
	MinWithdraw float64 `json:"min_withdraw"`
}

type Rate struct {
	Currency         string             `json:"currency"`
	Name             string             `json:"name"`
	WithdrawNetworks []withdrawNetwork  `json:"withdraw_networks"`
	ConversionRates  map[string]float64 `json:"conversion_rates"`
}

// AllRates returns all possible exchange rates using /trades/allRates endpoint.
func (c *Client) AllRates(ctx context.Context) (*[]Rate, error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: "trades/allRates",
	}

	data, err := c.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	var rates *[]Rate
	if err = json.Unmarshal(data, &rates); err != nil {
		return nil, err
	}
	return rates, nil
}

// RatesInDirection returns rates in a certain direction using /trades/ratesInDirection endpoint.
func (c *Client) RatesInDirection(ctx context.Context, direction string) (*Rate, error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: "trades/ratesInDirection",
		Body: struct {
			Direction string `json:"direction"`
		}{
			Direction: direction,
		},
	}

	data, err := c.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	var rate *Rate
	if err = json.Unmarshal(data, &rate); err != nil {
		return nil, err
	}
	return rate, nil
}

type Limits struct {
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
}

type commission struct {
	FeeAmount float64 `json:"fee_amount"`
}

type PayInDestination struct {
	Currency string `json:"currency"`
	Network  string `json:"network"`
	Limit    Limits `json:"limit"`
}

type PayOutDestination struct {
	PayInDestination
	Commission commission `json:"commission"`
}

type DestinationsList struct {
	PayIn  []PayInDestination  `json:"payin"`
	PayOut []PayOutDestination `json:"payout"`
}

// DestinationsList returns all possible exchange destinations using /trades/destinationsList endpoint.
func (c *Client) DestinationsList(ctx context.Context) (*DestinationsList, error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: "trades/destinationsList",
	}

	data, err := c.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	var destinations *DestinationsList
	if err = json.Unmarshal(data, &destinations); err != nil {
		return nil, err
	}
	return destinations, nil
}

// GetCurrencyLimits returns the limits for the selected currency for sending or withdrawing depending on the payIn value, respectively.
func (c *Client) GetCurrencyLimits(ctx context.Context, currency string, network string, payIn bool) (*Limits, error) {
	destinationsList, err := c.DestinationsList(ctx)
	if err != nil {
		return nil, err
	}
	if payIn {
		for _, currencyData := range destinationsList.PayIn {
			if currencyData.Currency == currency && currencyData.Network == network {
				return &currencyData.Limit, nil
			}
		}
	} else {
		for _, currencyData := range destinationsList.PayOut {
			if currencyData.Currency == currency && currencyData.Network == network {
				return &currencyData.Limit, nil
			}
		}
	}
	return nil, errorz.CurrencyNotFound
}

// CreateTrade is a structure that represents /trades/createTrade request body.
type CreateTrade struct {
	// SendName is a currency being sent.
	SendName string `json:"send_name"`
	//GetName is a currency being received.
	//
	// If GetName, GetCreds and GetNetwork are "empty", the funds will be credited to the site balance.
	GetName string `json:"get_name"`
	// SendNetwork is a network of currency being sent.
	SendNetwork string `json:"send_network"`
	// GetNetwork is a network of currency being received.
	//
	// If GetName, GetCreds and GetNetwork are "empty", the funds will be credited to the site balance.
	GetNetwork string `json:"get_network"`
	// SendValue is amount of the exchange you will pay.
	SendValue float64 `json:"send_value,omitempty"`
	// GetValue is amount of the exchange you will receive.
	GetValue float64 `json:"get_value,omitempty"`
	// GetCreds is a credentials (address, card or number).
	//
	// If GetName, GetCreds and GetNetwork are "empty", the funds will be credited to the site balance.
	GetCreds string `json:"get_creds"`
	// UniqID is optional. Unique id of trade.
	UniqID string `json:"uniq_id,omitempty"`
	// SbpBank is optional (needs only when get_network is SBPRUB). Bank for transfer via SBP.
	/*
		- Acceptable values:

		sbppsb, sbpakbars, sbprnkb, sbpotp, sbpozon, sbpmtc, sbppochtabank, sbpumoney, sbptinkoff, sbpsber, sbpraif, sbpalfa, sbpotkritie, sbpvtb, sbpsovkombank, sbpgazprom,sbprosbank
	*/
	SbpBank string `json:"sbp_bank,omitempty"`
}

// Trade is a struct that represents a trade in yellow-changer.
type Trade struct {
	SendName    string `json:"send_name"`
	SendNetwork string `json:"send_network"`
	GetNetwork  string `json:"get_network"`
	UniqId      string `json:"uniq_id"`
	/* Status is a status of trade.

	1 - waiting for payment from the customer.
	2 - waiting for network confirmations.
	3 - sent by us.
	4 - canceled.
	5 - AML block.
	*/
	Status            int    `json:"status"`
	PaymentWallet     string `json:"payment_wallet"`
	UserPaidHash      string `json:"userPaidHash"`
	OurHash           string `json:"ourHash"`
	GetCreds          string `json:"get_creds"`
	NetworkCommission int    `json:"network_commission"`
	Date              int    `json:"date"`
	TimeExpire        int    `json:"time_expire"`
	SendValue         string `json:"send_value"`
	GetValue          string `json:"get_value"`
}

// CreateTrade creates new trade and returns trade info using /trades/createTrade endpoint.
func (c *Client) CreateTrade(ctx context.Context, createTrade CreateTrade) (*Trade, error) {
	r := &Request{
		Method:   http.MethodPost,
		Endpoint: "trades/createTrade",
		Body:     createTrade,
	}

	data, err := c.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	var trade *Trade
	if err = json.Unmarshal(data, &trade); err != nil {
		return nil, err
	}
	return trade, nil
}

// GetInfo returns all information about transaction using /trades/getInfo endpoint.
func (c *Client) GetInfo(ctx context.Context, uniqID string) (*Trade, error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: "trades/getInfo",
		Body: struct {
			UniqID string `json:"uniq_id"`
		}{
			UniqID: uniqID,
		},
	}

	data, err := c.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	var trade *Trade
	if err = json.Unmarshal(data, &trade); err != nil {
		return nil, err
	}
	return trade, nil
}
