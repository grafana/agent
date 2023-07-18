package build

func (s *ScrapeConfigBuilder) AppendPushAPI() {
	if s.cfg.PushConfig == nil {
		return
	}

	//args := toLokiApiArguments(s.cfg.PushConfig)

}

//func toLokiApiArguments(config *scrapeconfig.PushTargetConfig) api.Arguments {
//	if config.ProfilingEnabled {
//		diags.Add(diag.SeverityLevelWarn, "server.profiling_enabled is not supported - use Agent's "+
//			"main HTTP server's profiling endpoints instead.")
//	}
//
//	if config.RegisterInstrumentation {
//		diags.Add(diag.SeverityLevelWarn, "server.register_instrumentation is not supported - Flow mode "+
//			"components expose their metrics automatically in their own metrics namespace")
//	}
//
//	if config.LogLevel.String() != "info" {
//		diags.Add(diag.SeverityLevelWarn, "server.log_level is not supported - Flow mode "+
//			"components may produce different logs")
//	}
//
//	if config.PathPrefix != "" {
//		diags.Add(diag.SeverityLevelWarn, "server.http_path_prefix is not supported - Flow mode's "+
//			"loki.source.api is available at /api/v1/push - see documentation for more details. If you are sending "+
//			"logs to this endpoint, the clients configuration may need to be updated.")
//	}
//
//	if config.HealthCheckTarget != nil && !*config.HealthCheckTarget {
//		diags.Add(diag.SeverityLevelWarn, "server.health_check_target disabling is not supported in Flow mode")
//	}
//
//	return api.Arguments{
//		Server: &fnet.ServerConfig{
//			HTTP: &fnet.HTTPConfig{
//				ListenAddress:      config.HTTPListenAddress,
//				ListenPort:         config.HTTPListenPort,
//				ConnLimit:          config.HTTPConnLimit,
//				ServerReadTimeout:  config.HTTPServerReadTimeout,
//				ServerWriteTimeout: config.HTTPServerWriteTimeout,
//				ServerIdleTimeout:  config.HTTPServerIdleTimeout,
//			},
//			GRPC: &fnet.GRPCConfig{
//				ListenAddress:              config.GRPCListenAddress,
//				ListenPort:                 config.GRPCListenPort,
//				ConnLimit:                  config.GRPCConnLimit,
//				MaxConnectionAge:           config.GRPCServerMaxConnectionAge,
//				MaxConnectionAgeGrace:      config.GRPCServerMaxConnectionAgeGrace,
//				MaxConnectionIdle:          config.GRPCServerMaxConnectionIdle,
//				ServerMaxRecvMsg:           config.GPRCServerMaxRecvMsgSize,
//				ServerMaxSendMsg:           config.GRPCServerMaxSendMsgSize,
//				ServerMaxConcurrentStreams: config.GPRCServerMaxConcurrentStreams,
//			},
//			GracefulShutdownTimeout: config.ServerGracefulShutdownTimeout,
//		},
//	}
//}
