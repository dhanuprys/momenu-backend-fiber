package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type IPCheckerService interface {
	GetCountryFromIP(ip string) (string, error)
}

type ipApiService struct {
	client *http.Client
}

func NewIPCheckerService() IPCheckerService {
	return &ipApiService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type ipApiResponse struct {
	Status  string `json:"status"`
	Country string `json:"country"`
}

func (s *ipApiService) GetCountryFromIP(ip string) (string, error) {
	// Skip check for loopback or empty
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return "Unknown", nil
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country", ip)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ip checker api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data ipApiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if data.Status != "success" {
		return "Unknown", nil
	}

	return data.Country, nil
}
