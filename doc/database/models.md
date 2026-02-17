# Models

Bourbon uses [GORM](https://gorm.io) for ORM, allowing you to define your data models as Go structs.

## Defining Models

Models are defined in `apps/<app_name>/models.go`.

```go
package posts

import (
    "time"
    "github.com/ishubhamsingh2e/bourbon/bourbon/models"
)

type Post struct {
    models.BaseModel // Includes ID, CreatedAt, UpdatedAt, DeletedAt
    Title     string `gorm:"size:255;not null" json:"title"`
    Content   string `gorm:"type:text" json:"content"`
    AuthorID  uint   `json:"author_id"`
    Published bool   `gorm:"default:false" json:"published"`
}
```

### Base Model

`models.BaseModel` provides the following fields:

- `ID`: Auto-incrementing primary key.
- `CreatedAt`: Timestamp of creation.
- `UpdatedAt`: Timestamp of last update.
- `DeletedAt`: Soft delete timestamp (optional).

## Relationships

Define relationships using standard GORM tags.

### One-to-One

```go
type User struct {
    models.BaseModel
    Profile Profile
}

type Profile struct {
    models.BaseModel
    UserID uint
    Bio    string
}
```

### One-to-Many

```go
type User struct {
    models.BaseModel
    Posts []Post
}

type Post struct {
    models.BaseModel
    UserID uint
}
```

### Many-to-Many

```go
type User struct {
    models.BaseModel
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    models.BaseModel
    Name string
}
```

## Advanced Configuration

You can customize table names, hooks, and other GORM features directly on your structs.

```go
func (Post) TableName() string {
    return "blog_posts"
}
```
