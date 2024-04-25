package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const LAST_UPDATED_MIN = 1

func processReloadThanosRules() error {
	organizationIds, err := systemNotificationRuleAccessor.GetRecentlyUpdatedOrganizations(LAST_UPDATED_MIN)
	if err != nil {
		return err
	}
	if len(organizationIds) == 0 {
		return nil
	}
	log.Info(context.TODO(), "[processReloadThanosRules] new updated organizationIds : ", organizationIds)

	for _, organizationId := range organizationIds {
		organization, err := organizationAccessor.Get(organizationId)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		url, err := GetThanosRulerUrl(organization.PrimaryClusterId)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		if err = Reload(url); err != nil {
			log.Error(context.TODO(), err)
			continue
		}
	}

	return nil
}

func GetThanosRulerUrl(primaryClusterId string) (url string, err error) {
	clientset_admin, err := kubernetes.GetClientAdminCluster(context.TODO())
	if err != nil {
		return url, errors.Wrap(err, "Failed to get client set for user cluster")
	}

	secrets, err := clientset_admin.CoreV1().Secrets(primaryClusterId).Get(context.TODO(), "tks-endpoint-secret", metav1.GetOptions{})
	if err != nil {
		log.Info(context.TODO(), "cannot found tks-endpoint-secret. so use LoadBalancer...")

		clientset_user, err := kubernetes.GetClientFromClusterId(context.TODO(), primaryClusterId)
		if err != nil {
			return url, errors.Wrap(err, "Failed to get client set for user cluster")
		}

		service, err := clientset_user.CoreV1().Services("lma").Get(context.TODO(), "thanos-ruler", metav1.GetOptions{})
		if err != nil {
			return url, errors.Wrap(err, "Failed to get services.")
		}

		// LoadBalaner 일경우, aws address 형태의 경우만 가정한다.
		if service.Spec.Type != "LoadBalancer" {
			return url, fmt.Errorf("Service type is not LoadBalancer. [%s] ", service.Spec.Type)
		}

		lbs := service.Status.LoadBalancer.Ingress
		ports := service.Spec.Ports
		if len(lbs) > 0 && len(ports) > 0 {
			url = ports[0].TargetPort.StrVal + "://" + lbs[0].Hostname + ":" + strconv.Itoa(int(ports[0].Port))
		}
	} else {
		url = "http://" + string(secrets.Data["thanos-ruler"])
	}
	return url, nil
}

func Reload(thanosRulerUrl string) (err error) {
	reqUrl := thanosRulerUrl + "/-/reload"

	log.Info(context.TODO(), "url : ", reqUrl)
	resp, err := http.Post(reqUrl, "text/plain", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
