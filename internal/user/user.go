package user

import (
	"time"

	"gorm.io/gorm"

	"github.com/gofrs/uuid"
)

type User struct {
	ID                uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId         string    `json:"accountId"`
	Password          string    `gorm:"-:all" json:"password"`
	Name              string    `json:"name"`
	Token             string    `json:"token"`
	RoleId            string
	OrganizationId    string
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	PasswordUpdatedAt time.Time `json:"passwordUpdatedAt"`
	PasswordExpired   bool      `json:"passwordExpired"`
	Email             string    `json:"email"`
	Department        string    `json:"department"`
	Description       string    `json:"description"`
}

type UserAccessor struct {
	db *gorm.DB
}

func New(db *gorm.DB) *UserAccessor {
	return &UserAccessor{
		db: db,
	}
}

func (x *UserAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *UserAccessor) GetOrganizationAdmin(organizationId string) (out User, err error) {
	res := x.db.
		Where("organization_id = ?", organizationId).
		First(&out)

	if res.Error != nil {
		return out, res.Error
	}
	return
}
