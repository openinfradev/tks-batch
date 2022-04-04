package main

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

func workferClusterStatus(ctx context.Context) error {
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

		clusterId := cluster.ID.String()
		workflowId := cluster.WorkflowId
		status := cluster.Status
		statusDesc := cluster.StatusDesc

		// update status
		var newStatus pb.ClusterStatus
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

			if status == pb.ClusterStatus_INSTALLING {
				switch phase {
				case wfv1.WorkflowRunning:
					newStatus = pb.ClusterStatus_INSTALLING
				case wfv1.WorkflowSucceeded:
					newStatus = pb.ClusterStatus_RUNNING
				case wfv1.WorkflowFailed:
					newStatus = pb.ClusterStatus_ERROR
				case wfv1.WorkflowError:
					newStatus = pb.ClusterStatus_ERROR
				}
			} else if status == pb.ClusterStatus_DELETING {
				switch phase {
				case wfv1.WorkflowRunning:
					newStatus = pb.ClusterStatus_DELETING
				case wfv1.WorkflowSucceeded:
					newStatus = pb.ClusterStatus_DELETED
				case wfv1.WorkflowFailed:
					newStatus = pb.ClusterStatus_ERROR
				case wfv1.WorkflowError:
					newStatus = pb.ClusterStatus_ERROR
				}
			}
			if newStatus == pb.ClusterStatus_UNSPECIFIED {
				continue
			}
		} else {
			newStatus = pb.ClusterStatus_ERROR
			newMessage = "empty workflowId"
		}

		if status != newStatus || statusDesc != newMessage {
			log.Info(fmt.Sprintf("update status!! clusterId [%s], newStatus [%s], newMessage [%s]", clusterId, newStatus, newMessage))
			err := clusterAccessor.UpdateClusterStatus(clusterId, newStatus, newMessage, workflowId)
			if err != nil {
				log.Error("Failed to update cluster status err : ", err)
				continue
			}
		}
	}
	return nil
}
