package cluster

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/openinfradev/tks-common/pkg/log"
)

// Cluster represents a kubernetes cluster information.
type Cluster struct {
	ID         string `gorm:"primarykey"`
	WorkflowId string
	Status     string
	StatusDesc string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

// Accessor accesses cluster info in DB.
type ClusterAccessor struct {
	db *gorm.DB
}

// NewClusterAccessor returns new Accessor to access clusters.
func New(db *gorm.DB) *ClusterAccessor {
	return &ClusterAccessor{
		db: db,
	}
}

// For Unittest
func (x *ClusterAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *ClusterAccessor) GetIncompleteClusters() ([]Cluster, error) {
	var clusters []Cluster

	res := x.db.
		Where("status IN ?", []string{"INSTALLING", "DELETING"}).
		Find(&clusters)

	if res.Error != nil {
		return nil, res.Error
	}

	return clusters, nil
}

func (x *ClusterAccessor) UpdateClusterStatus(clusterId string, status string, statusDesc string, workflowId string) error {
	log.Info(fmt.Sprintf("UpdateClusterStatus. clusterId[%s], status[%s], statusDesc[%s], workflowId[%s]", clusterId, status, statusDesc, workflowId))
	res := x.db.Model(Cluster{}).
		Where("ID = ?", clusterId).
		Updates(map[string]interface{}{"Status": status, "StatusDesc": statusDesc, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cluster with id %s", clusterId)
	}
	return nil
}
