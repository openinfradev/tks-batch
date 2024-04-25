package main

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

func processAppGroupStatus() error {

	// get appgroups
	appGroups, err := applicationAccessor.GetIncompleteAppGroups()
	if err != nil {
		return err
	}
	if len(appGroups) == 0 {
		return nil
	}
	log.Info(context.TODO(), "[processAppGroupStatus] appGroups : ", appGroups)

	for i := range appGroups {
		appGroup := appGroups[i]

		appGroupId := appGroup.ID
		workflowId := appGroup.WorkflowId
		status := appGroup.Status
		statusDesc := appGroup.StatusDesc

		// update appgroup status
		var newStatus domain.AppGroupStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow(context.TODO(), "argo", workflowId)
			if err != nil {
				log.Error(context.TODO(), "failed to get argo workflow. err : ", err)
				continue
			}
			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(context.TODO(), fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))
			if status == domain.AppGroupStatus_INSTALLING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.AppGroupStatus_INSTALLING
				case "Succeeded":
					newStatus = domain.AppGroupStatus_RUNNING
				case "Failed":
					newStatus = domain.AppGroupStatus_INSTALL_ERROR
				case "Error":
					newStatus = domain.AppGroupStatus_INSTALL_ERROR
				}
			} else if status == domain.AppGroupStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.AppGroupStatus_DELETING
				case "Succeeded":
					newStatus = domain.AppGroupStatus_DELETED
				case "Failed":
					newStatus = domain.AppGroupStatus_DELETE_ERROR
				case "Error":
					newStatus = domain.AppGroupStatus_DELETE_ERROR
				}
			}
			if newStatus == domain.AppGroupStatus_PENDING {
				continue
			}
		} else {
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(context.TODO(), fmt.Sprintf("update status!! appGroupId [%s], newStatus [%s], newMessage [%s]", appGroupId, newStatus, newMessage))
			err := applicationAccessor.UpdateAppGroupStatus(appGroupId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error(context.TODO(), "Failed to update appgroup status err : ", err)
				continue
			}
		}
	}
	return nil
}
