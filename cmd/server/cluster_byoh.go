package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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
	log.Info(context.TODO(), "[processClusterByoh] byoh clusters : ", clusters)

	token = getTksApiToken()
	if token != "" {
		apiClient.SetToken(token)
	}
	for _, cluster := range clusters {
		clusterId := cluster.ID

		// check agent node
		url := fmt.Sprintf("clusters/%s/nodes", clusterId)
		body, err := apiClient.Get(url)
		if err != nil {
			log.Error(context.TODO(), err)
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
		log.Info(context.TODO(), out.Nodes)

		//completed = true // FOR TEST
		if completed {
			log.Info(context.TODO(), fmt.Sprintf("all agents registered! starting stack creation. clusterId %s", clusterId))
			// clusterId, newStatus, newMessage, workflowId
			if err = clusterAccessor.UpdateClusterStatus(clusterId, domain.ClusterStatus_INSTALLING, "", ""); err != nil {
				log.Error(context.TODO(), "Failed to update cluster status err : ", err)
				continue
			}

			if cluster.IsStack {
				if _, err = apiClient.Post(fmt.Sprintf("organizations/%s/stacks/%s/install", cluster.OrganizationId, clusterId), nil); err != nil {
					log.Error(context.TODO(), err)
					continue
				}
			} else {
				if _, err = apiClient.Post("clusters/"+clusterId+"/install", nil); err != nil {
					log.Error(context.TODO(), err)
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
	_, err := apiClient.Get("auth/verify-token")
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

		log.Info(context.TODO(), out.User.Token)
		token = out.User.Token
	}

	return token
}
