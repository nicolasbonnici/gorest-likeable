# GoREST Likeable Plugin

[![Test](https://github.com/nicolasbonnici/gorest-likeable/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/nicolasbonnici/gorest-likeable/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-likeable)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-likeable)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A polymorphic like/reaction plugin for GoREST that allows adding likes to any resource type.

## Features

- **Polymorphic Likes**: Add likes to any resource type (posts, articles, comments, etc.)
- **User Likes**: Optional support for liking user profiles
- **Duplicate Prevention**: Unique constraint prevents duplicate likes
- **Configurable Allowed Types**: Control which resource types can be liked
- **User Association**: Automatic user authentication integration
- **Pagination**: Built-in pagination support for like lists
- **Go Migrations**: Database schema managed via Go code (not SQL files)

## Installation

```bash
go get github.com/nicolasbonnici/gorest-likeable
```

## Configuration

```yaml
plugins:
  - name: likeable
    enabled: true
    config:
      allowed_types: ["post", "comment", "article"]
      pagination_limit: 50
      max_pagination_limit: 200
      enable_user_likes: false
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `allowed_types` | `[]string` | `["post"]` | Resource types that can be liked |
| `pagination_limit` | `int` | `50` | Default pagination limit |
| `max_pagination_limit` | `int` | `200` | Maximum allowed pagination limit |
| `enable_user_likes` | `bool` | `false` | Allow liking user profiles |

## API Endpoints

### List Likes
```
GET /likes?likeable=post&likeableId={id}
```

### Get Like
```
GET /likes/:id
```

### Create Like (Toggle)
```
POST /likes
Content-Type: application/json

{
  "likeableId": "uuid",
  "likeable": "post",
  "likedId": "uuid"  // optional, for user likes
}
```

**Note**: If the same user tries to like the same resource twice, it returns a 409 Conflict error.

### Update Like (Refresh Timestamp)
```
PUT /likes/:id
```

### Delete Like (Unlike)
```
DELETE /likes/:id
```

## Database Schema

```sql
CREATE TABLE likes (
    id UUID PRIMARY KEY,
    liker_id UUID REFERENCES users(id) ON DELETE SET NULL,
    liked_id UUID REFERENCES users(id) ON DELETE SET NULL,
    likeable_id UUID NOT NULL,
    likeable TEXT NOT NULL,
    liked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (liker_id, likeable, likeable_id)
);

-- Indexes
CREATE INDEX idx_likeable ON likes(likeable, likeable_id, liked_at);
CREATE INDEX idx_liker_id ON likes(liker_id);
```

## Usage Example

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/nicolasbonnici/gorest"
    "github.com/nicolasbonnici/gorest-likeable"
)

func main() {
    app := fiber.New()

    // Initialize plugin with configuration
    plugin := likeable.NewPlugin()

    config := map[string]interface{}{
        "database": db,
        "allowed_types": []interface{}{"post", "comment"},
        "pagination_limit": 50,
        "enable_user_likes": true,
    }

    if err := plugin.Initialize(config); err != nil {
        panic(err)
    }

    plugin.SetupEndpoints(app)

    app.Listen(":3000")
}
```

## Development

### Run Tests
```bash
make test
```

### Run Linter
```bash
make lint
```

### Build
```bash
make build
```

### Coverage Report
```bash
make coverage
```

## Security Features

- **Unique Constraint**: Prevents duplicate likes from the same user on the same resource
- **Type Validation**: Only configured resource types are allowed
- **Foreign Key Constraints**: Maintains referential integrity where possible
- **Conflict Detection**: Returns 409 status for duplicate like attempts

## Use Cases

### Basic Post Like
```bash
# Like a post
POST /likes
{
  "likeableId": "123e4567-e89b-12d3-a456-426614174000",
  "likeable": "post"
}

# Unlike a post
DELETE /likes/{like_id}

# Get all likes for a post
GET /likes?likeable=post&likeableId=123e4567-e89b-12d3-a456-426614174000
```

### Comment Like
```bash
# Like a comment
POST /likes
{
  "likeableId": "987e6543-e21b-32d1-a654-426614174001",
  "likeable": "comment"
}
```

### User Profile Like (if enabled)
```bash
# Like a user profile
POST /likes
{
  "likeableId": "user_profile_id",
  "likeable": "user",
  "likedId": "user_id_being_liked"
}
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- Code is linted
- New features have test coverage
- Documentation is updated

## Part of GoREST Ecosystem

- [GoREST](https://github.com/nicolasbonnici/gorest) - Core framework
- [GoREST Auth](https://github.com/nicolasbonnici/gorest-auth) - Authentication plugin
- [GoREST Commentable](https://github.com/nicolasbonnici/gorest-commentable) - Comment plugin
- [GoREST Blog](https://github.com/nicolasbonnici/gorest-blog) - Blog plugin
