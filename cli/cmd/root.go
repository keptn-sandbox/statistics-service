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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	stats "github.com/keptn-sandbox/statistics-service/statistics-service/operations"
)

type csvColumn struct {
	key         string
	displayName string
	values      map[int]string
}

type exportedStatisticsService struct {
	Name       string `json:"name"`
	Executions int    `json:"Executions"`
	EventType  string `json:"eventType"`
}

type exportedStatisticsSummary struct {
	Granularity       string                      `json:"granularity"`
	Executions        int                         `json:"executions"`
	ServiceExecutions []exportedStatisticsService `json:"serviceExecutions"`
	Projects          []exportedStatisticsSummary `json:"projects,omitempty"`
	Services          []exportedStatisticsSummary `json:"services,omitempty"`
}

type exportedStatisticsOutput struct {
	Timeframe string                    `json:"timeframe"`
	Summary   exportedStatisticsSummary `json:"summary"`
}

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
	outputFile         string
)

var allowedPeriods = []string{"separated", "aggregated"}
var allowedGranularities = []string{"overall", "project", "service"}
var allowedExport = []string{"json", "csv"}
var allowedSeparator = []string{",", ";"}

const separatedPeriod = "separated"
const aggregatedPeriod = "aggregated"

const exportJSON = "json"
const exportCSV = "csv"

const separatorComma = ","
const separatorSemicolon = ";"

var csvColumns = []csvColumn{}
var csvColumnMapping = map[string]int{}
var columnIndex = 0

var csvRows = [][]string{}

const overallKeptnCSVColKey = "overall:keptn"
const overallKeptnTimeframeCSVColKey = "timeframe"

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
		var mergedStatistics []*statisticsOutput

		for range statisticsFiles {
			csvRows = append(csvRows, []string{})
		}
		if period == separatedPeriod {
			// get the merged statistics as well to know which columns we are going to have in the CSV file
			mergedStatistics = createAggregatedStatistics(statisticsFiles)
			createCSVColumns(mergedStatistics[0])

			statisticsArr = createSeparatedStatistics(statisticsFiles)
			createCSVRows(statisticsArr)
		} else {
			statisticsArr = createAggregatedStatistics(statisticsFiles)
			createCSVColumns(statisticsArr[0])
			createCSVRows(statisticsArr)
		}

		// print the statistics
		for _, s := range statisticsArr {
			printStats(s)
		}

		if export == exportCSV {
			exportToCSVMergeAllFiles(statisticsArr)
			return
		}

		// export the statistics to .json or .csv
		for index, s := range statisticsArr {
			exportToJSON(s, index)
		}
	},
}

func createCSVRows(stats []*statisticsOutput) {
	for index, s := range stats {

		csvRows[index][csvColumnMapping[overallKeptnCSVColKey]] = fmt.Sprintf("%d", s.overallStatistics.automationUnits)

		csvRows[index][csvColumnMapping[overallKeptnTimeframeCSVColKey]] = fmt.Sprintf("%s - %s", s.from.String(), s.to.String())

		for keptnServiceName, execution := range s.overallStatistics.keptnServiceExecutions {
			for eventType, count := range execution.eventTypeCount {
				overallKeptnExecutionKey, _ := getOverallKeptnExecutionColNames(keptnServiceName, eventType)
				csvRows[index][csvColumnMapping[overallKeptnExecutionKey]] = fmt.Sprintf("%d", count)
			}
		}

		for projectName, projectStats := range s.perProjectStatistics {
			projectOverallColumnKey, _ := getProjectColNames(projectName)

			csvRows[index][csvColumnMapping[projectOverallColumnKey]] = fmt.Sprintf("%d", projectStats.automationUnits)

			for keptnServiceName, execution := range projectStats.keptnServiceExecutions {
				for eventType, count := range execution.eventTypeCount {
					projectKeptnExecutionKey, _ := getProjectKeptnExecutionColNames(projectName, keptnServiceName, eventType)
					csvRows[index][csvColumnMapping[projectKeptnExecutionKey]] = fmt.Sprintf("%d", count)
				}
			}

			for svcName, svcStats := range projectStats.subStatistics {
				serviceOverallColumnKey, _ := getServiceColNames(projectName, svcName)

				csvRows[index][csvColumnMapping[serviceOverallColumnKey]] = fmt.Sprintf("%d", svcStats.automationUnits)

				for keptnServiceName, execution := range svcStats.keptnServiceExecutions {
					for eventType, count := range execution.eventTypeCount {
						serviceKeptnExecutionKey, _ := getSvcExecutionColNames(projectName, svcName, keptnServiceName, eventType)
						csvRows[index][csvColumnMapping[serviceKeptnExecutionKey]] = fmt.Sprintf("%d", count)
					}
				}
			}
		}
	}
}

