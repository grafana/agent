package build

func (s *ScrapeConfigBuilder) AppendStaticSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.StaticConfigs) == 0 {
		return
	}
	//targets := []string{"localhost"}
	//for i, sd := range s.cfg.ServiceDiscoveryConfig.StaticConfigs {
	//	sd.Source
	//}
}
