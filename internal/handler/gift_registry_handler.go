package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type GiftRegistryHandler struct {
	giftRegistryService service.GiftRegistryService
}

func NewGiftRegistryHandler(giftRegistryService service.GiftRegistryService) *GiftRegistryHandler {
	return &GiftRegistryHandler{giftRegistryService: giftRegistryService}
}

type GiftRegistryRequest struct {
	Type           string `json:"type" validate:"required"`
	ProviderName   string `json:"provider_name"`
	AccountNumber  string `json:"account_number"`
	AccountName    string `json:"account_name"`
	QRCodeImage    string `json:"qr_code_image"`
	PhoneNumber    string `json:"phone_number"`
	MailingAddress string `json:"mailing_address"`
}

func (h *GiftRegistryHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	registries, err := h.giftRegistryService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve gift registries", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Gift registries retrieved successfully", registries, nil)
}

func (h *GiftRegistryHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req GiftRegistryRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	registry, err := h.giftRegistryService.Create(project.ID, req.Type, req.ProviderName, req.AccountNumber, req.AccountName, req.QRCodeImage, req.PhoneNumber, req.MailingAddress)
	if err != nil {
		if err.Error() == "invalid gift registry type" ||
			err.Error() == "bank registry requires account_number and account_name" ||
			err.Error() == "ewallet registry requires either a QRIS image or a phone_number, and provider_name" ||
			err.Error() == "physical registry requires mailing_address" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create gift registry", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Gift registry created successfully", registry, nil)
}

func (h *GiftRegistryHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	registryIdParam := c.Params("registryId")
	registryID, err := strconv.ParseUint(registryIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid registry ID", "INVALID_ID")
	}

	var req GiftRegistryRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	registry, err := h.giftRegistryService.Update(uint(registryID), project.ID, req.Type, req.ProviderName, req.AccountNumber, req.AccountName, req.QRCodeImage, req.PhoneNumber, req.MailingAddress)
	if err != nil {
		if err.Error() == "gift registry not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		if err.Error() == "invalid gift registry type" ||
			err.Error() == "bank registry requires account_number and account_name" ||
			err.Error() == "ewallet registry requires either a QRIS image or a phone_number, and provider_name" ||
			err.Error() == "physical registry requires mailing_address" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update gift registry", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Gift registry updated successfully", registry, nil)
}

func (h *GiftRegistryHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	registryIdParam := c.Params("registryId")
	registryID, err := strconv.ParseUint(registryIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid registry ID", "INVALID_ID")
	}

	if err := h.giftRegistryService.Delete(uint(registryID), project.ID); err != nil {
		if err.Error() == "gift registry not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete gift registry", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Gift registry deleted successfully", nil, nil)
}