func createCSVColumns(s *statisticsOutput) {
	addColumn(overallKeptnTimeframeCSVColKey, "Timeframe")
	addColumn(overallKeptnCSVColKey, "Overall: Keptn")

	for keptnServiceName, execution := range s.overallStatistics.keptnServiceExecutions {
		for eventType, _ := range execution.eventTypeCount {
			overallKeptnExecutionKey, overallKeptnExecutionDisplayName := getOverallKeptnExecutionColNames(keptnServiceName, eventType)
			addColumn(overallKeptnExecutionKey, overallKeptnExecutionDisplayName)
		}
	}

	for projectName, projectStats := range s.perProjectStatistics {
		projectOverallColumnKey, projectOverallDisplayName := getProjectColNames(projectName)

		addColumn(projectOverallColumnKey, projectOverallDisplayName)

		for keptnServiceName, execution := range projectStats.keptnServiceExecutions {
			for eventType, _ := range execution.eventTypeCount {
				projectKeptnExecutionKey, projectKeptnExecutionDisplayName := getProjectKeptnExecutionColNames(projectName, keptnServiceName, eventType)
				addColumn(projectKeptnExecutionKey, projectKeptnExecutionDisplayName)
			}
		}

		for svcName, svcStats := range projectStats.subStatistics {
			serviceOverallColumnKey, serviceOverallDisplayName := getServiceColNames(projectName, svcName)

			addColumn(serviceOverallColumnKey, serviceOverallDisplayName)

			for keptnServiceName, execution := range svcStats.keptnServiceExecutions {
				for eventType, _ := range execution.eventTypeCount {
					serviceKeptnExecutionKey, serviceKeptnExecutionDisplayName := getSvcExecutionColNames(projectName, svcName, keptnServiceName, eventType)
					addColumn(serviceKeptnExecutionKey, serviceKeptnExecutionDisplayName)
				}
			}
		}
	}
}

func getSvcExecutionColNames(projectName string, svcName string, keptnServiceName string, eventType string) (string, string) {
	serviceKeptnExecutionKey := "service: " + projectName + ":" + svcName + ":" + keptnServiceName + ":" + eventType
	serviceKeptnExecutionDisplayName := keptnServiceName + " (" + eventType + ")"
	return serviceKeptnExecutionKey, serviceKeptnExecutionDisplayName
}

func getServiceColNames(projectName string, svcName string) (string, string) {
	serviceOverallColumnKey := "Service: Keptn > " + projectName + " > " + svcName
	serviceOverallDisplayName := "Service: Keptn > " + projectName + " > " + svcName
	return serviceOverallColumnKey, serviceOverallDisplayName
}

func getProjectKeptnExecutionColNames(projectName string, keptnServiceName string, eventType string) (string, string) {
	projectKeptnExecutionKey := "project: " + projectName + ":" + keptnServiceName + ":" + eventType
	projectKeptnExecutionDisplayName := keptnServiceName + " (" + eventType + ")"
	return projectKeptnExecutionKey, projectKeptnExecutionDisplayName
}

func getProjectColNames(projectName string) (string, string) {
	projectOverallColumnKey := "Project: Keptn > " + projectName
	projectOverallDisplayName := "Project: Keptn > " + projectName
	return projectOverallColumnKey, projectOverallDisplayName
}

