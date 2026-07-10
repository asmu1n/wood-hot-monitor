package hotspot

import (
	"time"
	"wood-hot-monitor/ent"
	"wood-hot-monitor/internal/core/models"
)

func mapHotspot(h *ent.Hotspot) models.Hotspot {
	m := models.Hotspot{
		ID:               h.ID,
		Title:            h.Title,
		Content:          h.Content,
		URL:              h.URL,
		Source:           h.Source,
		SourceID:         h.SourceID,
		IsReal:           h.IsReal,
		Relevance:        h.Relevance,
		RelevanceReason:  h.RelevanceReason,
		KeywordMentioned: h.KeywordMentioned,
		Importance:       h.Importance,
		Summary:          h.Summary,
		ViewCount:        h.ViewCount,
		LikeCount:        h.LikeCount,
		RetweetCount:     h.RetweetCount,
		ReplyCount:       h.ReplyCount,
		CommentCount:     h.CommentCount,
		QuoteCount:       h.QuoteCount,
		DanmakuCount:     h.DanmakuCount,
		AuthorName:       h.AuthorName,
		AuthorUsername:   h.AuthorUsername,
		AuthorAvatar:     h.AuthorAvatar,
		AuthorFollowers:  h.AuthorFollowers,
		AuthorVerified:   h.AuthorVerified,
		KeywordID:        h.KeywordID,
		IsNotified:       h.IsNotified,
		IsRead:           h.IsRead,
		CreatedAt:        h.CreatedAt.Format(time.DateTime),
	}
	if h.NotifiedAt != nil {
		s := h.NotifiedAt.Format(time.DateTime)
		m.NotifiedAt = &s
	}
	if h.PublishedAt != nil {
		s := h.PublishedAt.Format(time.DateTime)
		m.PublishedAt = &s
	}
	if h.Edges.Keyword != nil {
		kw := h.Edges.Keyword
		m.Keyword = &models.HotspotKeyword{ID: kw.ID, Text: kw.Text, Category: kw.Category}
	}
	return m
}
