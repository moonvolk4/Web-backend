package repository

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"column:login;unique;not null" json:"login"`
	Password    string `gorm:"column:password;not null" json:"-"`
	IsModerator bool   `gorm:"column:is_moderator;not null;default:false" json:"is_moderator"`
}

func (User) TableName() string { return "users" }

func (r *Repository) CreateUser(login, password string, isModerator bool) (*User, error) {
	if login == "" || password == "" {
		return nil, errors.New("login and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{Login: login, Password: string(hash), IsModerator: isModerator}
	if err := r.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetUserByLogin(login string) (*User, error) {
	var u User
	if err := r.db.Where("login = ?", login).Limit(1).Find(&u).Error; err != nil {
		return nil, err
	}
	if u.ID == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *Repository) GetUserByID(id uint) (*User, error) {
	var u User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) VerifyUser(login, password string) (*User, error) {
	u, err := r.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

func (r *Repository) UpdateUserLogin(id uint, newLogin string) error {
	if strings.TrimSpace(newLogin) == "" {
		return errors.New("login empty")
	}
	// ensure unique
	var cnt int64
	if err := r.db.Model(&User{}).Where("login = ? AND id <> ?", newLogin, id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("login already taken")
	}
	return r.db.Model(&User{}).Where("id = ?", id).Update("login", newLogin).Error
}

// duplicate removed; stricter implementation above ensures unique constraint
