package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/openinfradev/tks-common/pkg/argowf"
	mockargo "github.com/openinfradev/tks-common/pkg/argowf/mock"
	"github.com/openinfradev/tks-common/pkg/helper"
	pb "github.com/openinfradev/tks-proto/tks_pb"

	"github.com/openinfradev/tks-batch/internal/cluster"
)

func insertTestCluster(clusterId string, workflowId string, status pb.ClusterStatus, statusDesc string) {
	cluster := cluster.Cluster{
		ID:         clusterId,
		WorkflowId: workflowId,
		Status:     status,
		StatusDesc: statusDesc,
	}
	clusterAccessor.GetDb().Create(&cluster)
}

func getTestCluster(clusterId string) cluster.Cluster {
	var res cluster.Cluster
	clusterAccessor.GetDb().First(&res, "ID = ?", clusterId)
	clusterAccessor.GetDb().Delete(&cluster.Cluster{}, "ID = ?", clusterId)

	return res
}

func TestWorkferClusterStatus(t *testing.T) {
	clusterId := helper.GenerateClusterId()

	testCases := []struct {
		name        string
		setupTest   func()
		buildStubs  func(mockArgoClient *mockargo.MockClient)
		checkResult func(err error)
	}{
		{
			name: "INSTALLING_TO_RUNNING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Succeeded", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_RUNNING)
				require.Equal(t, cluster.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "INSTALLING_TO_INSTALLING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Running", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_INSTALLING)
				require.Equal(t, cluster.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "DELETING_TO_DELETED",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Succeeded", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_DELETED)
				require.Equal(t, cluster.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "DELETING_TO_INSTALLING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Running", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_DELETING)
				require.Equal(t, cluster.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "NOTHING_TO_DO_UNSPECIFIED",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_UNSPECIFIED, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.Status, pb.ClusterStatus_UNSPECIFIED)
			},
		},
		{
			name: "NOTHING_TO_DO_RUNNING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_RUNNING, "running")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.Status, pb.ClusterStatus_RUNNING)
			},
		},
		{
			name: "NOTHING_TO_DO_DELETED",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_DELETED, "deleted")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.Status, pb.ClusterStatus_DELETED)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mocking and injection
			mockArgoClient := mockargo.NewMockClient(ctrl)
			argowfClient = mockArgoClient

			tc.buildStubs(mockArgoClient)

			tc.setupTest()
			err := processClusterStatus(ctx)
			tc.checkResult(err)
		})
	}
}
