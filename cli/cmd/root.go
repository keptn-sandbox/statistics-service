/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
	"time"

	stats "github.com/keptn-sandbox/statistics-service/operations"
)

type statisticsOutput struct {
	from                 time.Time
	to                   time.Time
	overallStatistics    statistics
	perProjectStatistics map[string]*statistics
}

type statistics struct {
	name                   string
	automationUnits        int
	keptnServiceExecutions map[string]*keptnServiceExecution
	triggers               int
	triggersByType         map[string]*triggerExecution
	subStatistics          map[string]*statistics
}

type keptnServiceExecution struct {
	eventTypeCount map[string]int
}

type triggerExecution struct {
	name      string
	count     int
	eventType string
}

var cfgFile string

var (
	folder             string
	period             string
	granularity        string
	granularityArr     []string
	includeEvents      string
	includeEventsArr   []string
	includeServices    string
	includeServicesArr []string
	excludeProjects    string
	excludeProjectsArr []string
	includeTriggers    string
	includeTriggersArr []string
	export             string
	separator          string
)

var allowedPeriods = []string{"separated", "aggregated"}
var allowedGranularities = []string{"overall", "project", "service"}
var allowedExport = []string{"json", "csv"}
var allowedSeparator = []string{",", ";"}

const separatedPeriod = "separated"
const aggregatedPeriod = "aggregated"

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "keptn-usage-stats",
	Short: "Generates an overview of Keptn usage statistics",
	Long: `Generates an overview of Keptn usage statistics, based on a set of input files provided to the command. Example:

keptn-usage-stats
   --folder=./usage-statistics-xyz 
   --period=separated
   --granularity=overall,project 
   --includeEvents=deployment-finished,tests-finished,evaluation-done 
   --includeServices=all`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkPeriod(); err != nil {
			er(err.Error())
		}
		if err := checkGranularity(); err != nil {
			er(err.Error())
		}
		if err := checkIncludeEvents(); err != nil {
			er(err.Error())
		}
		if err := checkIncludeServices(); err != nil {
			er(err.Error())
		}
		if err := checkExcludeProjects(); err != nil {
			er(err.Error())
		}
		if err := checkIncludeTriggers(); err != nil {
			er(err.Error())
		}
		if err := checkExport(); err != nil {
			er(err.Error())
		}
		if err := checkSeparator(); err != nil {
			er(err.Error())
		}

		statisticsFiles := map[string]*stats.GetStatisticsResponse{}

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			er(fmt.Errorf("the provided folder %s does not exist", folder))
		}

		// read all files within the directory
		fileInfos, err := ioutil.ReadDir(folder)
		if err != nil {
			er(err.Error())
		}

		for _, file := range fileInfos {
			if statisticsStruct, err := readStatisticsFile(folder + "/" + file.Name()); err == nil && statisticsStruct != nil {
				statisticsFiles[file.Name()] = statisticsStruct
			} else if err != nil {
				fmt.Println("not adding file " + file.Name() + ": " + err.Error())
			}
		}

		if len(statisticsFiles) == 0 {
			fmt.Println(fmt.Sprintf("folder %s does not contain any usable files", folder))
			os.Exit(1)
		}

		var statisticsArr []*statisticsOutput
		if period == separatedPeriod {
			statisticsArr = createSeparatedStatistics(statisticsFiles)
		} else {
			statisticsArr = createAggregatedStatistics(statisticsFiles)
		}

		for _, s := range statisticsArr {
			printStats(s)
		}
	},
}

func printStats(s *statisticsOutput) {
	fmt.Println("---")
	fmt.Println("Timeframe: " + s.from.String() + " - " + s.to.String())
	fmt.Println("---")
	fmt.Println("")
	printSubStats(&s.overallStatistics)

	if isProjectGranularity() {
		for _, projectStat := range s.perProjectStatistics {
			printSubStats(projectStat)
			if isServiceGranularity() {
				for _, svcStat := range projectStat.subStatistics {
					printSubStats(svcStat)
				}
			}
		}
	}
}

func printSubStats(s *statistics) {
	fmt.Println(s.name)
	fmt.Println(fmt.Sprintf("- Executions: %d", s.automationUnits))
	fmt.Println("-------------------------------------------------")
	for keptnService, executions := range s.keptnServiceExecutions {
		for eventType, execution := range executions.eventTypeCount {
			fmt.Println(fmt.Sprintf("- %s: \t\t %d \t %s", keptnService, execution, eventType))
		}
	}
	fmt.Println("")
}

