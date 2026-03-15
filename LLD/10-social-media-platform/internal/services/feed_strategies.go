package services

import (
	"sort"

	"social-media-platform/internal/interfaces"
)

// ChronologicalFeedStrategy sorts feed by timestamp (newest first).
// Strategy pattern: interchangeable sorting algorithm
type ChronologicalFeedStrategy struct{}

func (s *ChronologicalFeedStrategy) Sort(items []*interfaces.FeedItem) []*interfaces.FeedItem {
	result := make([]*interfaces.FeedItem, len(items))
	copy(result, items)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Post.CreatedAt.After(result[j].Post.CreatedAt)
	})
	return result
}

// PopularityFeedStrategy sorts feed by engagement score (likes + comments).
// Strategy pattern: alternative sorting algorithm
type PopularityFeedStrategy struct{}

func (s *PopularityFeedStrategy) Sort(items []*interfaces.FeedItem) []*interfaces.FeedItem {
	result := make([]*interfaces.FeedItem, len(items))
	copy(result, items)
	sort.Slice(result, func(i, j int) bool {
		scoreI := result[i].LikeCount + result[i].CommentCount
		scoreJ := result[j].LikeCount + result[j].CommentCount
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		return result[i].Post.CreatedAt.After(result[j].Post.CreatedAt)
	})
	return result
}
