package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	_apiClient "github.com/openinfradev/tks-api/pkg/api-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

var token string

func processClusterByoh() error {
	// get clusters
	clusters, err := clusterAccessor.GetBootstrappedByohClusters()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return nil
	}
	log.Info("byoh clusters : ", clusters)

	token = getTksApiToken()
	for _, cluster := range clusters {
		clusterId := cluster.ID

		// check agent node
		apiClient, err := _apiClient.New(fmt.Sprintf("%s:%d", viper.GetString("tks-api-address"), viper.GetInt("tks-api-port")), token)
		if err != nil {
			log.Error(err)
			continue
		}

		url := fmt.Sprintf("clusters/%s/nodes", clusterId)
		body, err := apiClient.Get(url)
		if err != nil {
			log.Error(err)
			continue
		}

		var out domain.GetClusterNodesResponse
		transcode(body, &out)

		completed := true
		for _, node := range out.Nodes {
			if node.Status != "COMPLETED" {
				completed = false
			}
		}
		log.Info(out.Nodes)

		//completed = true // FOR TEST
		if completed {
			log.Info(fmt.Sprintf("all agents registered! starting stack creation. clusterId %s", clusterId))
			if err = clusterAccessor.UpdateClusterStatus(clusterId, domain.ClusterStatus_INSTALLING); err != nil {
				log.Error("Failed to update cluster status err : ", err)
				continue
			}

			apiClient, err := _apiClient.New(fmt.Sprintf("%s:%d", viper.GetString("tks-api-address"), viper.GetInt("tks-api-port")), token)
			if err != nil {
				log.Error(err)
				continue
			}

			if cluster.IsStack {
				if _, err = apiClient.Post(fmt.Sprintf("organizations/%s/stacks/%s/install", cluster.OrganizationId, clusterId), nil); err != nil {
					log.Error(err)
					continue
				}
			} else {
				if _, err = apiClient.Post("clusters/"+clusterId+"/install", nil); err != nil {
					log.Error(err)
					continue
				}
			}

		}
	}
	return nil
}

func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(in)
	if err != nil {
		fmt.Println(err)
	}
	err = json.NewDecoder(buf).Decode(out)
	if err != nil {
		fmt.Println(err)
	}
}

func getTksApiToken() string {
	apiClient, err := _apiClient.New(fmt.Sprintf("%s:%d", viper.GetString("tks-api-address"), viper.GetInt("tks-api-port")), "")
	if err != nil {
		log.Error(err)
		return ""
	}

	_, err = apiClient.Post("auth/ping", domain.PingTokenRequest{
		Token:          token,
		OrganizationId: "master",
	})
	if err != nil {
		body, err := apiClient.Post("auth/login", domain.LoginRequest{
			AccountId:      viper.GetString("tks-api-account"),
			Password:       viper.GetString("tks-api-password"),
			OrganizationId: "master",
		})
		if err != nil {
			return ""
		}

		var out domain.LoginResponse
		transcode(body, &out)

		log.Info(out.User.Token)
		token = out.User.Token
	}

	return token
}
