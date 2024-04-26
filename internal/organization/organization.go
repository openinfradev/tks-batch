package organization

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Organization represents a kubernetes organization information.
type Organization struct {
	ID               string `gorm:"primarykey"`
	WorkflowId       string
	Status           domain.OrganizationStatus
	StatusDesc       string
	PrimaryClusterId string
}

// Accessor accesses organization info in DB.
type OrganizationAccessor struct {
	db *gorm.DB
}

// NewOrganizationAccessor returns new Accessor to access organizations.
func New(db *gorm.DB) *OrganizationAccessor {
	return &OrganizationAccessor{
		db: db,
	}
}

// For Unittest
func (x *OrganizationAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *OrganizationAccessor) GetIncompleteOrganizations() ([]Organization, error) {
	var organizations []Organization

	res := x.db.
		Where("status IN ?", []domain.OrganizationStatus{domain.OrganizationStatus_CREATING, domain.OrganizationStatus_DELETING}).
		Find(&organizations)

	if res.Error != nil {
		return nil, res.Error
	}

	return organizations, nil
}

func (x *OrganizationAccessor) Get(id string) (organization Organization, err error) {
	res := x.db.Where("id = ?", id).First(&organization)
	if res.Error != nil {
		return organization, res.Error
	}

	return
}

func (x *OrganizationAccessor) UpdateOrganizationStatus(organizationId string, status domain.OrganizationStatus, statusDesc string, workflowId string) error {
	log.Info(context.TODO(), fmt.Sprintf("UpdateOrganizationStatus. organizationId[%s], status[%d], statusDesc[%s], workflowId[%s]", organizationId, status, statusDesc, workflowId))
	res := x.db.Model(Organization{}).
		Where("ID = ?", organizationId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in organization with id %s", organizationId)
	}
	return nil
}
