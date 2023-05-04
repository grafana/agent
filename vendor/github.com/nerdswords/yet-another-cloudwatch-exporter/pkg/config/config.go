package config

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
)

type ScrapeConf struct {
	APIVersion      string             `yaml:"apiVersion"`
	StsRegion       string             `yaml:"sts-region"`
	Discovery       Discovery          `yaml:"discovery"`
	Static          []*Static          `yaml:"static"`
	CustomNamespace []*CustomNamespace `yaml:"customNamespace"`
}

type Discovery struct {
	ExportedTagsOnMetrics ExportedTagsOnMetrics `yaml:"exportedTagsOnMetrics"`
	Jobs                  []*Job                `yaml:"jobs"`
}

type ExportedTagsOnMetrics map[string][]string

type JobLevelMetricFields struct {
	Statistics             []string `yaml:"statistics"`
	Period                 int64    `yaml:"period"`
	Length                 int64    `yaml:"length"`
	Delay                  int64    `yaml:"delay"`
	NilToZero              *bool    `yaml:"nilToZero"`
	AddCloudwatchTimestamp *bool    `yaml:"addCloudwatchTimestamp"`
}

type Job struct {
	Regions                   []string    `yaml:"regions"`
	Type                      string      `yaml:"type"`
	Roles                     []Role      `yaml:"roles"`
	SearchTags                []model.Tag `yaml:"searchTags"`
	CustomTags                []model.Tag `yaml:"customTags"`
	DimensionNameRequirements []string    `yaml:"dimensionNameRequirements"`
	Metrics                   []*Metric   `yaml:"metrics"`
	RoundingPeriod            *int64      `yaml:"roundingPeriod"`
	JobLevelMetricFields      `yaml:",inline"`
}

type Static struct {
	Name       string      `yaml:"name"`
	Regions    []string    `yaml:"regions"`
	Roles      []Role      `yaml:"roles"`
	Namespace  string      `yaml:"namespace"`
	CustomTags []model.Tag `yaml:"customTags"`
	Dimensions []Dimension `yaml:"dimensions"`
	Metrics    []*Metric   `yaml:"metrics"`
}

type CustomNamespace struct {
	Regions                   []string    `yaml:"regions"`
	Name                      string      `yaml:"name"`
	Namespace                 string      `yaml:"namespace"`
	Roles                     []Role      `yaml:"roles"`
	Metrics                   []*Metric   `yaml:"metrics"`
	CustomTags                []model.Tag `yaml:"customTags"`
	DimensionNameRequirements []string    `yaml:"dimensionNameRequirements"`
	RoundingPeriod            *int64      `yaml:"roundingPeriod"`
	JobLevelMetricFields      `yaml:",inline"`
}

type Metric struct {
	Name                   string   `yaml:"name"`
	Statistics             []string `yaml:"statistics"`
	Period                 int64    `yaml:"period"`
	Length                 int64    `yaml:"length"`
	Delay                  int64    `yaml:"delay"`
	NilToZero              *bool    `yaml:"nilToZero"`
	AddCloudwatchTimestamp *bool    `yaml:"addCloudwatchTimestamp"`
}

type Dimension struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Role struct {
	RoleArn    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalId"`
}

func (r *Role) ValidateRole(roleIdx int, parent string) error {
	if r.RoleArn == "" && r.ExternalID != "" {
		return fmt.Errorf("Role [%d] in %v: RoleArn should not be empty", roleIdx, parent)
	}

	return nil
}

func (c *ScrapeConf) Load(file *string) error {
	yamlFile, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}

	for _, job := range c.Discovery.Jobs {
		if len(job.Roles) == 0 {
			job.Roles = []Role{{}} // use current IAM role
		}
	}

	for _, job := range c.CustomNamespace {
		if len(job.Roles) == 0 {
			job.Roles = []Role{{}} // use current IAM role
		}
	}

	for _, job := range c.Static {
		if len(job.Roles) == 0 {
			job.Roles = []Role{{}} // use current IAM role
		}
	}

	err = c.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapeConf) Validate() error {
	if c.Discovery.Jobs == nil && c.Static == nil && c.CustomNamespace == nil {
		return fmt.Errorf("At least 1 Discovery job, 1 Static or one CustomNamespace must be defined")
	}

	if c.Discovery.Jobs != nil {
		for idx, job := range c.Discovery.Jobs {
			err := job.validateDiscoveryJob(idx)
			if err != nil {
				return err
			}
		}
	}

	if c.CustomNamespace != nil {
		for idx, job := range c.CustomNamespace {
			err := job.validateCustomNamespaceJob(idx)
			if err != nil {
				return err
			}
		}
	}

	if c.Static != nil {
		for idx, job := range c.Static {
			err := job.validateStaticJob(idx)
			if err != nil {
				return err
			}
		}
	}
	if c.APIVersion != "" && c.APIVersion != "v1alpha1" {
		return fmt.Errorf("apiVersion line missing or version is unknown (%s)", c.APIVersion)
	}

	return nil
}

