// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

// GlobalFlagKubeConfig - global flag for specifying the location of the kube config
const GlobalFlagKubeConfig = "kubeconfig"
const GlobalFlagKubeConfigHelp = "Path to the kubeconfig file to use"

// GlobalFlagContext - global flag for specifying which kube config context to use
const GlobalFlagContext = "context"
const GlobalFlagContextHelp = "The name of the kubeconfig context to use" // Flags that are common to more than one command
const (
	// GlobalFlagHelp - global help flag
	GlobalFlagHelp           = "help"
	WaitFlag                 = "wait"
	WaitFlagHelp             = "Wait for the command to complete and stream the logs to the console. The wait period is controlled by --timeout."
	WaitFlagDefault          = true
	TimeoutFlag              = "timeout"
	TimeoutFlagHelp          = "Limits the amount of time a command will wait to complete"
	VPOTimeoutFlag           = "platform-operator-timeout"
	VPOTimeoutFlagHelp       = "Limits the amount of time a command will wait for the Verrazzano Platform Operator to be ready"
	VersionFlag              = "version"
	VersionFlagDefault       = "latest"
	VersionFlagInstallHelp   = "The version of Verrazzano to install"
	VersionFlagUpgradeHelp   = "The version of Verrazzano to upgrade to"
	DryRunFlag               = "dry-run"
	SetFlag                  = "set"
	SetFlagShorthand         = "s"
	SetFlagHelp              = "Override a Verrazzano resource value (e.g. --set profile=dev).  This flag can be specified multiple times."
	OperatorFileFlag         = "operator-file" // an alias for the manifests flag
	OperatorFileDeprecateMsg = "Use --manifests instead"
	ManifestsFlag            = "manifests"
	ManifestsShorthand       = "m"
	ManifestsFlagHelp        = "The location of the manifests used to install or upgrade Verrazzano. This can be a URL or the path to a local file. The default is the verrazzano-platform-operator.yaml file of the specified (or most recent) version of Verrazzano."
	ImageRegistryFlag        = "image-registry"
	ImageRegistryFlagHelp    = "The private registry where Verrazzano image repositories are located. If unspecified, the public Verrazzano image registry will be used."
	ImageRegistryFlagDefault = ""
	ImagePrefixFlag          = "image-prefix"
	ImagePrefixFlagHelp      = "The prefix to use for all Verrazzano image names within the private image-registry. If unspecified, the default image prefixes in the Verrazzano image registry will be used."
	ImagePrefixFlagDefault   = ""
	LogFormatFlag            = "log-format"
	LogFormatHelp            = "The format of the log output. Valid output formats are \"simple\" and \"json\"."
	FilenameFlag             = "filename"
	FilenameFlagShorthand    = "f"
	FilenameFlagHelp         = "Path to file containing Verrazzano custom resource.  This flag can be specified multiple times to overlay multiple files.  Specifying \"-\" as the filename accepts input from stdin."
	SkipConfirmationFlagHelp = "Non-interactive mode - assumes the answers to all interactive questions to be 'Y'."
	SkipConfirmationFlag     = "skip-confirmation"
	SkipConfirmationShort    = "y"
	VerboseFlag              = "verbose"
	VerboseFlagShorthand     = "v"
	VerboseFlagDefault       = false
	VerboseFlagUsage         = "Enable verbose output."
	ReadOnly                 = "read-only file system"
	AutoBugReportFlag        = "auto-bug-report"
	AutoBugReportFlagDefault = true
	AutoBugReportFlagHelp    = "Automatically call vz bug-report if command fails"
	VzAnalysisReportTmpFile  = "details-*.out"
	// DatetimeFormat - suffix to vz bug report file in yyyymmddhhmmss format
	DatetimeFormat = "20060102150405"
)

// VerrazzanoReleaseList - API for getting the list of Verrazzano releases
const VerrazzanoReleaseList = "https://api.github.com/repos/verrazzano/verrazzano/releases"

// VerrazzanoOperatorURL - URL for downloading Verrazzano operator.yaml
const VerrazzanoOperatorURL = "https://github.com/verrazzano/verrazzano/releases/download/%s/operator.yaml"

// VerrazzanoPlatformOperatorURL - URL for downloading verrazzano-platform-operator.yaml
const VerrazzanoPlatformOperatorURL = "https://github.com/verrazzano/verrazzano/releases/download/%s/verrazzano-platform-operator.yaml"

const VerrazzanoPlatformOperator = "verrazzano-platform-operator"

const VerrazzanoPlatformOperatorWebhook = "verrazzano-platform-operator-webhook"

const VerrazzanoMysqlInstallValuesWebhook = "verrazzano-platform-mysqlinstalloverrides"

const VerrazzanoRequirementsValidatorWebhook = "verrazzano-platform-requirements-validator"

const VerrazzanoApplicationOperator = "verrazzano-application-operator"

const VerrazzanoClusterOperator = "verrazzano-cluster-operator"

