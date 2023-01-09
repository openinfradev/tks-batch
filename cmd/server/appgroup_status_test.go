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

	"github.com/openinfradev/tks-batch/internal/application"
)

func insertTestAppGroup(appGroupId string, workflowId string, status pb.AppGroupStatus, statusDesc string) {
	appgroup := application.ApplicationGroup{
		ID:         appGroupId,
		WorkflowId: workflowId,
		Status:     status,
		StatusDesc: statusDesc,
	}
	applicationAccessor.GetDb().Create(&appgroup)
}

func getTestAppGroup(appGroupId string) application.ApplicationGroup {
	var res application.ApplicationGroup
	applicationAccessor.GetDb().First(&res, "ID = ?", appGroupId)
	applicationAccessor.GetDb().Delete(&application.ApplicationGroup{}, "ID = ?", appGroupId)

	return res
}

func TestWorkferAppGroupStatus(t *testing.T) {
	appGroupId := helper.GenerateApplicaionGroupId()

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
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Succeeded", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_RUNNING)
				require.Equal(t, appGroup.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "INSTALLING_TO_INSTALLING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_INSTALLING, "installing")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Running", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_INSTALLING)
				require.Equal(t, appGroup.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "DELETING_TO_DELETED",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Succeeded", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_DELETED)
				require.Equal(t, appGroup.StatusDesc, "(0/1) message")
			},
		},
		{
			name: "DELETING_TO_INSTALLING",
			setupTest: func() {
				insertTestAppGroup(appGroupId, "WORKFLOWID", pb.AppGroupStatus_APP_GROUP_DELETING, "deleting")
			},
			buildStubs: func(mockArgoClient *mockargo.MockClient) {
				mockArgoClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any()).
					Return(&argowf.Workflow{Status: argowf.WorkflowStatus{Phase: "Running", Progress: "0/1", Message: "message"}}, nil)
			},
			checkResult: func(err error) {
				require.NoError(t, err)

				appGroup := getTestAppGroup(appGroupId)
				require.Equal(t, appGroup.WorkflowId, "WORKFLOWID")
				require.Equal(t, appGroup.Status, pb.AppGroupStatus_APP_GROUP_DELETING)
				require.Equal(t, appGroup.StatusDesc, "(0/1) message")
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
