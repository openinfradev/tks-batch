package main

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

func processClusterStatus() error {
	// get clusters
	clusters, err := clusterAccessor.GetIncompleteClusters()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return nil
	}
	log.Info(context.TODO(), "[processClusterStatus] clusters : ", clusters)

	for i := range clusters {
		cluster := clusters[i]

		clusterId := cluster.ID
		workflowId := cluster.WorkflowId
		status := cluster.Status
		statusDesc := cluster.StatusDesc

		// update status
		var newStatus domain.ClusterStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow(context.TODO(), "argo", workflowId)
			if err != nil {
				log.Error(context.TODO(), "failed to get argo workflow. err : ", err)
				continue
			}

			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(context.TODO(), fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))

			if status == domain.ClusterStatus_INSTALLING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.ClusterStatus_INSTALLING

					paused, err := argowfClient.IsPausedWorkflow(context.TODO(), "argo", workflowId)
					if err == nil && paused {
						newStatus = domain.ClusterStatus_STOPPED
					}
				case "Stopped":
					newStatus = domain.ClusterStatus_STOPPED
				case "Succeeded":
					newStatus = domain.ClusterStatus_RUNNING
				case "Failed":
					newStatus = domain.ClusterStatus_INSTALL_ERROR
				case "Error":
					newStatus = domain.ClusterStatus_INSTALL_ERROR
				}
			} else if status == domain.ClusterStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.ClusterStatus_DELETING
				case "Succeeded":
					newStatus = domain.ClusterStatus_DELETED
				case "Failed":
					newStatus = domain.ClusterStatus_DELETE_ERROR
				case "Error":
					newStatus = domain.ClusterStatus_DELETE_ERROR
				}
			} else if status == domain.ClusterStatus_BOOTSTRAPPING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = domain.ClusterStatus_BOOTSTRAPPING
				case "Succeeded":
					newStatus = domain.ClusterStatus_BOOTSTRAPPED
				case "Failed":
					newStatus = domain.ClusterStatus_BOOTSTRAP_ERROR
				case "Error":
					newStatus = domain.ClusterStatus_BOOTSTRAP_ERROR
				}
			}
			if newStatus == domain.ClusterStatus_PENDING {
				continue
			}
		} else {
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(context.TODO(), fmt.Sprintf("update status!! clusterId [%s], newStatus [%s], newMessage [%s]", clusterId, newStatus, newMessage))
			err := clusterAccessor.UpdateClusterStatus(clusterId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error(context.TODO(), "Failed to update cluster status err : ", err)
				continue
			}
		}
	}
	return nil
}