func createAggregatedStatistics(statisticsFiles map[string]*stats.GetStatisticsResponse) []*statisticsOutput {
	statsOutput := &statisticsOutput{
		overallStatistics: statistics{
			name:                   "Overall: Keptn",
			automationUnits:        0,
			keptnServiceExecutions: map[string]*keptnServiceExecution{},
			triggers:               0,
			triggersByType:         map[string]*triggerExecution{},
			subStatistics:          map[string]*statistics{},
		},
		perProjectStatistics: map[string]*statistics{},
	}

	to, from := getTimeFrame(statisticsFiles)
	statsOutput.from = from
	statsOutput.to = to

	for _, stats := range statisticsFiles {
		mergeStatisticsResponseIntoStatisticsOutput(stats, statsOutput)
	}

	return []*statisticsOutput{statsOutput}
}

func createSeparatedStatistics(statisticsFiles map[string]*stats.GetStatisticsResponse) []*statisticsOutput {
	result := []*statisticsOutput{}

	for fileName, stats := range statisticsFiles {
		newSeparateOutput := &statisticsOutput{
			overallStatistics: statistics{
				name:                   "Overall: " + fileName + " > Keptn",
				automationUnits:        0,
				keptnServiceExecutions: map[string]*keptnServiceExecution{},
				triggers:               0,
				triggersByType:         map[string]*triggerExecution{},
				subStatistics:          map[string]*statistics{},
			},
			perProjectStatistics: map[string]*statistics{},
			from:                 stats.From,
			to:                   stats.To,
		}
		mergeStatisticsResponseIntoStatisticsOutput(stats, newSeparateOutput)
		result = append(result, newSeparateOutput)
	}
	return result
}

func mergeStatisticsResponseIntoStatisticsOutput(stats *stats.GetStatisticsResponse, statsOutput *statisticsOutput) {
	for _, project := range stats.Projects {
		if len(excludeProjectsArr) > 0 && contains(excludeProjectsArr, project.Name) {
			continue
		}
		if isProjectGranularity() {
			if statsOutput.perProjectStatistics[project.Name] == nil {
				statsOutput.perProjectStatistics[project.Name] = &statistics{
					name:                   "Project: Keptn > " + project.Name,
					automationUnits:        0,
					keptnServiceExecutions: map[string]*keptnServiceExecution{},
					triggers:               0,
					triggersByType:         map[string]*triggerExecution{},
					subStatistics:          map[string]*statistics{},
				}
			}
		}
		for _, svc := range project.Services {
			if len(includeServicesArr) > 0 && !contains(includeServicesArr, svc.Name) {
				continue
			}
			if isServiceGranularity() {
				if statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name] == nil {
					statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name] = &statistics{
						name:                   "Service: Keptn > " + project.Name + " > " + svc.Name,
						automationUnits:        0,
						keptnServiceExecutions: map[string]*keptnServiceExecution{},
						triggers:               0,
						triggersByType:         map[string]*triggerExecution{},
						subStatistics:          nil,
					}
				}
			}
			for _, execution := range svc.KeptnServiceExecutions {
				for _, eventTypeExecution := range execution.Executions {
					if len(includeEventsArr) > 0 && !contains(includeEventsArr, eventTypeExecution.Type) {
						continue
					}
					statsOutput.overallStatistics.automationUnits = statsOutput.overallStatistics.automationUnits + eventTypeExecution.Count

					if statsOutput.overallStatistics.keptnServiceExecutions[execution.Name] == nil {
						statsOutput.overallStatistics.keptnServiceExecutions[execution.Name] = &keptnServiceExecution{
							eventTypeCount: map[string]int{},
						}
					}
					statsOutput.overallStatistics.keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] =
						statsOutput.overallStatistics.keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] + eventTypeExecution.Count

					if isProjectGranularity() {
						statsOutput.perProjectStatistics[project.Name].automationUnits = statsOutput.perProjectStatistics[project.Name].automationUnits + eventTypeExecution.Count
						if statsOutput.perProjectStatistics[project.Name].keptnServiceExecutions[execution.Name] == nil {
							statsOutput.perProjectStatistics[project.Name].keptnServiceExecutions[execution.Name] = &keptnServiceExecution{
								eventTypeCount: map[string]int{},
							}
						}
						statsOutput.perProjectStatistics[project.Name].keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] =
							statsOutput.perProjectStatistics[project.Name].keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] + eventTypeExecution.Count

						if isServiceGranularity() {
							statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].automationUnits =
								statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].automationUnits + eventTypeExecution.Count

							if statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].keptnServiceExecutions[execution.Name] == nil {
								statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].keptnServiceExecutions[execution.Name] = &keptnServiceExecution{
									eventTypeCount: map[string]int{},
								}
							}
							statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] =
								statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name].keptnServiceExecutions[execution.Name].eventTypeCount[eventTypeExecution.Type] + eventTypeExecution.Count
						}
					}
				}
			}
		}
	}
}

