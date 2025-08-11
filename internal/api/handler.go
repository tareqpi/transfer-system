package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/tareqpi/transfer-system/internal/domain"
	"github.com/tareqpi/transfer-system/internal/logger"
	"github.com/tareqpi/transfer-system/internal/service"
	"go.uber.org/zap"
)

type CreateAccountRequest struct {
	AccountID      int64           `json:"account_id" binding:"required"`
	InitialBalance decimal.Decimal `json:"initial_balance" binding:"required"`
}

type AccountResponse struct {
	AccountID int64           `json:"account_id" binding:"required"`
	Balance   decimal.Decimal `json:"balance" binding:"required"`
}

type TransferMoneyRequest struct {
	SourceAccountID      int64           `json:"source_account_id" binding:"required"`
	DestinationAccountID int64           `json:"destination_account_id" binding:"required"`
	Amount               decimal.Decimal `json:"amount" binding:"required"`
}

type Handler struct {
	Service service.Service
}

func NewHandler(applicationService service.Service) *Handler {
	return &Handler{Service: applicationService}
}

func (handler *Handler) CreateAccount(c *gin.Context) {
	var request CreateAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "invalid_request", err.Error())
		return
	}

	_, err := handler.Service.CreateAccount(c.Request.Context(), domain.Account{
		ID:      request.AccountID,
		Balance: request.InitialBalance,
	})
	if err != nil {
		logger.L().Error("create account failed", zap.Error(err))
		Internal(c, http.StatusText(http.StatusInternalServerError))
		return
	}
	c.Status(http.StatusCreated)
}

func (handler *Handler) GetAccount(c *gin.Context) {
	accountID := c.Param("account_id")
	account, err := handler.Service.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		logger.L().Error("get account failed", zap.Error(err), zap.String("account_id", accountID))
		Internal(c, http.StatusText(http.StatusInternalServerError))
		return
	}
	accountResponse := AccountResponse{
		AccountID: account.ID,
		Balance:   account.Balance,
	}
	c.JSON(http.StatusOK, accountResponse)
}

func (handler *Handler) TransferMoney(c *gin.Context) {
	var request TransferMoneyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "invalid_request", err.Error())
		return
	}

	err := handler.Service.TransferMoney(c.Request.Context(), domain.Transaction{
		SourceAccountID:      request.SourceAccountID,
		DestinationAccountID: request.DestinationAccountID,
		Amount:               request.Amount,
	})

	if err != nil {
		switch {
		case errors.Is(err, service.ErrSameSourceAndDestination):
			BadRequest(c, "same_account", err.Error())
		case errors.Is(err, service.ErrNonPositiveAmount):
			BadRequest(c, "invalid_amount", err.Error())
		case errors.Is(err, service.ErrInvalidAccountIDs):
			BadRequest(c, "invalid_account_ids", err.Error())
		case errors.Is(err, service.ErrInsufficientBalance):
			Conflict(c, "insufficient_balance", err.Error())
		default:
			logger.L().Error("transfer failed", zap.Error(err), zap.Any("request", request))
			Internal(c, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	c.Status(http.StatusOK)
}
