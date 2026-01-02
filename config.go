package likeable

import (
	"errors"
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
)

type Config struct {
	Database           database.Database
	AllowedTypes       []string `json:"allowed_types" yaml:"allowed_types"`
	PaginationLimit    int      `json:"pagination_limit" yaml:"pagination_limit"`
	MaxPaginationLimit int      `json:"max_pagination_limit" yaml:"max_pagination_limit"`
	EnableUserLikes    bool     `json:"enable_user_likes" yaml:"enable_user_likes"`
}

func DefaultConfig() Config {
	return Config{
		AllowedTypes:       []string{"post"},
		PaginationLimit:    50,
		MaxPaginationLimit: 200,
		EnableUserLikes:    false,
	}
}

func (c *Config) Validate() error {
	if len(c.AllowedTypes) == 0 {
		return errors.New("allowed_types cannot be empty")
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, likeableType := range c.AllowedTypes {
		if likeableType == "" {
			return errors.New("allowed_types cannot contain empty strings")
		}
		if seen[likeableType] {
			return fmt.Errorf("duplicate type in allowed_types: %s", likeableType)
		}
		seen[likeableType] = true
	}

	if c.PaginationLimit < 1 || c.PaginationLimit > c.MaxPaginationLimit {
		return errors.New("pagination_limit must be between 1 and max_pagination_limit")
	}

	return nil
}

func (c *Config) IsAllowedType(likeableType string) bool {
	for _, allowed := range c.AllowedTypes {
		if allowed == likeableType {
			return true
		}
	}
	return false
}
