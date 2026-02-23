package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/safebites/backend-go/internal/model"
	"github.com/safebites/backend-go/internal/repository"
)

type scanService struct {
	scans repository.ScanRepository
}

func NewScanService(scans repository.ScanRepository) ScanService {
	return &scanService{scans: scans}
}

func (s *scanService) ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.scans.ListByUser(ctx, userID, limit)
}

func (s *scanService) Create(ctx context.Context, scan *model.Scan) (*model.Scan, error) {
	if scan == nil {
		return nil, fmt.Errorf("scan is required")
	}
	if strings.TrimSpace(scan.UserID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(scan.ProductName) == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if scan.SafetyScore < 0 || scan.SafetyScore > 100 {
		return nil, fmt.Errorf("safety score must be between 0 and 100")
	}

	return s.scans.Create(ctx, scan)
}

func (s *scanService) GetStats(ctx context.Context, userID string) (*model.UserStats, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.scans.GetStats(ctx, userID)
}
