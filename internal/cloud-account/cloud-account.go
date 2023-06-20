package cloudAccount

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type CloudAccount struct {
	ID         string `gorm:"primarykey"`
	WorkflowId string
	Status     domain.CloudAccountStatus
	StatusDesc string
}

type CloudAccountAccessor struct {
	db *gorm.DB
}

func New(db *gorm.DB) *CloudAccountAccessor {
	return &CloudAccountAccessor{
		db: db,
	}
}

// For Unittest
func (x *CloudAccountAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *CloudAccountAccessor) GetIncompleteCloudAccounts() ([]CloudAccount, error) {
	var cloudAccounts []CloudAccount

	res := x.db.
		Where("status IN ?", []domain.CloudAccountStatus{domain.CloudAccountStatus_CREATING, domain.CloudAccountStatus_DELETING}).
		Find(&cloudAccounts)

	if res.Error != nil {
		return nil, res.Error
	}

	return cloudAccounts, nil
}

func (x *CloudAccountAccessor) UpdateCloudAccountStatus(cloudAccountId string, status domain.CloudAccountStatus, statusDesc string, workflowId string) error {
	log.Info(fmt.Sprintf("UpdateCloudAccountStatus. cloudAccountId[%s], status[%d], statusDesc[%s], workflowId[%s]", cloudAccountId, status, statusDesc, workflowId))
	res := x.db.Model(CloudAccount{}).
		Where("ID = ?", cloudAccountId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cloudAccount with id %s", cloudAccountId)
	}
	return nil
}

func (x *CloudAccountAccessor) UpdateCreatedIAM(cloudAccountId string, createdIAM bool) error {
	log.Info(fmt.Sprintf("UpdateCreatedIAM. cloudAccountId[%s], createdIAM[%t]", cloudAccountId, createdIAM))
	res := x.db.Model(CloudAccount{}).
		Where("ID = ?", cloudAccountId).
		Updates(map[string]interface{}{"CreatedIAM": createdIAM})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cloudAccount with id %s", cloudAccountId)
	}
	return nil
}
