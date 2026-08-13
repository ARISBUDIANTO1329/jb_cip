package service

import (
	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
)

// Rule definition
type Rule struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
}

// Finding is the result of evaluating a rule against audit data
type Finding struct {
	RuleID         string  `json:"rule_id"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	Severity       string  `json:"severity"`
	Confidence     float64 `json:"confidence"`
	Triggered      bool    `json:"triggered"`
	Breakdown      string  `json:"breakdown,omitempty"`
	Recommendation string  `json:"recommendation,omitempty"`
}

type ruleFunc func(data interface{}) *Finding

type ruleEntry struct {
	rule Rule
	fn   ruleFunc
}

type AuditService struct {
	auditRepo    *repository.AuditRepository
	channelRepo  *repository.ChannelRepository
	videoRules   []ruleEntry
	channelRules []ruleEntry
}

func NewAuditService(auditRepo *repository.AuditRepository, channelRepo *repository.ChannelRepository) *AuditService {
	s := &AuditService{
		auditRepo:   auditRepo,
		channelRepo: channelRepo,
	}
	s.registerVideoRules()
	s.registerChannelRules()
	return s
}

func (s *AuditService) registerVideoRules() {
	s.videoRules = []ruleEntry{
		{rule: Rule{"DQ-001", "data_quality", "Missing title", "low", 0.9}, fn: videoMissingTitle},
		{rule: Rule{"DQ-002", "data_quality", "Missing description", "medium", 0.85}, fn: videoMissingDescription},
		{rule: Rule{"DQ-003", "data_quality", "Missing thumbnail", "medium", 0.9}, fn: videoMissingThumbnail},
		{rule: Rule{"DQ-004", "data_quality", "Zero duration", "medium", 0.8}, fn: videoZeroDuration},
		{rule: Rule{"CTR-001", "ctr", "CTR below 3 percent", "medium", 0.75}, fn: videoLowCTR},
		{rule: Rule{"CTR-002", "ctr", "Zero impressions", "high", 0.9}, fn: videoZeroImpressions},
		{rule: Rule{"ENG-001", "engagement", "Like ratio below 2 percent", "low", 0.7}, fn: videoLowLikeRatio},
		{rule: Rule{"ENG-002", "engagement", "Zero comments despite views", "low", 0.6}, fn: videoZeroComments},
		{rule: Rule{"RET-001", "retention", "AVD below 30 percent of duration", "medium", 0.7}, fn: videoLowAVD},
		{rule: Rule{"RET-002", "retention", "Retention below 10 percent", "high", 0.8}, fn: videoLowRetention},
	}
}

func (s *AuditService) registerChannelRules() {
	s.channelRules = []ruleEntry{
		{rule: Rule{"PUB-001", "publishing", "No upload in 14 plus days", "high", 0.9}, fn: channelUploadGap},
		{rule: Rule{"PUB-002", "publishing", "Low upload frequency (under 5 videos)", "low", 0.7}, fn: channelLowVideoCount},
		{rule: Rule{"ENG-010", "engagement", "Low average views per video", "medium", 0.7}, fn: channelLowAvgViews},
	}
}

func (s *AuditService) resolveChannel(channelID, workspaceID string) (string, error) {
	if _, err := uuid.Parse(channelID); err != nil {
		return "", errors.New("AUDIT_002", "Invalid channel ID", 400)
	}

	channel, err := s.channelRepo.GetByID(uuid.MustParse(channelID))
	if err != nil {
		return "", errors.New("AUDIT_003", "Channel not found", 404)
	}

	if channel.WorkspaceID != workspaceID {
		return "", errors.New("AUDIT_004", "Channel does not belong to this workspace", 403)
	}

	return channel.ID, nil
}

func (s *AuditService) AuditVideo(videoID, workspaceID string) ([]Finding, error) {
	data, err := s.auditRepo.GetVideoAuditData(videoID, workspaceID)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, entry := range s.videoRules {
		f := entry.fn(data)
		if f != nil {
			f.RuleID = entry.rule.ID
			f.Category = entry.rule.Category
			f.Description = entry.rule.Description
			f.Severity = entry.rule.Severity
			if f.Confidence == 0 {
				f.Confidence = entry.rule.Confidence
			}
			findings = append(findings, *f)
		}
	}

	return findings, nil
}

func (s *AuditService) AuditChannel(channelID, workspaceID string, videoLimit int) ([]Finding, error) {
	chID, err := s.resolveChannel(channelID, workspaceID)
	if err != nil {
		return nil, err
	}

	var findings []Finding

	channelData, err := s.auditRepo.GetChannelAuditData(chID, workspaceID)
	if err != nil {
		return nil, err
	}

	for _, entry := range s.channelRules {
		f := entry.fn(channelData)
		if f != nil {
			f.RuleID = entry.rule.ID
			f.Category = entry.rule.Category
			f.Description = entry.rule.Description
			f.Severity = entry.rule.Severity
			if f.Confidence == 0 {
				f.Confidence = entry.rule.Confidence
			}
			findings = append(findings, *f)
		}
	}

	if videoLimit <= 0 {
		videoLimit = 50
	}
	videos, err := s.auditRepo.GetChannelVideosAuditData(chID, workspaceID, videoLimit)
	if err != nil {
		return nil, err
	}

	for _, video := range videos {
		for _, entry := range s.videoRules {
			f := entry.fn(&video)
			if f != nil {
				f.RuleID = entry.rule.ID
				f.Category = entry.rule.Category
				f.Description = entry.rule.Description
				f.Severity = entry.rule.Severity
				if f.Confidence == 0 {
					f.Confidence = entry.rule.Confidence
				}
				f.Breakdown = "video: " + video.Title
				findings = append(findings, *f)
			}
		}
	}

	return findings, nil
}

// ----- Rule implementations (video level) -----

func videoMissingTitle(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.Title == "" {
		return &Finding{Triggered: true, Recommendation: "Add a descriptive title for the video", Confidence: 0.9}
	}
	return nil
}

func videoMissingDescription(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.Description == "" {
		return &Finding{Triggered: true, Recommendation: "Add description (SEO benefit + viewer context)", Confidence: 0.85}
	}
	return nil
}

func videoMissingThumbnail(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.ThumbnailURL == "" {
		return &Finding{Triggered: true, Recommendation: "Upload a custom thumbnail", Confidence: 0.9}
	}
	return nil
}

func videoZeroDuration(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.Duration <= 0 {
		return &Finding{Triggered: true, Recommendation: "Video duration not synced. Re-sync video data", Confidence: 0.8}
	}
	return nil
}

func videoLowCTR(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.TotalImpressions > 0 && d.AvgCTR < 3.0 {
		r := "CTR is low. Review thumbnail, title, and metadata"
		if d.AvgCTR < 1.0 {
			r = "CTR very low. Consider changing thumbnail and title"
		}
		return &Finding{Triggered: true, Recommendation: r, Confidence: 0.75}
	}
	return nil
}

func videoZeroImpressions(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.TotalImpressions <= 0 {
		return &Finding{Triggered: true, Recommendation: "No impressions. Video may not be indexed or has very limited reach", Confidence: 0.9}
	}
	return nil
}

func videoLowLikeRatio(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.ViewCount > 0 {
		ratio := float64(d.LikeCount) / float64(d.ViewCount) * 100
		if ratio < 2.0 {
			r := "Low like ratio. Add CTA for likes"
			if d.LikeCount == 0 {
				r = "No likes. Consider adding subscribe/like CTA in video"
			}
			return &Finding{Triggered: true, Recommendation: r, Confidence: 0.7}
		}
	}
	return nil
}

func videoZeroComments(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.ViewCount > 10 && d.CommentCount == 0 {
		return &Finding{Triggered: true, Recommendation: "No comments despite views. Ask a question to drive discussion", Confidence: 0.6}
	}
	return nil
}

func videoLowAVD(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.Duration > 0 && d.AvgAVD > 0 {
		pct := d.AvgAVD / float64(d.Duration) * 100
		if pct < 30.0 {
			return &Finding{Triggered: true, Recommendation: "Most viewers leave early. Review hook and content pacing", Confidence: 0.7}
		}
	}
	return nil
}

func videoLowRetention(data interface{}) *Finding {
	d := data.(*repository.VideoAuditData)
	if d.AvgRetention > 0 && d.AvgRetention < 10.0 {
		return &Finding{Triggered: true, Recommendation: "Very low retention. Content not matching viewer expectations", Confidence: 0.8}
	}
	return nil
}

// ----- Rule implementations (channel level) -----

func channelUploadGap(data interface{}) *Finding {
	d := data.(*repository.ChannelAuditData)
	if d.DaysSinceLastUpload > 14 {
		return &Finding{Triggered: true, Recommendation: "Upload gap over 14 days. Publish new content to maintain algorithm distribution", Confidence: 0.9}
	}
	return nil
}

func channelLowVideoCount(data interface{}) *Finding {
	d := data.(*repository.ChannelAuditData)
	if d.VideoCount < 5 {
		return &Finding{Triggered: true, Recommendation: "Only a few videos uploaded. Increase content volume for algorithm discovery", Confidence: 0.7}
	}
	return nil
}

func channelLowAvgViews(data interface{}) *Finding {
	d := data.(*repository.ChannelAuditData)
	if d.AvgViewsPerVideo < 50 && d.VideoCount > 0 {
		return &Finding{Triggered: true, Recommendation: "Average views below 50 per video. Focus on audience growth", Confidence: 0.7}
	}
	return nil
}
