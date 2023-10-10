package main

import (
	"flag"
	"fmt"
	"time"

	argo "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/openinfradev/tks-batch/internal/application"
	cloudAccount "github.com/openinfradev/tks-batch/internal/cloud-account"
	"github.com/openinfradev/tks-batch/internal/cluster"
	"github.com/openinfradev/tks-batch/internal/database"
	"github.com/openinfradev/tks-batch/internal/organization"
)

const INTERVAL_SEC = 5

var (
	argowfClient         argo.ArgoClient
	clusterAccessor      *cluster.ClusterAccessor
	applicationAccessor  *application.ApplicationAccessor
	cloudAccountAccessor *cloudAccount.CloudAccountAccessor
	organizationAccessor *organization.OrganizationAccessor
)

func init() {
	flag.Int("port", 9112, "service port")
	flag.String("argo-address", "localhost", "server address for argo-workflow-server")
	flag.Int("argo-port", 2746, "server port for argo-workflow-server")
	flag.String("tks-api-address", "http://localhost", "server address for tks-api")
	flag.Int("tks-api-port", 8080, "server port number for tks-api")
	flag.String("tks-api-account", "", "account name for tks-api")
	flag.String("tks-api-password", "", "the password for tks-api account")

	flag.String("dbhost", "localhost", "host of postgreSQL")
	flag.String("dbport", "5432", "port of postgreSQL")
	flag.String("dbuser", "postgres", "postgreSQL user")
	flag.String("dbpassword", "password", "password for postgreSQL user")
	flag.String("dbname", "tks", "the name of database")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Error("Failed to bindFlags ", err)
	}

}

func main() {
	log.Info("*** Arguments *** ")
	for i, s := range viper.AllSettings() {
		log.Info(fmt.Sprintf("%s : %v", i, s))
	}
	log.Info("****************** ")

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("cannot connect gormDB")
	}
	clusterAccessor = cluster.New(db)
	applicationAccessor = application.New(db)
	cloudAccountAccessor = cloudAccount.New(db)
	organizationAccessor = organization.New(db)

	// initialize external clients
	argowfClient, err = argo.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
	if err != nil {
		log.Fatal("failed to create argowf client : ", err)
	}

	for {
		err = processClusterStatus()
		if err != nil {
			log.Error(err)
		}
		err = processAppGroupStatus()
		if err != nil {
			log.Error(err)
		}
		err = processCloudAccountStatus()
		if err != nil {
			log.Error(err)
		}
		err = processOrganizationStatus()
		if err != nil {
			log.Error(err)
		}
		err = processClusterByoh()
		if err != nil {
			log.Error(err)
		}

		time.Sleep(time.Second * INTERVAL_SEC)
	}

}
