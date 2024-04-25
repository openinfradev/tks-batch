package main

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

func processOrganizationStatus() error {
	// get organizations
	organizations, err := organizationAccessor.GetIncompleteOrganizations()
	if err != nil {
		return err
	}
	if len(organizations) == 0 {
		return nil
	}
	log.Info(context.TODO(), "[processOrganizationStatus] organizations : ", organizations)

	for i := range organizations {
		organization := organizations[i]

		organizationId := organization.ID
		workflowId := organization.WorkflowId
		status := organization.Status
		statusDesc := organization.StatusDesc

		// update status
		var newStatus domain.OrganizationStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow(context.TODO(), "argo", workflowId)
			if err != nil {
				log.Error(context.TODO(), "failed to get argo workflow. err : ", err)
				continue
			}

			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(context.TODO(), fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))

			if status == domain.OrganizationStatus_CREATING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.OrganizationStatus_CREATING
				case "Succeeded":
					newStatus = domain.OrganizationStatus_CREATED
				case "Failed":
					newStatus = domain.OrganizationStatus_ERROR
				case "Error":
					newStatus = domain.OrganizationStatus_ERROR
				}
			} else if status == domain.OrganizationStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.OrganizationStatus_DELETING
				case "Succeeded":
					newStatus = domain.OrganizationStatus_DELETED
				case "Failed":
					newStatus = domain.OrganizationStatus_ERROR
				case "Error":
					newStatus = domain.OrganizationStatus_ERROR
				}
			}
			if newStatus == domain.OrganizationStatus_PENDING {
				continue
			}
		} else {
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(context.TODO(), fmt.Sprintf("update status!! organizationId [%s], newStatus [%s], newMessage [%s]", organizationId, newStatus, newMessage))
			err := organizationAccessor.UpdateOrganizationStatus(organizationId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error(context.TODO(), "Failed to update organization status err : ", err)
				continue
			}
		}
	}
	return nil
}
