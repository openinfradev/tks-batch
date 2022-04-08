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

	"github.com/openinfradev/tks-batch/internal/application"
)

func insertTestAppGroup(appGroupId uuid.UUID, workflowId string, status pb.AppGroupStatus, statusDesc string) {
	appgroup := application.ApplicationGroup{
		ID:         appGroupId,
		WorkflowId: workflowId,
		Status:     status,
		StatusDesc: statusDesc,
	}
	applicationAccessor.GetDb().Create(&appgroup)
}

func getTestAppGroup(appGroupId uuid.UUID) application.ApplicationGroup {
	var res application.ApplicationGroup
	applicationAccessor.GetDb().First(&res, appGroupId)
	applicationAccessor.GetDb().Delete(&application.ApplicationGroup{}, appGroupId)

	return res
}

func TestWorkferAppGroupStatus(t *testing.T) {
	appGroupId := uuid.New()

	testCases := []struct {
		name        string
		setupTest   func()
		buildStubs  func(mockArgoClient *mockargo.MockClient)
		checkResult func(err error)
	}{
		{
			name: "INSTALLING_TO_RUNNING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowSucceeded, "msg_seccueded", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_RUNNING)
				require.Equal(t, appGroup.StatusDesc, "msg_seccueded")
			},
		},
		{
			name: "INSTALLING_TO_INSTALLING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowRunning, "msg_running", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_INSTALLING)
				require.Equal(t, appGroup.StatusDesc, "msg_running")
			},
		},
		{
			name: "DELETING_TO_DELETED",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowSucceeded, "msg_seccueded", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_DELETED)
				require.Equal(t, appGroup.StatusDesc, "msg_seccueded")
			},
		},
		{
			name: "DELETING_TO_INSTALLING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(wfv1.WorkflowRunning, "msg_running", nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_DELETING)
				require.Equal(t, appGroup.StatusDesc, "msg_running")
			},
		},
		{
			name: "NOTHING_TO_DO_UNSPECIFIED",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_UNSPECIFIED, "")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_UNSPECIFIED)
			},
		},
		{
			name: "NOTHING_TO_DO_RUNNING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_RUNNING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_RUNNING)
			},
		},
		{
			name: "NOTHING_TO_DO_DELETED",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_DELETED, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_DELETED)
			},
		},
		{
			name: "ERROR_NO_WORKFLOW_ID",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "", pb.AppGroupStatus_APP_GROUP_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_ERROR)
				require.Equal(t, appGroup.StatusDesc, "empty workflowId")
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
			err := processAppGroupStatus(ctx)
			tc.checkResult(err)
		})
	}
}
