package stack

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Stack struct {
	ID         string `gorm:"primarykey"`
	WorkflowId string
	Status     domain.StackStatus
	StatusDesc string
}

type StackAccessor struct {
	db *gorm.DB
}

func New(db *gorm.DB) *StackAccessor {
	return &StackAccessor{
		db: db,
	}
}

// For Unittest
func (x *StackAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *StackAccessor) GetIncompleteStacks() ([]Stack, error) {
	var stacks []Stack

	res := x.db.
		Where("status IN ?", []domain.StackStatus{domain.StackStatus_INSTALLING, domain.StackStatus_DELETING}).
		Find(&stacks)

	if res.Error != nil {
		return nil, res.Error
	}

	return stacks, nil
}

func (x *StackAccessor) UpdateStackStatus(stackId string, status domain.StackStatus, statusDesc string, workflowId string) error {
	log.Info(fmt.Sprintf("UpdateStackStatus. stackId[%s], status[%d], statusDesc[%s], workflowId[%s]", stackId, status, statusDesc, workflowId))
	res := x.db.Model(Stack{}).
		Where("ID = ?", stackId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in stack with id %s", stackId)
	}
	return nil
}
