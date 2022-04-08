package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

type ApplicationGroup struct {
	ID         uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
	Status     pb.AppGroupStatus
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

func (x *ApplicationAccessor) GetIncompleteAppGroups() ([]ApplicationGroup, error) {
	var appGroups []ApplicationGroup

	res := x.db.
		Where("status IN ?", []pb.AppGroupStatus{pb.AppGroupStatus_APP_GROUP_INSTALLING, pb.AppGroupStatus_APP_GROUP_DELETING}).
		Find(&appGroups)

	if res.Error != nil {
		return nil, res.Error
	}

	return appGroups, nil
}

func (x *ApplicationAccessor) UpdateAppGroupStatus(appGroupId string, status pb.AppGroupStatus, statusDesc string, workflowId string) error {
	log.Info(fmt.Sprintf("UpdateAppGroupStatus. appGroupId[%s], status[%s], statusDesc[%s], workflowId[%s]", appGroupId, status, statusDesc, workflowId))
	res := x.db.Model(ApplicationGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}
	return nil
}
