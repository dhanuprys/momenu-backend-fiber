package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type GiftRegistryService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.GiftRegistry, error)
	Create(projectID uuid.UUID, registryType string, providerName string, accountNumber string, accountName string, qrCodeImage string, mailingAddress string) (*models.GiftRegistry, error)
	Update(id uint, projectID uuid.UUID, registryType string, providerName string, accountNumber string, accountName string, qrCodeImage string, mailingAddress string) (*models.GiftRegistry, error)
	Delete(id uint, projectID uuid.UUID) error
}

type giftRegistryService struct {
	repo repository.GiftRegistryRepository
}

func NewGiftRegistryService(repo repository.GiftRegistryRepository) GiftRegistryService {
	return &giftRegistryService{repo: repo}
}

func (s *giftRegistryService) GetByProjectID(projectID uuid.UUID) ([]models.GiftRegistry, error) {
	return s.repo.GetByProjectID(projectID)
}

func validateGiftRegistryFields(t models.GiftRegistryType, provider, accNum, accName, qrCodeImage, address string) error {
	switch t {
	case models.GiftRegistryTypeBank:
		if accNum == "" || accName == "" {
			return errors.New("bank registry requires account_number and account_name")
		}
	case models.GiftRegistryTypeEWallet:
		if qrCodeImage == "" || provider == "" {
			return errors.New("ewallet registry requires a QRIS image and provider_name")
		}
	case models.GiftRegistryTypePhysical:
		if address == "" {
			return errors.New("physical registry requires mailing_address")
		}
	default:
		return errors.New("invalid gift registry type")
	}
	return nil
}

func (s *giftRegistryService) Create(projectID uuid.UUID, registryType string, providerName string, accountNumber string, accountName string, qrCodeImage string, mailingAddress string) (*models.GiftRegistry, error) {
	t := models.GiftRegistryType(registryType)
	if err := validateGiftRegistryFields(t, providerName, accountNumber, accountName, qrCodeImage, mailingAddress); err != nil {
		return nil, err
	}

	registry := &models.GiftRegistry{
		ProjectID:      projectID,
		Type:           t,
		ProviderName:   providerName,
		AccountNumber:  accountNumber,
		AccountName:    accountName,
		QRCodeImage:    qrCodeImage,
		MailingAddress: mailingAddress,
	}

	if err := s.repo.Create(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

func (s *giftRegistryService) Update(id uint, projectID uuid.UUID, registryType string, providerName string, accountNumber string, accountName string, qrCodeImage string, mailingAddress string) (*models.GiftRegistry, error) {
	registry, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if registry == nil || registry.ProjectID != projectID {
		return nil, errors.New("gift registry not found or does not belong to project")
	}

	t := models.GiftRegistryType(registryType)
	if err := validateGiftRegistryFields(t, providerName, accountNumber, accountName, qrCodeImage, mailingAddress); err != nil {
		return nil, err
	}

	registry.Type = t
	registry.ProviderName = providerName
	registry.AccountNumber = accountNumber
	registry.AccountName = accountName
	registry.QRCodeImage = qrCodeImage
	registry.MailingAddress = mailingAddress

	if err := s.repo.Update(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

func (s *giftRegistryService) Delete(id uint, projectID uuid.UUID) error {
	registry, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if registry == nil || registry.ProjectID != projectID {
		return errors.New("gift registry not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
