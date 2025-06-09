package store

import "S3_project/auth/internal/app/model"

type UserRepository interface {
	Create(user *model.User) error
	Find(id int) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
}
