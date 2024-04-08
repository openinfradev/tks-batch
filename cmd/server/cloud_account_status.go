package main

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

func processCloudAccountStatus() error {
	// get cloudAccount
	cloudAccounts, err := cloudAccountAccessor.GetIncompleteCloudAccounts()
	if err != nil {
		return err
	}
	if len(cloudAccounts) == 0 {
		return nil
	}
	log.Info(context.TODO(), "cloudAccounts : ", cloudAccounts)

	for i := range cloudAccounts {
		cloudaccount := cloudAccounts[i]

		cloudAccountId := cloudaccount.ID
		workflowId := cloudaccount.WorkflowId
		status := cloudaccount.Status
		statusDesc := cloudaccount.StatusDesc

		// update status
		var newStatus domain.CloudAccountStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow(context.TODO(), "argo", workflowId)
			if err != nil {
				log.Error(context.TODO(), "failed to get argo workflow. err : ", err)
				continue
			}

			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(context.TODO(), fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))

			if status == domain.CloudAccountStatus_CREATING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.CloudAccountStatus_CREATING
				case "Succeeded":
					newStatus = domain.CloudAccountStatus_CREATED
				case "Failed":
					newStatus = domain.CloudAccountStatus_CREATE_ERROR
				case "Error":
					newStatus = domain.CloudAccountStatus_CREATE_ERROR
				}
			} else if status == domain.CloudAccountStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.CloudAccountStatus_DELETING
				case "Succeeded":
					newStatus = domain.CloudAccountStatus_DELETED
				case "Failed":
					newStatus = domain.CloudAccountStatus_DELETE_ERROR
				case "Error":
					newStatus = domain.CloudAccountStatus_DELETE_ERROR
				}
			}
			if newStatus == domain.CloudAccountStatus_PENDING {
				continue
			}
		} else {
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(context.TODO(), fmt.Sprintf("update status!! cloudAccountId [%s], newStatus [%s], newMessage [%s]", cloudAccountId, newStatus, newMessage))
			err := cloudAccountAccessor.UpdateCloudAccountStatus(cloudAccountId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error(context.TODO(), "Failed to update cloudaccount status err : ", err)
				continue
			}

			if newStatus == domain.CloudAccountStatus_CREATED {
				err = cloudAccountAccessor.UpdateCreatedIAM(cloudAccountId, true)
				if err != nil {
					log.Error(context.TODO(), "Failed to update cloudaccount createdIAM err : ", err)
					continue
				}

			}
		}
	}
	return nil
}