func getOverallKeptnExecutionColNames(keptnServiceName string, eventType string) (string, string) {
	overallKeptnExecutionKey := "overall: " + keptnServiceName + ":" + eventType
	overallKeptnExecutionDisplayName := keptnServiceName + " (" + eventType + ")"
	return overallKeptnExecutionKey, overallKeptnExecutionDisplayName
}

func exportToCSVMergeAllFiles(s []*statisticsOutput) {
	columns := map[string]*csvColumn{}

	columns["timeframe"] = &csvColumn{
		key:         "timeframe",
		displayName: "Timeframe",
		values:      map[int]string{},
	}

	for index, stats := range s {
		columns = generateSubStatsForMergedCSV(index, &stats.overallStatistics, "overall", columns)
		columns["timeframe"].values[index] = stats.from.String() + " - " + stats.to.String()

		if isProjectGranularity() {
			for _, projectStat := range stats.perProjectStatistics {
				columns = generateSubStatsForMergedCSV(index, projectStat, projectStat.name, columns)
				if isServiceGranularity() {
					for _, svcStat := range projectStat.subStatistics {
						columns = generateSubStatsForMergedCSV(index, svcStat, projectStat.name+"-"+svcStat.name, columns)
					}
				}
			}
		}
	}

	headers := []string{}

	for _, column := range csvColumns {
		headers = append(headers, column.displayName)
	}
	headersLine := strings.Join(headers, separator)
	fileContent := fmt.Sprintf("%s\n", headersLine)

	for _, row := range csvRows {
		rowLine := strings.Join(row, separator)
		fileContent = fileContent + fmt.Sprintf("%s\n", rowLine)
	}

	fileName := outputFile

	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	} else if strings.HasSuffix(fileName, ".json") {
		fileName = strings.TrimSuffix(fileName, ".json")
		fileName = fileName + ".csv"
	}

	writeFile(fileName, fileContent)
}

func generateSubStatsForMergedCSV(index int, s *statistics, columnKey string, columns map[string]*csvColumn) map[string]*csvColumn {
	if columns[columnKey] == nil {
		columns[columnKey] = &csvColumn{
			key:         columnKey,
			displayName: s.name,
			values:      map[int]string{},
		}
	}

	columns[columnKey].values[index] = fmt.Sprintf("%d", s.automationUnits)

	for keptnServiceName, keptnServiceExecution := range s.keptnServiceExecutions {
		for eventType, executions := range keptnServiceExecution.eventTypeCount {
			executionColumnKey := columnKey + "-executions-" + keptnServiceName + "-" + eventType
			if columns[executionColumnKey] == nil {
				columns[executionColumnKey] = &csvColumn{
					key:         executionColumnKey,
					displayName: fmt.Sprintf("%s (%s)", keptnServiceName, eventType),
					values:      map[int]string{},
				}
			}
			columns[executionColumnKey].values[index] = fmt.Sprintf("%d", executions)
		}
	}

	return columns
}

func exportToCSV(s *statisticsOutput, index int) {
	headers := []string{}
	values := []string{}

	headers = append(headers, "Timeframe")
	values = append(values, fmt.Sprintf("%s - %s", s.from.String(), s.to.String()))

	headers, values = generateSubStatsCSV(&s.overallStatistics, headers, values)

	if isProjectGranularity() {
		for _, projectStat := range s.perProjectStatistics {
			headers, values = generateSubStatsCSV(projectStat, headers, values)
			if isServiceGranularity() {
				for _, svcStat := range projectStat.subStatistics {
					headers, values = generateSubStatsCSV(svcStat, headers, values)
				}
			}
		}
	}
	headersLine := strings.Join(headers, separator)
	valuesLine := strings.Join(values, separator)

	fileName := getIndexedFileName(outputFile, index)

	if !strings.HasSuffix(fileName, ".csv") {
		fileName = fileName + ".csv"
	} else if strings.HasSuffix(fileName, ".json") {
		fileName = strings.TrimSuffix(fileName, ".json")
		fileName = fileName + ".csv"
	}

	fileContent := fmt.Sprintf("%s\n%s", headersLine, valuesLine)
	writeFile(fileName, fileContent)
}