func (j *Job) validateDiscoveryJob(jobIdx int) error {
	if j.Type != "" {
		if SupportedServices.GetService(j.Type) == nil {
			return fmt.Errorf("Discovery job [%d]: Service is not in known list!: %s", jobIdx, j.Type)
		}
	} else {
		return fmt.Errorf("Discovery job [%d]: Type should not be empty", jobIdx)
	}
	parent := fmt.Sprintf("Discovery job [%s/%d]", j.Type, jobIdx)
	if len(j.Roles) > 0 {
		for roleIdx, role := range j.Roles {
			if err := role.ValidateRole(roleIdx, parent); err != nil {
				return err
			}
		}
	}
	if len(j.Regions) == 0 {
		return fmt.Errorf("Discovery job [%s/%d]: Regions should not be empty", j.Type, jobIdx)
	}
	if len(j.Metrics) == 0 {
		return fmt.Errorf("Discovery job [%s/%d]: Metrics should not be empty", j.Type, jobIdx)
	}
	for metricIdx, metric := range j.Metrics {
		err := metric.validateMetric(metricIdx, parent, &j.JobLevelMetricFields)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *CustomNamespace) validateCustomNamespaceJob(jobIdx int) error {
	if j.Name == "" {
		return fmt.Errorf("CustomNamespace job [%v]: Name should not be empty", jobIdx)
	}
	if j.Namespace == "" {
		return fmt.Errorf("CustomNamespace job [%v]: Namespace should not be empty", jobIdx)
	}
	parent := fmt.Sprintf("CustomNamespace job [%s/%d]", j.Namespace, jobIdx)
	if len(j.Roles) > 0 {
		for roleIdx, role := range j.Roles {
			if err := role.ValidateRole(roleIdx, parent); err != nil {
				return err
			}
		}
	}
	if j.Regions == nil || len(j.Regions) == 0 {
		return fmt.Errorf("CustomNamespace job [%s/%d]: Regions should not be empty", j.Name, jobIdx)
	}
	if len(j.Metrics) == 0 {
		return fmt.Errorf("CustomNamespace job [%s/%d]: Metrics should not be empty", j.Name, jobIdx)
	}
	for metricIdx, metric := range j.Metrics {
		err := metric.validateMetric(metricIdx, parent, &j.JobLevelMetricFields)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *Static) validateStaticJob(jobIdx int) error {
	if j.Name == "" {
		return fmt.Errorf("Static job [%v]: Name should not be empty", jobIdx)
	}
	if j.Namespace == "" {
		return fmt.Errorf("Static job [%s/%d]: Namespace should not be empty", j.Name, jobIdx)
	}
	parent := fmt.Sprintf("Static job [%s/%d]", j.Name, jobIdx)
	if len(j.Roles) > 0 {
		for roleIdx, role := range j.Roles {
			if err := role.ValidateRole(roleIdx, parent); err != nil {
				return err
			}
		}
	}
	if len(j.Regions) == 0 {
		return fmt.Errorf("Static job [%s/%d]: Regions should not be empty", j.Name, jobIdx)
	}
	for metricIdx, metric := range j.Metrics {
		err := metric.validateMetric(metricIdx, parent, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Metric) validateMetric(metricIdx int, parent string, discovery *JobLevelMetricFields) error {
	if m.Name == "" {
		return fmt.Errorf("Metric [%s/%d] in %v: Name should not be empty", m.Name, metricIdx, parent)
	}

	mStatistics := m.Statistics
	if len(mStatistics) == 0 && discovery != nil {
		if len(discovery.Statistics) > 0 {
			mStatistics = discovery.Statistics
		} else {
			return fmt.Errorf("Metric [%s/%d] in %v: Statistics should not be empty", m.Name, metricIdx, parent)
		}
	}

	mPeriod := m.Period
	if mPeriod == 0 && discovery != nil {
		if discovery.Period != 0 {
			mPeriod = discovery.Period
		} else {
			mPeriod = model.DefaultPeriodSeconds
		}
	}
	if mPeriod < 1 {
		return fmt.Errorf("Metric [%s/%d] in %v: Period value should be a positive integer", m.Name, metricIdx, parent)
	}
	mLength := m.Length
	if mLength == 0 && discovery != nil {
		if discovery.Length != 0 {
			mLength = discovery.Length
		} else {
			mLength = model.DefaultLengthSeconds
		}
	}

	mDelay := m.Delay
	if mDelay == 0 && discovery != nil {
		if discovery.Delay != 0 {
			mDelay = discovery.Delay
		} else {
			mDelay = model.DefaultDelaySeconds
		}
	}

	mNilToZero := m.NilToZero
	if mNilToZero == nil && discovery != nil {
		if discovery.NilToZero != nil {
			mNilToZero = discovery.NilToZero
		} else {
			mNilToZero = aws.Bool(false)
		}
	}

	mAddCloudwatchTimestamp := m.AddCloudwatchTimestamp
	if mAddCloudwatchTimestamp == nil && discovery != nil {
		if discovery.AddCloudwatchTimestamp != nil {
			mAddCloudwatchTimestamp = discovery.AddCloudwatchTimestamp
		} else {
			mAddCloudwatchTimestamp = aws.Bool(false)
		}
	}

	if mLength < mPeriod {
		log.Warningf(
			"Metric [%s/%d] in %v: length(%d) is smaller than period(%d). This can cause that the data requested is not ready and generate data gaps",
			m.Name, metricIdx, parent, mLength, mPeriod)
	}
	m.Length = mLength
	m.Period = mPeriod
	m.Delay = mDelay
	m.NilToZero = mNilToZero
	m.AddCloudwatchTimestamp = mAddCloudwatchTimestamp
	m.Statistics = mStatistics

	return nil
}
