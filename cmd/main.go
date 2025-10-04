/*
Copyright 2025 shtsukada.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	observabilityv1alpha1 "github.com/shtsukada/cloudnative-observability-operator/api/v1alpha1"
	internalcontrollers "github.com/shtsukada/cloudnative-observability-operator/internal/controller"
	tel "github.com/shtsukada/cloudnative-observability-operator/internal/shared/telemetry"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(observabilityv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func newZapLogger() *zap.Logger {
	json := os.Getenv("ZAP_LOG_JSON")
	level := os.Getenv("ZAP_LOG_LEVEL")
	sample := os.Getenv("ZAP_SAMPLE")

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if json == "false" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	if level == "" {
		level = "info"
	}
	var lvl zapcore.Level
	_ = lvl.UnmarshalText([]byte(level))
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	if sample == "" || sample == "false" {
		cfg.Sampling = nil
	}

	lg, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return lg
}

func main() {
	if err := run(); err != nil {
		// ここでは defer は使わない
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metrics endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.Parse()

	// ロガー初期化と controller-runtime への設定
	zl := newZapLogger()
	defer func() {
		// errcheck 対応：戻り値は明示破棄
		_ = zl.Sync()
	}()
	ctrl.SetLogger(zapr.NewLogger(zl))

	// Manager 構築
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cno-operator.shtsukada.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	// Tracer 初期化（shutdown は必ず実行される）
	shutdown, err := tel.InitTracer(context.Background())
	if err != nil {
		setupLog.Error(err, "failed to init tracer")
		return err
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	// コントローラ登録
	{
		gb := &internalcontrollers.GrpcBurnerReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("cloudnative-observability-operator"),
		}
		if err := gb.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "GrpcBurner")
			return err
		}
	}
	{
		oc := &internalcontrollers.ObservabilityConfigReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("cloudnative-observability-operator"),
		}
		if err := oc.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ObservabilityConfig")
			return err
		}
	}

	// ヘルスチェック
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}
	return nil
}