func getIndexedFileName(outputFile string, index int) string {
	index = index + 1

	lastIndex := strings.LastIndex(outputFile, ".")
	if lastIndex == -1 {
		outputFile = fmt.Sprintf("%s_%d", outputFile, index)
	} else {
		outputFile = outputFile[:lastIndex] + fmt.Sprintf("_%d", index) + outputFile[lastIndex+1:]
	}
	return outputFile
}

func writeFile(outputFile, fileContent string) {
	if _, err := os.Stat(outputFile); os.IsExist(err) {
		if err := os.Remove(outputFile); err != nil {
			fmt.Println(fmt.Sprintf("could not delete file %s: %s", outputFile, err.Error()))
		}
	}
	if err := ioutil.WriteFile(outputFile, []byte(fileContent), 0664); err != nil {
		fmt.Println(fmt.Sprintf("could not write file %s: %s", outputFile, err.Error()))
	}
}

func generateSubStatsCSV(s *statistics, headers, values []string) ([]string, []string) {
	headers = append(headers, s.name)
	values = append(values, fmt.Sprintf("%d", s.automationUnits))

	for keptnServiceName, keptnServiceExecution := range s.keptnServiceExecutions {
		for eventType, executions := range keptnServiceExecution.eventTypeCount {
			headers = append(headers, fmt.Sprintf("%s (%s)", keptnServiceName, eventType))
			values = append(values, fmt.Sprintf("%d", executions))
		}
	}

	return headers, values
}

func exportToJSON(s *statisticsOutput, index int) {

	result := exportedStatisticsOutput{
		Timeframe: fmt.Sprintf("%s - %s", s.from, s.to),
		Summary: exportedStatisticsSummary{
			Granularity:       s.overallStatistics.name,
			Executions:        s.overallStatistics.automationUnits,
			ServiceExecutions: []exportedStatisticsService{},
			Projects:          []exportedStatisticsSummary{},
			Services:          []exportedStatisticsSummary{},
		},
	}

	appendKeptnServiceExecutions(&s.overallStatistics, &result.Summary)

	if isProjectGranularity() {
		for _, project := range s.perProjectStatistics {
			newExportedProjectStats := exportedStatisticsSummary{
				Granularity:       project.name,
				Executions:        project.automationUnits,
				ServiceExecutions: []exportedStatisticsService{},
				Projects:          nil,
				Services:          []exportedStatisticsSummary{},
			}

			appendKeptnServiceExecutions(project, &newExportedProjectStats)

			if isServiceGranularity() {
				for _, svc := range project.subStatistics {
					newExportedServiceStats := exportedStatisticsSummary{
						Granularity:       svc.name,
						Executions:        svc.automationUnits,
						ServiceExecutions: []exportedStatisticsService{},
						Projects:          nil,
						Services:          []exportedStatisticsSummary{},
					}

					appendKeptnServiceExecutions(svc, &newExportedServiceStats)
					newExportedProjectStats.Services = append(newExportedProjectStats.Services, newExportedServiceStats)
				}
			}
			result.Summary.Projects = append(result.Summary.Projects, newExportedProjectStats)
		}
	}

	fileContent, err := json.MarshalIndent(&result, "", fmt.Sprintf("  "))
	if err != nil {
		fmt.Println("could not generate json file content: " + err.Error())
		os.Exit(1)
	}

	fileName := getIndexedFileName(outputFile, index)
	if !strings.HasSuffix(fileName, ".json") {
		fileName = fileName + ".json"
	} else if strings.HasSuffix(fileName, ".csv") {
		fileName = strings.TrimSuffix(fileName, ".csv")
		fileName = fileName + ".json"
	}

	writeFile(fileName, string(fileContent))
}

