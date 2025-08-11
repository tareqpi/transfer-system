package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/tareqpi/transfer-system/internal/domain"
	"github.com/tareqpi/transfer-system/internal/service"
)

var errTest = errors.New("assert error")

type fakeService struct {
	createAccountFunc func(domain.Account) (*domain.Account, error)
	getAccountFunc    func(string) (*domain.Account, error)
	transferMoneyFunc func(domain.Transaction) error
}

func (m fakeService) CreateAccount(ctx context.Context, account domain.Account) (*domain.Account, error) {
	return m.createAccountFunc(account)
}
func (m fakeService) GetAccount(ctx context.Context, accountID string) (*domain.Account, error) {
	return m.getAccountFunc(accountID)
}
func (m fakeService) TransferMoney(ctx context.Context, transaction domain.Transaction) error {
	return m.transferMoneyFunc(transaction)
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), Recovery())
	handler := NewHandler(fakeService{
		createAccountFunc: func(account domain.Account) (*domain.Account, error) { return &account, nil },
		getAccountFunc: func(accountID string) (*domain.Account, error) {
			return &domain.Account{ID: 7, Balance: decimal.RequireFromString("42.50")}, nil
		},
		transferMoneyFunc: func(transaction domain.Transaction) error { return nil },
	})
	router.POST("/api/v1/accounts", handler.CreateAccount)
	router.GET("/api/v1/accounts/:account_id", handler.GetAccount)
	router.POST("/api/v1/transactions", handler.TransferMoney)
	return router
}

func TestCreateAccount_Success(t *testing.T) {
	router := newTestRouter()

	requestBody := `{"account_id": 1, "initial_balance": "100.00"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
	if requestID := recorder.Header().Get(headerRequestID); requestID == "" {
		t.Fatalf("expected %s header to be set", headerRequestID)
	}
}

func TestCreateAccount_InvalidRequest(t *testing.T) {
	router := newTestRouter()
	requestBody := `{"account_id": 1, "initial_balance": "not-a-number"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerRequestID, "test-rid-1")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d. body=%s headers=%v", recorder.Code, recorder.Body.String(), recorder.Header())
	}

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.Error.Code != "invalid_request" {
		t.Fatalf("expected error code invalid_request, got %s", response.Error.Code)
	}
	if response.RequestID == "" || recorder.Header().Get(headerRequestID) == "" {
		t.Fatalf("expected request id to be present in body and header")
	}
}

func TestCreateAccount_InternalError(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Recovery())
	handler := NewHandler(fakeService{createAccountFunc: func(account domain.Account) (*domain.Account, error) { return nil, errTest }, getAccountFunc: nil, transferMoneyFunc: nil})
	router.POST("/api/v1/accounts", handler.CreateAccount)

	requestBody := `{"account_id": 1, "initial_balance": "100.00"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.Error.Code != "internal_error" {
		t.Fatalf("expected error code internal_error, got %s", response.Error.Code)
	}
}

func TestGetAccount_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/7", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response AccountResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.AccountID != 7 {
		t.Fatalf("expected account_id 7, got %d", response.AccountID)
	}
	if response.Balance.Round(2).StringFixed(2) != "42.50" {
		t.Fatalf("expected balance 42.50, got %s", response.Balance.String())
	}
}

func TestGetAccount_InternalError(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Recovery())
	handler := NewHandler(fakeService{getAccountFunc: func(accountID string) (*domain.Account, error) { return nil, errTest }})
	router.GET("/api/v1/accounts/:account_id", handler.GetAccount)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/99", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
}

func TestTransferMoney_Success(t *testing.T) {
	router := newTestRouter()

	requestBody := `{"source_account_id": 1, "destination_account_id": 2, "amount": "25.50"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestTransferMoney_BadRequests(t *testing.T) {
	testCases := []struct {
		testName           string
		serviceError       error
		expectedStatusCode int
		expectedErrorCode  string
	}{
		{"same_account", service.ErrSameSourceAndDestination, http.StatusBadRequest, "same_account"},
		{"invalid_amount", service.ErrNonPositiveAmount, http.StatusBadRequest, "invalid_amount"},
		{"invalid_account_ids", service.ErrInvalidAccountIDs, http.StatusBadRequest, "invalid_account_ids"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestID(), Recovery())
			handler := NewHandler(fakeService{transferMoneyFunc: func(transaction domain.Transaction) error { return testCase.serviceError }})
			router.POST("/api/v1/transactions", handler.TransferMoney)

			requestBody := `{"source_account_id": 1, "destination_account_id": 2, "amount": "10"}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(headerRequestID, "rid-123")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != testCase.expectedStatusCode {
				t.Fatalf("expected status %d, got %d", testCase.expectedStatusCode, recorder.Code)
			}
			var response ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if response.Error.Code != testCase.expectedErrorCode {
				t.Fatalf("expected error code %s, got %s", testCase.expectedErrorCode, response.Error.Code)
			}
			if response.RequestID != "rid-123" {
				t.Fatalf("expected request_id to echo header, got %s", response.RequestID)
			}
		})
	}
}

func TestTransferMoney_InsufficientBalance(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Recovery())
	handler := NewHandler(fakeService{transferMoneyFunc: func(domain.Transaction) error { return service.ErrInsufficientBalance }})
	router.POST("/api/v1/transactions", handler.TransferMoney)

	requestBody := `{"source_account_id": 1, "destination_account_id": 2, "amount": "1000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.Error.Code != "insufficient_balance" {
		t.Fatalf("expected error code insufficient_balance, got %s", response.Error.Code)
	}
}

func TestTransferMoney_InternalError(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Recovery())
	handler := NewHandler(fakeService{transferMoneyFunc: func(domain.Transaction) error { return errTest }})
	router.POST("/api/v1/transactions", handler.TransferMoney)

	requestBody := `{"source_account_id": 1, "destination_account_id": 2, "amount": "1000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.Error.Code != "internal_error" {
		t.Fatalf("expected error code internal_error, got %s", response.Error.Code)
	}
}
