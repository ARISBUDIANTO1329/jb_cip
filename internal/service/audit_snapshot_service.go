package service

import (
	"encoding/json"
	"time"

	"github.com/jaybani/jb_cip/internal/repository"
)

type AuditSnapshotService struct {
	snapshotRepo *repository.AuditSnapshotRepository
	auditService *AuditService
}

func NewAuditSnapshotService(snapshotRepo *repository.AuditSnapshotRepository, auditService *AuditService) *AuditSnapshotService {
	return &AuditSnapshotService{
		snapshotRepo: snapshotRepo,
		auditService: auditService,
	}
}

func (s *AuditSnapshotService) CreateSnapshot(channelID, workspaceID string) error {
	findings, err := s.auditService.AuditChannel(channelID, workspaceID, 50)
	if err != nil {
		return err
	}

	now := time.Now()
	year, week := now.ISOWeek()

	findingsJSON, _ := json.Marshal(findings)

	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, f := range findings {
		switch f.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		case "medium":
			mediumCount++
		case "low":
			lowCount++
		}
	}

	summary := map[string]interface{}{
		"total":    len(findings),
		"critical": criticalCount,
		"high":     highCount,
		"medium":   mediumCount,
		"low":      lowCount,
	}
	summaryJSON, _ := json.Marshal(summary)

	snapshot := &repository.AuditSnapshot{
		ChannelID:     channelID,
		SnapshotDate:  now,
		WeekNumber:    week,
		Year:          year,
		Findings:      findingsJSON,
		Summary:       summaryJSON,
		TotalFindings: len(findings),
		CriticalCount: criticalCount,
		HighCount:     highCount,
		MediumCount:   mediumCount,
		LowCount:      lowCount,
	}

	return s.snapshotRepo.Create(snapshot)
}

type ComparisonResult struct {
	Current  *SnapshotSummary `json:"current"`
	Previous *SnapshotSummary `json:"previous"`
	Delta    *DeltaSummary    `json:"delta"`
}

type SnapshotSummary struct {
	Date          time.Time `json:"date"`
	Week          int       `json:"week"`
	Year          int       `json:"year"`
	TotalFindings int       `json:"total_findings"`
	Critical      int       `json:"critical"`
	High          int       `json:"high"`
	Medium        int       `json:"medium"`
	Low           int       `json:"low"`
}

type DeltaSummary struct {
	TotalFindings int `json:"total_findings"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
}

func (s *AuditSnapshotService) CompareSnapshots(channelID string) (*ComparisonResult, error) {
	snapshots, err := s.snapshotRepo.GetLastTwoSnapshots(channelID)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return &ComparisonResult{}, nil
	}

	current := &SnapshotSummary{
		Date:          snapshots[0].SnapshotDate,
		Week:          snapshots[0].WeekNumber,
		Year:          snapshots[0].Year,
		TotalFindings: snapshots[0].TotalFindings,
		Critical:      snapshots[0].CriticalCount,
		High:          snapshots[0].HighCount,
		Medium:        snapshots[0].MediumCount,
		Low:           snapshots[0].LowCount,
	}

	result := &ComparisonResult{
		Current: current,
	}

	if len(snapshots) == 2 {
		previous := &SnapshotSummary{
			Date:          snapshots[1].SnapshotDate,
			Week:          snapshots[1].WeekNumber,
			Year:          snapshots[1].Year,
			TotalFindings: snapshots[1].TotalFindings,
			Critical:      snapshots[1].CriticalCount,
			High:          snapshots[1].HighCount,
			Medium:        snapshots[1].MediumCount,
			Low:           snapshots[1].LowCount,
		}

		delta := &DeltaSummary{
			TotalFindings: current.TotalFindings - previous.TotalFindings,
			Critical:      current.Critical - previous.Critical,
			High:          current.High - previous.High,
			Medium:        current.Medium - previous.Medium,
			Low:           current.Low - previous.Low,
		}

		result.Previous = previous
		result.Delta = delta
	}

	return result, nil
}
