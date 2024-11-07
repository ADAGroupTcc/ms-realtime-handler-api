package sorterApi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/http"
)

type SorterApi interface {
	Sort(ctx context.Context, userId string) (*domain.SortResponse, error)
}

type sorterApi struct {
	sorterClient http.HttpClienter
}

func New(sorterClient http.HttpClienter) SorterApi {
	return &sorterApi{
		sorterClient,
	}
}

func (s *sorterApi) Sort(ctx context.Context, userId string) (*domain.SortResponse, error) {
	url := fmt.Sprintf("/v1/search?user_id=%s", userId)
	response, err := s.sorterClient.Get(ctx, http.ClientConfig{
		Endpoint: url,
	})

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error sorting: %s", string(response.Body))
	}

	var sortResponse *domain.SortResponse
	err = json.Unmarshal(response.Body, &sortResponse)
	if err != nil {
		return nil, err
	}

	return sortResponse, nil
}
