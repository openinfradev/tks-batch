package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	systemNotification "github.com/openinfradev/tks-batch/internal/system-notification-rule"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const RULER_FILE_NAME = "ruler-user.yml"

type RuleAnnotation struct {
	CheckPoint               string `yaml:"CheckPoint"`
	Description              string `yaml:"description"`
	Discriminative           string `yaml:"discriminative"`
	Message                  string `yaml:"message"`
	Summary                  string `yaml:"summary"`
	AlertType                string `yaml:"alertType"`
	SystemNotificationRuleId string `yaml:"systemNotificationRuleId,omitempty"`
	PolicyName               string `yaml:"policyName,omitempty"`
	PolicyTemplateName       string `yaml:"policyTemplateName,omitempty"`
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
	Groups []RulerConfigGroup `yaml:"groups"`
}

func processSystemNotificationRule() error {
	rules, err := systemNotificationRuleAccessor.GetIncompletedRules()
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	log.Info(context.TODO(), "[processSystemNotificationRule] incompleted rules : ", len(rules))

	incompletedOrganizations := []string{}

	for _, rule := range rules {
		exist := false
		for _, organization := range incompletedOrganizations {
			if organization == rule.Organization.ID {
				exist = true
				break
			}
		}
		if !exist {
			incompletedOrganizations = append(incompletedOrganizations, rule.Organization.ID)
		}
	}

	for _, organizationId := range incompletedOrganizations {
		systemNotificationRules, err := systemNotificationRuleAccessor.GetRules(organizationId)
		if err != nil {
			log.Error(context.TODO(), err)
			continue
		}

		if len(rules) == 0 {
			continue
		}
		primaryClusterId := rules[0].Organization.PrimaryClusterId
		if primaryClusterId == "" {
			log.Error(context.TODO(), fmt.Sprintf("Invalid primary cluster for organization %s", organizationId))
			continue
		}

		log.Infof(context.TODO(), "imcompletedOrganizationId[%s] primaryClusterId[%s] rules[%d]", organizationId, primaryClusterId, len(rules))

		config := RulerConfig{}
		config.Groups = make([]RulerConfigGroup, 1)
		config.Groups[0] = RulerConfigGroup{
			Name: "tks",
		}
		config.Groups[0].Rules = make([]Rule, 0)
		for _, systemNotificationRule := range systemNotificationRules {
			rule := makeRuleForConfigMap(systemNotificationRule)
			config.Groups[0].Rules = append(config.Groups[0].Rules, rule)
		}

		err = applyRules(organizationId, primaryClusterId, config)
		if err != nil {
			log.Error(context.TODO(), fmt.Sprintf("Failed to apply rules. organizationId[%s] primaryClusterId[%s]", organizationId, primaryClusterId))
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

func makeRuleForConfigMap(systemNotificationRule systemNotification.SystemNotificationRule) (out Rule) {
	var parameters []domain.SystemNotificationParameter
	err := json.Unmarshal(systemNotificationRule.SystemNotificationCondition.Parameter, &parameters)
	if err != nil {
		log.Error(context.TODO(), err)
	}

	// expr
	expr := systemNotificationRule.SystemNotificationTemplate.MetricQuery
	if len(parameters) == 1 {
		expr = fmt.Sprintf("%s %s %s", expr, parameters[0].Operator, parameters[0].Value)
	} else {
		log.Error(context.TODO(), "Not support multiple parameters")
	}

	// metric paramters
	discriminative := ""
	for i, parameter := range systemNotificationRule.SystemNotificationTemplate.MetricParameters {
		if i == 0 {
			discriminative = parameter.Value
		} else {
			discriminative = discriminative + ", " + parameter.Value
		}
	}

	out = Rule{
		Alert: systemNotificationRule.Name,
		Expr:  expr,
		For:   systemNotificationRule.SystemNotificationCondition.Duration,
		Annotations: RuleAnnotation{
			CheckPoint:               replaceMetricParameter(systemNotificationRule.SystemNotificationTemplate.MetricParameters, systemNotificationRule.MessageActionProposal),
			Description:              replaceMetricParameter(systemNotificationRule.SystemNotificationTemplate.MetricParameters, systemNotificationRule.MessageContent),
			Message:                  replaceMetricParameter(systemNotificationRule.SystemNotificationTemplate.MetricParameters, systemNotificationRule.MessageTitle),
			Discriminative:           discriminative,
			AlertType:                systemNotificationRule.NotificationType,
			SystemNotificationRuleId: systemNotificationRule.ID.String(),
		},
		Labels: RuleLabels{
			Severity: systemNotificationRule.SystemNotificationCondition.Severity,
		},
	}

	if systemNotificationRule.NotificationType == "POLICY_NOTIFICATION" {
		out.Annotations.PolicyName = "{{$labels.name}}"
		out.Annotations.PolicyTemplateName = "{{$labels.kind}}"
	}

	return out
}

func applyRules(organizationId string, primaryClusterId string, rc RulerConfig) (err error) {
	clientset, err := kubernetes.GetClientFromClusterId(context.TODO(), primaryClusterId)
	if err != nil {
		log.Error(context.TODO(), err)
		return err
	}

	cm, err := clientset.CoreV1().ConfigMaps("lma").Get(context.TODO(), "thanos-ruler-configmap", metav1.GetOptions{})
	if err != nil {
		log.Error(context.TODO(), err)
		return err
	}

	var rulerConfig RulerConfig
	err = yaml.Unmarshal([]byte(cm.Data[RULER_FILE_NAME]), &rulerConfig)
	if err != nil {
		log.Error(context.TODO(), err)
		return err
	}

	if rc.Groups == nil || len(rc.Groups) == 0 {
		return fmt.Errorf("empty rc.Groups")
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
		return err
	}
	cm.Data[RULER_FILE_NAME] = string(b)

	_, err = clientset.CoreV1().ConfigMaps("lma").Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		log.Error(context.TODO(), err)
		return err
	}

	// restart thanos-ruler
	// thanos-ruler reload 방식으로 변경했으나, 혹시 몰라 일단 코드는 주석처리해둠
	/*
		deletePolicy := metav1.DeletePropagationForeground
		err = clientset.CoreV1().Pods("lma").Delete(context.TODO(), "thanos-ruler-0", metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if err != nil {
			log.Error(context.TODO(), err)
			return err
		}
	*/

	// update status
	err = systemNotificationRuleAccessor.UpdateSystemNotificationRuleStatus(organizationId, domain.SystemNotificationRuleStatus_APPLIED)
	if err != nil {
		log.Error(context.TODO(), err)
		return err
	}
	return nil
}
