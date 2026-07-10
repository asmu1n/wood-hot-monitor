package ent

import (
	"wood-hot-monitor/ent"
	domain "wood-hot-monitor/internal/domain/hotspot"
)

func mapToDomain(row *ent.Hotspot) domain.Hotspot {
	h := domain.Hotspot{
		ID:               row.ID,
		Title:            row.Title,
		Content:          row.Content,
		URL:              row.URL,
		Source:           row.Source,
		SourceID:         row.SourceID,
		IsReal:           row.IsReal,
		Relevance:        row.Relevance,
		RelevanceReason:  row.RelevanceReason,
		KeywordMentioned: row.KeywordMentioned,
		Importance:       domain.Importance(row.Importance),
		Summary:          row.Summary,
		ViewCount:        row.ViewCount,
		LikeCount:        row.LikeCount,
		RetweetCount:     row.RetweetCount,
		ReplyCount:       row.ReplyCount,
		CommentCount:     row.CommentCount,
		QuoteCount:       row.QuoteCount,
		DanmakuCount:     row.DanmakuCount,
		KeywordID:        row.KeywordID,
		IsNotified:       row.IsNotified,
		IsRead:           row.IsRead,
		CreatedAt:        row.CreatedAt,
		PublishedAt:      row.PublishedAt,
		NotifiedAt:       row.NotifiedAt,
	}

	if row.AuthorName != nil || row.AuthorUsername != nil {
		h.Author = &domain.Author{
			Name:     deref(row.AuthorName),
			Username: deref(row.AuthorUsername),
			Avatar:   deref(row.AuthorAvatar),
		}
		if row.AuthorFollowers != nil {
			h.Author.Followers = *row.AuthorFollowers
		}
		if row.AuthorVerified != nil {
			h.Author.Verified = *row.AuthorVerified
		}
	}

	if row.Edges.Keyword != nil {
		kw := row.Edges.Keyword
		h.Keyword = &domain.HotspotKeyword{
			ID:       kw.ID,
			Text:     kw.Text,
			Category: kw.Category,
		}
	}
	return h
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
