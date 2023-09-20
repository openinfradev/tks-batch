package main

import (
	"fmt"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

func processStackStatus() error {
	// get stacks
	stacks, err := stackAccessor.GetIncompleteStacks()
	if err != nil {
		return err
	}
	if len(stacks) == 0 {
		return nil
	}
	log.Info("stacks : ", stacks)

	for i := range stacks {
		stack := stacks[i]

		stackId := stack.ID
		workflowId := stack.WorkflowId
		status := stack.Status
		statusDesc := stack.StatusDesc

		// update status
		var newStatus domain.StackStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error("failed to get argo workflow. err : ", err)
				continue
			}

			log.Info(workflow)
			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))

			if status == domain.StackStatus_INSTALLING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.StackStatus_INSTALLING
				case "Succeeded":
					newStatus = domain.StackStatus_RUNNING
				case "Failed":
					newStatus = domain.StackStatus_INSTALL_ERROR
				case "Error":
					newStatus = domain.StackStatus_INSTALL_ERROR
				}
			} else if status == domain.StackStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.StackStatus_DELETING
				case "Succeeded":
					newStatus = domain.StackStatus_DELETED
				case "Failed":
					newStatus = domain.StackStatus_DELETE_ERROR
				case "Error":
					newStatus = domain.StackStatus_DELETE_ERROR
				}
			}
			if newStatus == domain.StackStatus_PENDING {
				continue
			}
		} else {
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(fmt.Sprintf("update status!! stackId [%s], newStatus [%s], newMessage [%s]", stackId, newStatus, newMessage))
			err := stackAccessor.UpdateStackStatus(stackId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error("Failed to update stack status err : ", err)
				continue
			}
		}
	}
	return nil
}
