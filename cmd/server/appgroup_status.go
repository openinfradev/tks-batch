package main

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

func processAppGroupStatus(ctx context.Context) error {

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

		appGroupId := appGroup.ID.String()
		workflowId := appGroup.WorkflowId
		status := appGroup.Status
		statusDesc := appGroup.StatusDesc

		// update appgroup status
		var newStatus pb.AppGroupStatus
		var newMessage string
		var err error

		if workflowId != "" {
			var phase wfv1.WorkflowPhase
			phase, newMessage, err = argowfClient.GetWorkflow(ctx, workflowId, "argo")
			if err != nil {
				log.Error("failed to get argo workflow. err : ", err)
				continue
			}
			log.Debug(fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, phase))

			if status == pb.AppGroupStatus_APP_GROUP_INSTALLING {
				switch phase {
				case wfv1.WorkflowRunning:
					newStatus = pb.AppGroupStatus_APP_GROUP_INSTALLING
				case wfv1.WorkflowSucceeded:
					newStatus = pb.AppGroupStatus_APP_GROUP_RUNNING
				case wfv1.WorkflowFailed:
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				case wfv1.WorkflowError:
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				}
			} else if status == pb.AppGroupStatus_APP_GROUP_DELETING {
				switch phase {
				case wfv1.WorkflowRunning:
					newStatus = pb.AppGroupStatus_APP_GROUP_DELETING
				case wfv1.WorkflowSucceeded:
					newStatus = pb.AppGroupStatus_APP_GROUP_DELETED
				case wfv1.WorkflowFailed:
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				case wfv1.WorkflowError:
					newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
				}
			}
			if newStatus == pb.AppGroupStatus_APP_GROUP_UNSPECIFIED {
				continue
			}
		} else {
			newStatus = pb.AppGroupStatus_APP_GROUP_ERROR
			newMessage = "empty workflowId"
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