const VerrazzanoMonitoringOperator = "verrazzano-monitoring-operator"

const VerrazzanoUninstall = "verrazzano-uninstall"

const VerrazzanoInstall = "verrazzano-install"

const VerrazzanoManagedCluster = "verrazzano-managed-cluster"

const VerrazzanoPlatformOperatorWait = 1

const OAMAppConfigurations = "applicationconfigurations"

const OAMMCAppConfigurations = "multiclusterapplicationconfigurations"

const OAMMCCompConfigurations = "multiclustercomponents"

const OAMComponents = "components"

const OAMMetricsTraits = "metricstraits"

const OAMIngressTraits = "ingresstraits"

const OAMProjects = "verrazzanoprojects"

const OAMManagedClusters = "verrazzanomanagedclusters"

const VerrazzanoManagedLabel = "verrazzano-managed=true"

const LineSeparator = "-"

// MysqlBackupMutatingWebhookName specifies the name of mysql webhook.
const MysqlBackupMutatingWebhookName = "verrazzano-mysql-backup"

// Analysis tool flags
const (
	DirectoryFlagName  = "capture-dir"
	DirectoryFlagValue = ""
	DirectoryFlagUsage = "Directory holding the captured data."

	ReportFileFlagName  = "report-file"
	ReportFileFlagValue = ""
	ReportFileFlagUsage = "Name of the report output file. (default stdout)"

	ReportFormatFlagName  = "report-format"
	ReportFormatFlagUsage = "The format of the report output. Valid report formats are \"summary\" and \"detailed\"."

	SummaryReport  = "summary"
	DetailedReport = "detailed"
)

// Constants for bug report
const (
	BugReportLogFlagDefault   = false
	BugReportFileFlagName     = "report-file"
	BugReportFileFlagValue    = ""
	BugReportFileFlagShort    = "r"
	BugReportFileFlagUsage    = "The report file created by the vz bug-report command, as a *.tar.gz file. Defaults to vz-bug-report-datetime-xxxx.tar.gz in the current directory."
	BugReportFileDefaultValue = "vz-bug-report-dt-*.tar.gz"

	BugReportIncludeNSFlagName  = "include-namespaces"
	BugReportIncludeNSFlagShort = "i"
	BugReportIncludeNSFlagUsage = "A comma-separated list of namespaces, in addition to the ones collected by default (system namespaces), for collecting cluster information. This flag can be specified multiple times, such as --include-namespaces ns1 --include-namespaces ns..."

	BugReportDir = "bug-report"

	// File name for the log captured from the pod
	LogFile = "logs.txt"

	// File containing list of resources captured by the tool
	BugReportOut = "bug-report.out"
	BugReportErr = "bug-report.err"

	BugReportError   = "ERROR: The bug report noticed one or more issues while capturing the resources. Please go through error(s) in the standard error."
	BugReportWarning = "WARNING: Please examine the contents of the bug report for any sensitive data"

	// File names for the various resources
	VzResource       = "verrazzano-resources.json"
	DeploymentsJSON  = "deployments.json"
	EventsJSON       = "events.json"
	PodsJSON         = "pods.json"
	ServicesJSON     = "services.json"
	ReplicaSetsJSON  = "replicasets.json"
	DaemonSetsJSON   = "daemonsets.json"
	IngressJSON      = "ingress.json"
	StatefulSetsJSON = "statefulsets.json"
	AppConfigJSON    = "application-configurations.json"
	ComponentJSON    = "components.json"
	IngressTraitJSON = "ingress-traits.json"
	MetricsTraitJSON = "metrics-traits.json"
	McAppConfigJSON  = "multicluster-application-configurations.json"
	McComponentJSON  = "multicluster-components.json"
	VzProjectsJSON   = "verrazzano-projects.json"
	VmcJSON          = "verrazzano-managed-clusters.json"

	// Indentation when the resource is marshalled as Json
	JSONIndent = "  "

	// The prefix used for the json.MarshalIndent
	JSONPrefix = ""

	// Top level directory for the bug report, keeping cluster-snapshot for now to support the analyze the command
	BugReportRoot = "cluster-snapshot"

	// Label for application
	AppLabel               = "app"
	K8SAppLabel            = "k8s-app"
	K8sAppLabelExternalDNS = "app.kubernetes.io/name"
	// Message prefix for bug-report and live cluster analysis
	BugReportMsgPrefix = "Capturing "
	AnalysisMsgPrefix  = "Analyzing "

	// Flag for capture pods logs( both additional and system namespaces)
	BugReportLogFlagName         = "include-logs"
	BugReportLogFlagNameShort    = "l"
	BugReportLogFlagNameUsage    = "Include logs from the pods in one or more namespaces; this is specified along with the --include-namespaces flag."
	BugReportTimeFlagName        = "duration"
	BugReportTimeFlagNameShort   = "d"
	BugReportTimeFlagDefaultTime = 0
	BugReportTimeFlagNameUsage   = "The time period during which the logs are collected in seconds, minutes, and hours."
)
