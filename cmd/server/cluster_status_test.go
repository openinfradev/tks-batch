package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	mockargo "github.com/openinfradev/tks-common/pkg/argowf/mock"
	pb "github.com/openinfradev/tks-proto/tks_pb"

	"github.com/openinfradev/tks-batch/internal/cluster"
)

func insertTestCluster(clusterId uuid.UUID, workflowId string, status pb.ClusterStatus, statusDesc string) {
	cluster := cluster.Cluster{
		ID:         clusterId,
		WorkflowId: workflowId,
		Status:     status,
		StatusDesc: statusDesc,
	}
	clusterAccessor.GetDb().Create(&cluster)
}

func getTestCluster(clusterId uuid.UUID) cluster.Cluster {
	var res cluster.Cluster
	clusterAccessor.GetDb().First(&res, clusterId)
	clusterAccessor.GetDb().Delete(&cluster.Cluster{}, clusterId)

	return res
}

func TestWorkferClusterStatus(t *testing.T) {
	clusterId := uuid.New()

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
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowSucceeded, "msg_seccueded", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_RUNNING)
				require.Equal(t, cluster.StatusDesc, "msg_seccueded")
			},
		},
		{
			name: "INSTALLING_TO_INSTALLING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowRunning, "msg_running", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_INSTALLING)
				require.Equal(t, cluster.StatusDesc, "msg_running")
			},
		},
		{
			name: "DELETING_TO_DELETED",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowSucceeded, "msg_seccueded", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_DELETED)
				require.Equal(t, cluster.StatusDesc, "msg_seccueded")
			},
		},
		{
			name: "DELETING_TO_INSTALLING",
			setupTest: func() {
				insertTestCluster(clusterId, "WORKFLOWID", pb.ClusterStatus_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowRunning, "msg_running", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.WorkflowId, "WORKFLOWID")
				require.Equal(t, cluster.Status, pb.ClusterStatus_DELETING)
				require.Equal(t, cluster.StatusDesc, "msg_running")
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
		{
			name: "ERROR_NO_WORKFLOW_ID",
			setupTest: func() {
				insertTestCluster(clusterId, "", pb.ClusterStatus_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				cluster := getTestCluster(clusterId)
				require.Equal(t, cluster.Status, pb.ClusterStatus_ERROR)
				require.Equal(t, cluster.StatusDesc, "empty workflowId")
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
