package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-common/pkg/argowf"
	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"

	"github.com/openinfradev/tks-batch/internal/application"
	"github.com/openinfradev/tks-batch/internal/cluster"
)

const MAX_WORKERS = 10
const INTERVAL_SEC = 1

var (
	argowfClient        argowf.Client
	clusterAccessor     *cluster.ClusterAccessor
	applicationAccessor *application.ApplicationAccessor
)

var (
	port        int
	argoAddress string
	argoPort    int

	dbhost     string
	dbport     string
	dbuser     string
	dbpassword string
)

func init() {
	flag.IntVar(&port, "port", 9112, "service port")
	flag.StringVar(&argoAddress, "argo-address", "argowf-test-1641626969.ap-northeast-2.elb.amazonaws.com", "server address for argo-workflow-server")
	flag.IntVar(&argoPort, "argo-port", 2746, "server port for argo-workflow-server")

	flag.StringVar(&dbhost, "dbhost", "localhost", "host of postgreSQL")
	flag.StringVar(&dbport, "dbport", "5432", "port of postgreSQL")
	flag.StringVar(&dbuser, "dbuser", "postgres", "postgreSQL user")
	flag.StringVar(&dbpassword, "dbpassword", "password", "password for postgreSQL user")
}

type ClusterStatus struct {
	ClusterId  string
	WorkflowId string
	Status     pb.ClusterStatus
	StatusDesc string
}

type AppGroupStatus struct {
	AppGroupId string
	WorkflowId string
	Status     pb.AppGroupStatus
	StatusDesc string
}

func main() {
	flag.Parse()

	log.Info("*** Arguments *** ")
	log.Info("argoAddress : ", argoAddress)
	log.Info("argoPort : ", argoPort)
	log.Info("dbhost : ", dbhost)
	log.Info("dbport : ", dbport)
	log.Info("dbuser : ", dbuser)
	log.Info("dbpassword : ", dbpassword)
	log.Info("****************** ")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// initialize database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=tks port=%s sslmode=disable TimeZone=Asia/Seoul", dbhost, dbuser, dbpassword, dbport)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open database ", err)
	}
	clusterAccessor = cluster.New(db)
	applicationAccessor = application.New(db)

	// initialize external clients
	argowfClient, err = argowf.New(argoAddress, argoPort)
	if err != nil {
		log.Fatal("failed to create argowf client : ", err)
	}

	// initialize workers
	chanClusters := make(chan ClusterStatus)
	chanAppGroups := make(chan AppGroupStatus)

	for i := 0; i < MAX_WORKERS; i++ {
		go workferClusterStatus(ctx, i, chanClusters)
		go workferAppGroupStatus(ctx, i, chanAppGroups)
	}

	for {
		processClusterStatus(chanClusters)
		processAppGroupStatus(chanAppGroups)

		time.Sleep(time.Second * INTERVAL_SEC)

	}
}

func processClusterStatus(chanClusters chan ClusterStatus) {
	clusters, err := clusterAccessor.GetIncompleteClusters()
	if err != nil {
		log.Error("ERROR ", err)
		return
	}
	if len(clusters) > 0 {
		log.Info("clusters : ", clusters)
	}

	for _, cluster := range clusters {
		clusterId := cluster.ID.String()
		workflowId := cluster.WorkflowId
		status := cluster.Status
		statusDesc := cluster.StatusDesc

		if status == pb.ClusterStatus_INSTALLING || status == pb.ClusterStatus_DELETING {
			chanClusters <- ClusterStatus{ClusterId: clusterId, WorkflowId: workflowId, Status: status, StatusDesc: statusDesc}
		}
	}
}

func processAppGroupStatus(chanAppGroups chan AppGroupStatus) {
	appGroups, err := applicationAccessor.GetIncompleteAppGroups()
	if err != nil {
		log.Error("ERROR ", err)
		return
	}
	if len(appGroups) > 0 {
		log.Info("appGroups : ", appGroups)
	}

	for _, appGroup := range appGroups {
		appGroupId := appGroup.ID.String()
		workflowId := appGroup.WorkflowId
		status := appGroup.Status
		statusDesc := appGroup.StatusDesc

		if status == pb.AppGroupStatus_APP_GROUP_INSTALLING || status == pb.AppGroupStatus_APP_GROUP_DELETING {
			chanAppGroups <- AppGroupStatus{AppGroupId: appGroupId, WorkflowId: workflowId, Status: status, StatusDesc: statusDesc}
		}
	}
}
