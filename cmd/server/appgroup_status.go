package main

import (
	"fmt"

	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
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
	log.Info("appGroups : ", appGroups)

	for i := range appGroups {
		appGroup := appGroups[i]

		appGroupId := appGroup.ID
		workflowId := appGroup.WorkflowId
		status := appGroup.Status
		statusDesc := appGroup.StatusDesc

		// update appgroup status
		var newStatus pb.AppGroupStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error("failed to get argo workflow. err : ", err)
				continue
			}
			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))
			if status == pb.AppGroupStatus_APP_GROUP_INSTALLING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = pb.AppGroupStatus_APP_GROUP_INSTALLING
				case "Succeeded":
					newStatus = pb.AppGroupStatus_APP_GROUP_RUNNING
				case "Failed":
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				case "Error":
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				}
			} else if status == pb.AppGroupStatus_APP_GROUP_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = pb.AppGroupStatus_APP_GROUP_DELETING
				case "Succeeded":
					newStatus = pb.AppGroupStatus_APP_GROUP_DELETED
				case "Failed":
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				case "Error":
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				}
			}
			if newStatus == pb.AppGroupStatus_APP_GROUP_UNSPECIFIED {
				continue
			}
		} else {
			// [TODO] READY 상태를 추가하도록 할 것
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(fmt.Sprintf("update status!! appGroupId [%s], newStatus [%s], newMessage [%s]", appGroupId, newStatus, newMessage))
			err := applicationAccessor.UpdateAppGroupStatus(appGroupId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error("Failed to update appgroup status err : ", err)
				continue
			}
		}
	}
	return nil
}
