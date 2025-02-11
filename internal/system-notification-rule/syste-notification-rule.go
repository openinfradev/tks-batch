package systemNotification

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gofrs/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Organization struct {
	ID               string `gorm:"primarykey"`
	Name             string
	PrimaryClusterId string
}

type SystemNotificationMetricParameter struct {
	gorm.Model

	SystemNotificationTemplateId uuid.UUID
	Order                        int
	Key                          string
	Value                        string
}

type SystemNotificationTemplate struct {
	gorm.Model

	ID               uuid.UUID      `gorm:"primarykey"`
	Name             string         `gorm:"index:idx_name,unique"`
	NotificationType string         `gorm:"default:SYSTEM_NOTIFICATION"`
	Organizations    []Organization `gorm:"many2many:system_notification_template_organizations;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	OrganizationIds  []string       `gorm:"-:all"`
	Description      string
	MetricQuery      string
	MetricParameters []SystemNotificationMetricParameter `gorm:"foreignKey:SystemNotificationTemplateId;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

type SystemNotificationCondition struct {
	gorm.Model

	SystemNotificationRuleId uuid.UUID
	Order                    int
	Severity                 string
	Duration                 string
	Parameter                datatypes.JSON
	Parameters               []domain.SystemNotificationParameter `gorm:"-:all"`
	EnableEmail              bool                                 `gorm:"default:false"`
	EnablePortal             bool                                 `gorm:"default:true"`
}

type SystemNotificationRule struct {
	gorm.Model

	ID                           uuid.UUID `gorm:"primarykey"`
	Name                         string    `gorm:"index,unique"`
	NotificationType             string    `gorm:"default:SYSTEM_NOTIFICATION"`
	Description                  string
	OrganizationId               string
	Organization                 Organization               `gorm:"foreignKey:OrganizationId"`
	SystemNotificationTemplate   SystemNotificationTemplate `gorm:"foreignKey:SystemNotificationTemplateId"`
	SystemNotificationTemplateId string
	SystemNotificationCondition  SystemNotificationCondition `gorm:"foreignKey:SystemNotificationRuleId"`
	MessageTitle                 string
	MessageContent               string
	MessageActionProposal        string
	Status                       domain.SystemNotificationRuleStatus
	CreatorId                    *uuid.UUID `gorm:"type:uuid"`
}

type SystemNotificationAccessor struct {
	db *gorm.DB
}

// NewSystemNotificationAccessor returns new Accessor to access clusters.
func New(db *gorm.DB) *SystemNotificationAccessor {
	return &SystemNotificationAccessor{
		db: db,
	}
}

// For Unittest
func (x *SystemNotificationAccessor) GetDb() *gorm.DB {
	return x.db
}

func (x *SystemNotificationAccessor) GetIncompletedRules() ([]SystemNotificationRule, error) {
	var rules []SystemNotificationRule

	res := x.db.Model(&SystemNotificationRule{}).
		Preload(clause.Associations).
		Preload("SystemNotificationTemplate.MetricParameters").
		Joins("join organizations on organizations.id = system_notification_rules.organization_id").
		Joins("join clusters on clusters.id = organizations.primary_cluster_id AND clusters.status = ?", domain.ClusterStatus_RUNNING).
		Joins("join app_groups on app_groups.cluster_id = clusters.id AND app_groups.status = ?", domain.AppGroupStatus_RUNNING).
		Where("system_notification_rules.status = ?", domain.SystemNotificationRuleStatus_PENDING).
		Unscoped().
		Find(&rules)

	if res.Error != nil {
		return nil, res.Error
	}

	return rules, nil
}

func (x *SystemNotificationAccessor) GetRecentlyUpdatedOrganizations(lastUpdateMin int) ([]string, error) {
	var organizationIds []string

	res := x.db.Model(&SystemNotificationRule{}).
		Select("system_notification_rules.organization_id").
		Joins("join organizations on organizations.id = system_notification_rules.organization_id").
		Joins("join clusters on clusters.id = organizations.primary_cluster_id AND clusters.status = ?", domain.ClusterStatus_RUNNING).
		Joins("join app_groups on app_groups.cluster_id = clusters.id AND app_groups.status = ?", domain.AppGroupStatus_RUNNING).
		Where("system_notification_rules.status = ?", domain.SystemNotificationRuleStatus_APPLIED).
		Where(fmt.Sprintf("system_notification_rules.updated_at between now()-interval '%d minutes' and now() OR system_notification_rules.deleted_at between now()-interval '%d minutes' and now()", lastUpdateMin, lastUpdateMin)).
		Group("system_notification_rules.organization_id").
		Unscoped().
		Find(&organizationIds)

	if res.Error != nil {
		return nil, res.Error
	}
	return organizationIds, nil
}

func (x *SystemNotificationAccessor) GetRules(organizationId string) ([]SystemNotificationRule, error) {
	var rules []SystemNotificationRule

	res := x.db.Model(&SystemNotificationRule{}).
		Preload(clause.Associations).
		Preload("SystemNotificationTemplate.MetricParameters").
		Joins("join organizations on organizations.id = system_notification_rules.organization_id").
		Where("organization_id = ?", organizationId).
		Find(&rules)

	if res.Error != nil {
		return nil, res.Error
	}

	return rules, nil
}

func (x SystemNotificationAccessor) UpdateSystemNotificationRuleStatus(organizationId string, status domain.SystemNotificationRuleStatus) error {
	log.Info(context.TODO(), fmt.Sprintf("organizationId[%v], status[%d]", organizationId, status))
	res := x.db.Model(SystemNotificationRule{}).
		Where("organization_id = ?", organizationId).
		Unscoped().
		Updates(map[string]interface{}{"Status": status})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in SystemNotificationRuleStatus with organizationId %s", organizationId)
	}
	return nil
}