func isProjectGranularity() bool {
	return granularity == "project" || granularity == "service"
}

func isServiceGranularity() bool {
	return granularity == "service"
}

func getTimeFrame(statistics map[string]*stats.GetStatisticsResponse) (time.Time, time.Time) {
	fromTime := time.Time{}
	toTime := time.Time{}

	for _, stat := range statistics {
		if fromTime.Equal(time.Time{}) || fromTime.After(stat.From) {
			fromTime = stat.From
		}
		if toTime.Equal(time.Time{}) || toTime.Before(stat.To) {
			toTime = stat.To
		}
	}
	return fromTime, toTime
}

func contains(arr []string, val string) bool {
	for _, s := range arr {
		if strings.ToLower(s) == strings.ToLower(val) {
			return true
		}
	}
	return false
}

func readStatisticsFile(fileName string) (*stats.GetStatisticsResponse, error) {
	stat := &stats.GetStatisticsResponse{}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %s", fileName, err.Error())
	}

	err = json.Unmarshal(bytes, stat)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file %s: %s", fileName, err.Error())
	}
	return stat, nil
}

func checkPeriod() error {
	period = strings.TrimSpace(strings.ToLower(period))
	if !checkAllowedValues(period, allowedPeriods) {
		return fmt.Errorf("unsupported value '%s' for period. allowed values are: %v", period, allowedPeriods)
	}
	return nil
}

func checkExport() error {
	export = strings.TrimSpace(strings.ToLower(export))
	if !checkAllowedValues(export, allowedExport) {
		return fmt.Errorf("unsupported value '%s' for export. allowed values are: %v", export, allowedExport)
	}
	return nil
}

func checkSeparator() error {
	separator = strings.TrimSpace(strings.ToLower(separator))
	if !checkAllowedValues(separator, allowedSeparator) {
		return fmt.Errorf("unsupported value '%s' for separator. allowed values are: %v", separator, allowedSeparator)
	}
	return nil
}

func checkGranularity() error {
	granularityArr = strings.Split(strings.TrimSpace(strings.ToLower(granularity)), ",")
	for _, gr := range granularityArr {
		if !checkAllowedValues(gr, allowedGranularities) {
			return fmt.Errorf("unsupported value '%s' for granularity. allowed values are: %v", granularity, allowedGranularities)
		}
	}
	return nil
}

func checkIncludeEvents() error {
	includeEvents = strings.TrimSpace(strings.ToLower(includeEvents))
	if includeEvents == "all" {
		includeEventsArr = []string{}
		return nil
	}
	includeEventsArr = strings.Split(includeEvents, ",")
	return nil
}

func checkIncludeServices() error {
	includeServices = strings.TrimSpace(strings.ToLower(includeServices))
	if includeServices == "all" {
		includeServicesArr = []string{}
		return nil
	}
	includeServicesArr = strings.Split(includeServices, ",")
	return nil
}

func checkIncludeTriggers() error {
	includeTriggers = strings.TrimSpace(strings.ToLower(includeTriggers))
	if includeTriggers == "all" {
		includeTriggersArr = []string{}
		return nil
	}
	includeTriggersArr = strings.Split(includeTriggers, ",")
	return nil
}

func checkExcludeProjects() error {
	excludeProjects = strings.TrimSpace(strings.ToLower(excludeProjects))
	excludeProjectsArr = strings.Split(excludeProjects, ",")
	return nil
}

func checkAllowedValues(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&folder, "folder", "f", "", "The folder containing the JSON files exported from the statistics-service")
	rootCmd.PersistentFlags().StringVarP(&period, "period", "p", "separated", "The period under consideration, one option of: [separated, aggregated]")
	rootCmd.PersistentFlags().StringVarP(&granularity, "granularity", "g", "overall", "The level of details, list of [overall, project, service], default is 'overall'")
	rootCmd.PersistentFlags().StringVarP(&includeEvents, "includeEvents", "", "all", "List of events that define an automation unit, default is 'all'")
	rootCmd.PersistentFlags().StringVarP(&includeServices, "includeServices", "", "all", "List of services that define an automation unit, default is 'all'")
	rootCmd.PersistentFlags().StringVarP(&excludeProjects, "excludeProjects", "", "", "List of project names that are excluded from the summary")
	rootCmd.PersistentFlags().StringVarP(&includeTriggers, "includeTriggers", "", "all", "list of sequence triggers: [configuration-change, problem.open, evaluation-started]")
	rootCmd.PersistentFlags().StringVarP(&export, "export", "", "json", "The format to export the statistics, supported are [json, csv]")
	rootCmd.PersistentFlags().StringVarP(&separator, "separator", "", ",", "The separator used for the CSV exporter, allowed values are ',' or ';'")
	cobra.OnInitialize(initConfig)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
