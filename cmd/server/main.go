package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/openinfradev/tks-common/pkg/argowf"
	"github.com/openinfradev/tks-common/pkg/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/openinfradev/tks-batch/internal/application"
	"github.com/openinfradev/tks-batch/internal/cluster"
	"github.com/openinfradev/tks-batch/internal/database"
)

const INTERVAL_SEC = 1

var (
	argowfClient        argowf.Client
	clusterAccessor     *cluster.ClusterAccessor
	applicationAccessor *application.ApplicationAccessor
)

func init() {
	flag.Int("port", 9112, "service port")
	flag.String("argo-address", "localhost", "server address for argo-workflow-server")
	flag.Int("argo-port", 2746, "server port for argo-workflow-server")

	flag.String("dbhost", "localhost", "host of postgreSQL")
	flag.String("dbport", "5432", "port of postgreSQL")
	flag.String("dbuser", "postgres", "postgreSQL user")
	flag.String("dbpassword", "password", "password for postgreSQL user")
	flag.String("dbname", "tks", "the name of database")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()
	viper.BindPFlags(pflag.CommandLine)

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

	// initialize external clients
	argowfClient, err = argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
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

		time.Sleep(time.Second * INTERVAL_SEC)
	}

}
