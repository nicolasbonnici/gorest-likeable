package likeable

import (
	"time"
)

type Like struct {
	Id         string     `json:"id,omitempty" db:"id"`
	LikerId    *string    `json:"likerId,omitempty" db:"liker_id"`
	LikedId    *string    `json:"likedId,omitempty" db:"liked_id"`
	LikeableId string     `json:"likeableId" db:"likeable_id"`
	Likeable   string     `json:"likeable" db:"likeable"`
	IpAddress  *string    `json:"ipAddress,omitempty" db:"ip_address"`
	UserAgent  *string    `json:"userAgent,omitempty" db:"user_agent"`
	LikedAt    time.Time  `json:"likedAt" db:"liked_at"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty" db:"updated_at"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" db:"created_at"`
}

func (Like) TableName() string {
	return "likes"
}
