package cli

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/cache"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
	logzap "github.com/kanopy-platform/hedgetrimmer/internal/log/zap"
	pkgadmission "github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	pkghandlers "github.com/kanopy-platform/hedgetrimmer/pkg/admission/handlers"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/pkg/mutators"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var scheme = runtime.NewScheme()

type RootCommand struct {
	k8sFlags *genericclioptions.ConfigFlags
}

func NewRootCommand() *cobra.Command {
	k8sFlags := genericclioptions.NewConfigFlags(true)

	root := &RootCommand{k8sFlags}

	cmd := &cobra.Command{
		Use:               "hedgetrimmer",
		PersistentPreRunE: root.persistentPreRunE,
		RunE:              root.runE,
	}

	cmd.PersistentFlags().String("log-level", "info", "Configure log level")
	cmd.PersistentFlags().Int("webhook-listen-port", 8443, "Admission webhook listen port")
	cmd.PersistentFlags().String("webhook-certs-dir", "/etc/webhook/certs", "Admission webhook TLS certificate directory")
	cmd.PersistentFlags().Bool("dry-run", false, "Controller dry-run changes only")
	cmd.PersistentFlags().Float64("default-memory-limit-request-ratio", 1.1, "Default memory limit/request ratio")
	cmd.PersistentFlags().StringSlice("resources", all_resources, "List of resources to enforce")

	k8sFlags.AddFlags(cmd.PersistentFlags())
	// no need to check err, this only checks if variadic args != 0
	_ = viper.BindEnv("kubeconfig", "KUBECONFIG")

	cmd.AddCommand(newVersionCommand())
	return cmd
}

func (c *RootCommand) persistentPreRunE(cmd *cobra.Command, args []string) error {
	// bind flags to viper
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// set log level
	logLevel, err := logzap.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return err
	}

	klog.SetLogger(zap.New(zap.Level(logLevel)))

	return nil
}

func (c *RootCommand) runE(cmd *cobra.Command, args []string) error {
	dryRun := viper.GetBool("dry-run")
	if dryRun {
		klog.Log.Info("running in dry-run mode.")
	}

	cfg, err := c.k8sFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	ctx := signals.SetupSignalHandler()

	mgr, err := manager.New(cfg, manager.Options{
		Scheme:                 scheme,
		Host:                   "0.0.0.0",
		Port:                   viper.GetInt("webhook-listen-port"),
		CertDir:                viper.GetString("webhook-certs-dir"),
		MetricsBindAddress:     "0.0.0.0:80",
		HealthProbeBindAddress: ":8080",
		LeaderElection:         true,
		LeaderElectionID:       "hedgetrimmer",
		DryRunClient:           dryRun,
	})

	if err != nil {
		klog.Log.Error(err, "unable to set up controller manager")
		return err
	}

	if err := configureHealthChecks(mgr); err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(cs, 1*time.Minute)

	lri := informerFactory.Core().V1().LimitRanges()
	lri.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {},
	})

	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	limitRanger := limitrange.NewLimitRanger(lri.Lister())

	ptm := mutators.NewPodTemplateSpec(
		mutators.WithDefaultMemoryLimitRequestRatio(viper.GetFloat64("default-memory-limit-request-ratio")),
		mutators.WithDryRun(viper.GetBool("dry-run")),
	)

	handlers, err := getHandlers(viper.GetStringSlice("resources"), ptm)
	if err != nil {
		return err
	}

	admissionRouter, err := admission.NewRouter(limitRanger,
		admission.WithAdmissionHandlers(handlers...),
	)
	if err != nil {
		return err
	}

	admissionRouter.SetupWithManager(mgr)

	return mgr.Start(ctx)
}

func configureHealthChecks(mgr manager.Manager) error {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}
	return nil
}

func getHandlers(resources []string, ptm pkgadmission.PodTemplateSpecMutator) ([]admission.AdmissionHandler, error) {
	var handlers []admission.AdmissionHandler
	var unexpected []string

	dedupedResources := make(map[string]bool)
	for _, resource := range resources {
		dedupedResources[strings.TrimSpace(resource)] = true
	}

	for resource := range dedupedResources {
		switch resource {
		case cronjobs:
			handlers = append(handlers, pkghandlers.NewCronjobHandler(ptm))
		case daemonsets:
			handlers = append(handlers, pkghandlers.NewDaemonSetHandler(ptm))
		case deployments:
			handlers = append(handlers, pkghandlers.NewDeploymentHandler(ptm))
		case jobs:
			handlers = append(handlers, pkghandlers.NewJobHandler(ptm))
		case pods:
			handlers = append(handlers, pkghandlers.NewPodHandler(ptm))
		case replicasets:
			handlers = append(handlers, pkghandlers.NewReplicaSetHandler(ptm))
		case replicationcontrollers:
			handlers = append(handlers, pkghandlers.NewReplicationControllerHandler(ptm))
		case statefulsets:
			handlers = append(handlers, pkghandlers.NewStatefulSetHandler(ptm))
		default:
			unexpected = append(unexpected, resource)
		}
	}

	if len(unexpected) > 0 {
		return []admission.AdmissionHandler{}, fmt.Errorf("unexpected resources: %v", unexpected)
	}

	return handlers, nil
}
