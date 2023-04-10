package main

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

func processClusterStatus(ctx context.Context) error {
	// get clusters
	clusters, err := clusterAccessor.GetIncompleteClusters()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return nil
	}
	log.Info("clusters : ", clusters)

	for i := range clusters {
		cluster := clusters[i]

		clusterId := cluster.ID
		workflowId := cluster.WorkflowId
		status := cluster.Status
		statusDesc := cluster.StatusDesc

		// update status
		var newStatus pb.ClusterStatus
		var newMessage string

		if workflowId != "" {
			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error("failed to get argo workflow. err : ", err)
				continue
			}

			newMessage = fmt.Sprintf("(%s) %s", workflow.Status.Progress, workflow.Status.Message)
			log.Debug(fmt.Sprintf("status [%s], newMessage [%s], phase [%s]", status, newMessage, workflow.Status.Phase))

			if status == pb.ClusterStatus_INSTALLING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = pb.ClusterStatus_INSTALLING
				case "Succeeded":
					newStatus = pb.ClusterStatus_RUNNING
				case "Failed":
					newStatus = pb.ClusterStatus_ERROR
				case "Error":
					newStatus = pb.ClusterStatus_ERROR
				}
			} else if status == pb.ClusterStatus_DELETING {
				switch workflow.Status.Phase {
				case "Running":
					newStatus = pb.ClusterStatus_DELETING
				case "Succeeded":
					newStatus = pb.ClusterStatus_DELETED
				case "Failed":
					newStatus = pb.ClusterStatus_ERROR
				case "Error":
					newStatus = pb.ClusterStatus_ERROR
				}
			}
			if newStatus == pb.ClusterStatus_UNSPECIFIED {
				continue
			}
		} else {
			// [TODO] READY 상태를 추가하도록 할 것
			continue
		}

		if status != newStatus || statusDesc != newMessage {
			log.Debug(fmt.Sprintf("update status!! clusterId [%s], newStatus [%s], newMessage [%s]", clusterId, newStatus, newMessage))
			err := clusterAccessor.UpdateClusterStatus(clusterId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error("Failed to update cluster status err : ", err)
				continue
			}
		}
	}
	return nil
}
