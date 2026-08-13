package repository

import (
	"database/sql"
	"time"
)

type AuditSnapshotRepository struct {
	db *sql.DB
}

func NewAuditSnapshotRepository(db *sql.DB) *AuditSnapshotRepository {
	return &AuditSnapshotRepository{db: db}
}

type AuditSnapshot struct {
	ID            string     `json:"id" db:"id"`
	ChannelID     string     `json:"channel_id" db:"channel_id"`
	SnapshotDate  time.Time  `json:"snapshot_date" db:"snapshot_date"`
	WeekNumber    int        `json:"week_number" db:"week_number"`
	Year          int        `json:"year" db:"year"`
	Findings      []byte     `json:"findings" db:"findings"`
	Summary       []byte     `json:"summary" db:"summary"`
	TotalFindings int        `json:"total_findings" db:"total_findings"`
	CriticalCount int        `json:"critical_count" db:"critical_count"`
	HighCount     int        `json:"high_count" db:"high_count"`
	MediumCount   int        `json:"medium_count" db:"medium_count"`
	LowCount      int        `json:"low_count" db:"low_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

func (r *AuditSnapshotRepository) Create(snapshot *AuditSnapshot) error {
	query := `
		INSERT INTO analytics.audit_snapshots (
			channel_id, snapshot_date, week_number, year, findings, summary,
			total_findings, critical_count, high_count, medium_count, low_count
		) VALUES (
			\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11
		) ON CONFLICT (channel_id, year, week_number) DO NOTHING
	`
	_, err := r.db.Exec(
		query,
		snapshot.ChannelID,
		snapshot.SnapshotDate,
		snapshot.WeekNumber,
		snapshot.Year,
		snapshot.Findings,
		snapshot.Summary,
		snapshot.TotalFindings,
		snapshot.CriticalCount,
		snapshot.HighCount,
		snapshot.MediumCount,
		snapshot.LowCount,
	)
	return err
}

func (r *AuditSnapshotRepository) GetLastSnapshot(channelID string) (*AuditSnapshot, error) {
	query := `
		SELECT id, channel_id, snapshot_date, week_number, year, findings, summary,
		       total_findings, critical_count, high_count, medium_count, low_count, created_at
		FROM analytics.audit_snapshots
		WHERE channel_id = \$1
		ORDER BY snapshot_date DESC
		LIMIT 1
	`

	var snapshot AuditSnapshot
	err := r.db.QueryRow(query, channelID).Scan(
		&snapshot.ID,
		&snapshot.ChannelID,
		&snapshot.SnapshotDate,
		&snapshot.WeekNumber,
		&snapshot.Year,
		&snapshot.Findings,
		&snapshot.Summary,
		&snapshot.TotalFindings,
		&snapshot.CriticalCount,
		&snapshot.HighCount,
		&snapshot.MediumCount,
		&snapshot.LowCount,
		&snapshot.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (r *AuditSnapshotRepository) GetLastTwoSnapshots(channelID string) ([]*AuditSnapshot, error) {
	query := `
		SELECT id, channel_id, snapshot_date, week_number, year, findings, summary,
		       total_findings, critical_count, high_count, medium_count, low_count, created_at
		FROM analytics.audit_snapshots
		WHERE channel_id = \$1
		ORDER BY snapshot_date DESC
		LIMIT 2
	`

	rows, err := r.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*AuditSnapshot
	for rows.Next() {
		var snapshot AuditSnapshot
		err := rows.Scan(
			&snapshot.ID,
			&snapshot.ChannelID,
			&snapshot.SnapshotDate,
			&snapshot.WeekNumber,
			&snapshot.Year,
			&snapshot.Findings,
			&snapshot.Summary,
			&snapshot.TotalFindings,
			&snapshot.CriticalCount,
			&snapshot.HighCount,
			&snapshot.MediumCount,
			&snapshot.LowCount,
			&snapshot.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, nil
}
