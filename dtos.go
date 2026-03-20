package likeable

import (
	"time"
)

type LikeCreateDTO struct {
	LikeableId string  `json:"likeableId"`
	Likeable   string  `json:"likeable"`
	LikedId    *string `json:"likedId,omitempty"`
}

type LikeUpdateDTO struct {
}

type LikeResponseDTO struct {
	ID         string     `json:"id"`
	LikerID    *string    `json:"likerId,omitempty"`
	LikedID    *string    `json:"likedId,omitempty"`
	LikeableID string     `json:"likeableId"`
	Likeable   string     `json:"likeable"`
	IPAddress  *string    `json:"ipAddress,omitempty"`
	UserAgent  *string    `json:"userAgent,omitempty"`
	LikedAt    time.Time  `json:"likedAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
}
