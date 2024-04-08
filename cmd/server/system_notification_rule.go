package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	systemNotification "github.com/openinfradev/tks-batch/internal/system-notification-rule"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RuleAnnotation struct {
	CheckPoint     string `yaml:"CheckPoint"`
	Description    string `yaml:"description"`
	Discriminative string `yaml:"discriminative"`
	Message        string `yaml:"message"`
	Summary        string `yaml:"summary"`
}

type RuleLabels struct {
	Severity string `yaml:"severity"`
}

type Rule struct {
	Alert       string         `yaml:"alert"`
	Expr        string         `yaml:"expr"`
	For         string         `yaml:"for"`
	Labels      RuleLabels     `yaml:"labels"`
	Annotations RuleAnnotation `yaml:"annotations"`
}

type RulerConfigGroup struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

type RulerConfig struct {
	Groups           []RulerConfigGroup `yaml:"groups"`
	PrimaryClusterId string             `yaml:"-"`
}

var rulerConfigOrganizations map[string]RulerConfig

func processSystemNotificationRule() error {
	rules, err := systemNotificationRuleAccessor.GetIncompletedRules()
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	log.Info(context.TODO(), "incompleted rules : ", len(rules))

	rulerConfigOrganizations = make(map[string]RulerConfig)

	for _, rule := range rules {
		config, exists := rulerConfigOrganizations[rule.OrganizationId]
		if !exists {
			config = RulerConfig{
				PrimaryClusterId: rule.Organization.PrimaryClusterId,
			}
			config.Groups = make([]RulerConfigGroup, 1)
			config.Groups[0] = RulerConfigGroup{
				Name: "tks",
			}
			config.Groups[0].Rules = make([]Rule, 0)

			rulerConfigOrganizations[rule.OrganizationId] = config
		}

		var parameters []domain.SystemNotificationParameter
		err = json.Unmarshal(rule.SystemNotificationCondition.Parameter, &parameters)
		if err != nil {
			log.Error(context.TODO(), err)
		}

		// expr
		expr := rule.SystemNotificationTemplate.MetricQuery
		if len(parameters) == 1 {
			expr = fmt.Sprintf("%s %s %s", expr, parameters[0].Operator, parameters[0].Value)
		} else {
			log.Error(context.TODO(), "Not support multiple parameters")
		}

		// metric paramters
		discriminative := ""
		for i, parameter := range rule.SystemNotificationTemplate.MetricParameters {
			if i == 0 {
				discriminative = parameter.Value
			} else {
				discriminative = discriminative + ", " + parameter.Value
			}
		}

		rulerConfigOrganizations[rule.OrganizationId].Groups[0].Rules = append(rulerConfigOrganizations[rule.OrganizationId].Groups[0].Rules,
			Rule{
				Alert: rule.Name,
				Expr:  expr,
				For:   rule.SystemNotificationCondition.Duration,
				Annotations: RuleAnnotation{
					CheckPoint:     replaceMetricParameter(rule.SystemNotificationTemplate.MetricParameters, rule.MessageActionProposal),
					Description:    replaceMetricParameter(rule.SystemNotificationTemplate.MetricParameters, rule.MessageContent),
					Message:        replaceMetricParameter(rule.SystemNotificationTemplate.MetricParameters, rule.MessageTitle),
					Discriminative: discriminative,
				},
				Labels: RuleLabels{
					Severity: rule.SystemNotificationCondition.Severity,
				},
			},
		)
	}

	for organizationId, rc := range rulerConfigOrganizations {
		log.Infof(context.TODO(), "organizationId [%s] primaryClusterId [%s] ", organizationId, rc.PrimaryClusterId)

		clientset, err := kubernetes.GetClientFromClusterId(context.TODO(), rc.PrimaryClusterId)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		cm, err := clientset.CoreV1().ConfigMaps("lma").Get(context.TODO(), "thanos-ruler-configmap", metav1.GetOptions{})
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		var rulerConfig RulerConfig
		err = yaml.Unmarshal([]byte(cm.Data["ruler.yml"]), &rulerConfig)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}
		rulerConfig.Groups = rc.Groups

		/*
			outYaml, err := yaml.Marshal(rulerConfig)
			if err != nil {
				log.Error(context.TODO(), err)
				continue
			}
			log.Info(context.TODO(), string(outYaml))
		*/

		b, err := yaml.Marshal(rulerConfig)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}
		cm.Data["ruler.yml"] = string(b)

		_, err = clientset.CoreV1().ConfigMaps("lma").Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		// restart thanos-ruler
		deletePolicy := metav1.DeletePropagationForeground
		err = clientset.CoreV1().Pods("lma").Delete(context.TODO(), "thanos-ruler-0", metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		// update status
		var organizationRuleIds []uuid.UUID
		for _, r := range rules {
			if r.OrganizationId == organizationId {
				organizationRuleIds = append(organizationRuleIds, r.ID)
			}
		}
		err = systemNotificationRuleAccessor.UpdateSystemNotificationRuleStatus(organizationRuleIds, domain.SystemNotificationRuleStatus_APPLYED)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}
	}

	return nil
}

func replaceMetricParameter(l []systemNotification.SystemNotificationMetricParameter, s string) (out string) {
	for _, v := range l {
		s = strings.Replace(s, "<<"+v.Key+">>", "{{"+v.Value+"}}", -1)
	}
	return s
}

/*
func modelToYaml(in any) string {
	a, _ := yaml.Marshal(in)
	n := len(a)        //Find the length of the byte array
	s := string(a[:n]) //convert to string
	return s
}
*/