func appendKeptnServiceExecutions(s *statistics, summary *exportedStatisticsSummary) {
	for keptnServiceName, serviceExecution := range s.keptnServiceExecutions {
		for eventType, count := range serviceExecution.eventTypeCount {
			keptnServiceExecs := exportedStatisticsService{
				Name:       keptnServiceName,
				Executions: count,
				EventType:  eventType,
			}

			summary.ServiceExecutions = append(summary.ServiceExecutions, keptnServiceExecs)
		}
	}
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

	from, to := getTimeFrame(statisticsFiles)
	statsOutput.from = from
	statsOutput.to = to

	for _, stats := range statisticsFiles {
		mergeStatisticsResponseIntoStatisticsOutput(stats, statsOutput, 0)
	}

	return []*statisticsOutput{statsOutput}
}

func addColumn(key, displayname string) {

	for _, col := range csvColumns {
		if col.key == key {
			return
		}
	}
	csvColumns = append(csvColumns, csvColumn{
		key:         key,
		displayName: displayname,
		values:      nil,
	})
	csvColumnMapping[key] = columnIndex
	columnIndex = columnIndex + 1

	for index, _ := range csvRows {
		csvRows[index] = append(csvRows[index], "")
	}
}

func createSeparatedStatistics(statisticsFiles map[string]*stats.GetStatisticsResponse) []*statisticsOutput {
	result := []*statisticsOutput{}

	index := 0
	for _, stats := range statisticsFiles {
		newSeparateOutput := &statisticsOutput{
			overallStatistics: statistics{
				name:                   "Overall: > Keptn",
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

		mergeStatisticsResponseIntoStatisticsOutput(stats, newSeparateOutput, index)
		result = append(result, newSeparateOutput)
		index = index + 1
	}
	return result
}

func mergeStatisticsResponseIntoStatisticsOutput(stats *stats.GetStatisticsResponse, statsOutput *statisticsOutput, rowIndex int) {
	for _, project := range stats.Projects {
		if len(excludeProjectsArr) > 0 && contains(excludeProjectsArr, project.Name) {
			continue
		}

		_, projectOverallDisplayName := getProjectColNames(project.Name)

		if isProjectGranularity() {
			if statsOutput.perProjectStatistics[project.Name] == nil {
				statsOutput.perProjectStatistics[project.Name] = &statistics{
					name:                   projectOverallDisplayName,
					automationUnits:        0,
					keptnServiceExecutions: map[string]*keptnServiceExecution{},
					triggers:               0,
					triggersByType:         map[string]*triggerExecution{},
					subStatistics:          map[string]*statistics{},
				}

			}
		}
		for _, svc := range project.Services {

			_, serviceOverallDisplayName := getServiceColNames(project.Name, svc.Name)

			if isServiceGranularity() {
				if statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name] == nil {
					statsOutput.perProjectStatistics[project.Name].subStatistics[svc.Name] = &statistics{
						name:                   serviceOverallDisplayName,
						automationUnits:        0,
						keptnServiceExecutions: map[string]*keptnServiceExecution{},
						triggers:               0,
						triggersByType:         map[string]*triggerExecution{},
						subStatistics:          nil,
					}
				}
			}
			for _, execution := range svc.KeptnServiceExecutions {

				if len(includeServicesArr) > 0 && !contains(includeServicesArr, execution.Name) {
					continue
				}
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
	rootCmd.PersistentFlags().StringVarP(&includeServices, "includeServices", "", "all", "List of Services that define an automation unit, default is 'all'")
	rootCmd.PersistentFlags().StringVarP(&excludeProjects, "excludeProjects", "", "", "List of project names that are excluded from the Summary")
	rootCmd.PersistentFlags().StringVarP(&includeTriggers, "includeTriggers", "", "all", "List of sequence triggers: [configuration-change, problem.open, evaluation-started] - supported with Keptn >0.8")
	rootCmd.PersistentFlags().StringVarP(&export, "export", "", "json", "The format to export the statistics, supported are [json, csv]")
	rootCmd.PersistentFlags().StringVarP(&separator, "separator", "", ",", "The separator used for the CSV exporter, allowed values are ',' or ';'")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "stats", "The name of the output file")
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

		// Search config in home directory with Name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
