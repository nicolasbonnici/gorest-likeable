package likeable

import (
	"time"

	"github.com/google/uuid"
)

type LikeConverter struct{}

func (c *LikeConverter) CreateDTOToModel(dto LikeCreateDTO) Like {
	return Like{
		Id:         uuid.New().String(),
		LikeableId: dto.LikeableId,
		Likeable:   dto.Likeable,
		LikedId:    dto.LikedId,
		LikedAt:    time.Now(),
	}
}

func (c *LikeConverter) UpdateDTOToModel(dto LikeUpdateDTO) Like {
	return Like{}
}

func (c *LikeConverter) ModelToResponseDTO(model Like) LikeResponseDTO {
	return LikeResponseDTO{
		ID:         model.Id,
		LikerID:    model.LikerId,
		LikedID:    model.LikedId,
		LikeableID: model.LikeableId,
		Likeable:   model.Likeable,
		IPAddress:  model.IpAddress,
		UserAgent:  model.UserAgent,
		LikedAt:    model.LikedAt,
		UpdatedAt:  model.UpdatedAt,
		CreatedAt:  model.CreatedAt,
	}
}

func (c *LikeConverter) ModelsToResponseDTOs(models []Like) []LikeResponseDTO {
	dtos := make([]LikeResponseDTO, len(models))
	for i, model := range models {
		dtos[i] = c.ModelToResponseDTO(model)
	}
	return dtos
}
