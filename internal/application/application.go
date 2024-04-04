package application

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type AppGroup struct {
	ID         string `gorm:"primarykey"`
	WorkflowId string
	Status     domain.AppGroupStatus
	StatusDesc string
}

type ApplicationAccessor struct {
	db *gorm.DB
}

// New returns new accessor's ptr.
func New(db *gorm.DB) *ApplicationAccessor {
	return &ApplicationAccessor{
		db: db,
	}
}

// For Unittest
func (x *ApplicationAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *ApplicationAccessor) GetIncompleteAppGroups() ([]AppGroup, error) {
	var appGroups []AppGroup

	res := x.db.
		Where("status IN ?", []domain.AppGroupStatus{domain.AppGroupStatus_INSTALLING, domain.AppGroupStatus_DELETING}).
		Find(&appGroups)

	if res.Error != nil {
		return nil, res.Error
	}

	return appGroups, nil
}

func (x *ApplicationAccessor) UpdateAppGroupStatus(appGroupId string, status domain.AppGroupStatus, statusDesc string, workflowId string) error {
	log.Info(context.TODO(), fmt.Sprintf("UpdateAppGroupStatus. appGroupId[%s], status[%d], statusDesc[%s], workflowId[%s]", appGroupId, status, statusDesc, workflowId))
	res := x.db.Model(AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}
	return nil
}
