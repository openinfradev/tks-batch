package application

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-common/pkg/log"
)

type ApplicationGroup struct {
	ID         string `gorm:"primarykey"`
	Status     string
	StatusDesc string
	WorkflowId string
	UpdatedAt  time.Time
	CreatedAt  time.Time
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

func (x *ApplicationAccessor) GetIncompleteAppGroups() ([]ApplicationGroup, error) {
	var appGroups []ApplicationGroup

	res := x.db.
		Where("status IN ?", []string{"INSTALLING", "DELETING"}).
		Find(&appGroups)

	if res.Error != nil {
		return nil, res.Error
	}

	return appGroups, nil
}

func (x *ApplicationAccessor) UpdateAppGroupStatus(appGroupId string, status string, statusDesc string, workflowId string) error {
	log.Info(fmt.Sprintf("UpdateAppGroupStatus. appGroupId[%s], status[%s], statusDesc[%s], workflowId[%s]", appGroupId, status, statusDesc, workflowId))
	res := x.db.Model(ApplicationGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}
	return nil
}
