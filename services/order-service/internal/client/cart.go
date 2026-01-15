package client

import (
	"context"
	"net/http"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CartClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

type CartItem struct {
    ID         uuid.UUID `json:"id"`
    ProductID  uuid.UUID `json:"product_id"`
    Quantity   int32     `json:"quantity"`
    PriceCents int32     `json:"price_cents"`
}

type Cart struct {
    ID         uuid.UUID  `json:"id"`
    Items      []CartItem `json:"items"`
    TotalCents int64      `json:"total_cents"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
    return &Client{
        BaseURL:    baseURL,
        HTTPClient: &http.Client{Timeout: timeout},
    }
}

func (c *CartClient) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, bool, error) {
	if userID == uuid.Nil {
		return nil, false, fmt.Errorf("invalid UUID: nil")
	}

	url := fmt.Sprintf("%s/internal/cart/%s", c.BaseURL, userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("error getting the response: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("calling cart service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var cart Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, false, fmt.Errorf("decoding response: %w", err)
	}
	return &cart, true, nil
}

func (c *Client) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("invalid UUID: nil")
	}
	url := fmt.Sprintf("%s/internal/cart/%s", c.BaseURL, userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling cart service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
