package magic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"
	"text/template"

	"github.com/grafana/agent/pkg/config"

	"gopkg.in/yaml.v3"
)

type confighandler struct {
	cfgMutex sync.Mutex
	cfgText  string
	cfgPath  string
}

func newConfigHandle(cfgText string, cfgPath string) *confighandler {
	return &confighandler{
		cfgText: cfgText,
		cfgPath: cfgPath,
	}
}

func (ac *confighandler) IntegrationExists(name string) bool {
	ac.cfgMutex.Lock()
	defer ac.cfgMutex.Unlock()
	configNode := &yaml.Node{}
	err := yaml.Unmarshal([]byte(ac.cfgText), configNode)
	if err != nil {
		return false
	}
	k, _ := findNode(name, configNode.Content)
	return k != nil
}

func (ac *confighandler) GetIntegrationConfig(name string) bool {
	ac.cfgMutex.Lock()
	defer ac.cfgMutex.Unlock()
	configNode := &yaml.Node{}
	err := yaml.Unmarshal([]byte(ac.cfgText), configNode)
	if err != nil {
		return false
	}
	k, _ := findNode(name, configNode.Content)
	return k != nil
}

func (ac *confighandler) writeFile(config string) error {
	err := ioutil.WriteFile(ac.cfgPath, []byte(config), 0644)
	if err != nil {
		return err
	}
	ac.cfgText = config
	return nil
}

func (ac *confighandler) AddWindowsIntegration() error {
	ac.cfgMutex.Lock()
	defer ac.cfgMutex.Unlock()
	configString, err := ac.addWindowsIntegration()
	if err != nil {
		return err
	}
	return ac.writeFile(configString)
}

func (ac *confighandler) addWindowsIntegration() (string, error) {
	winInt := `
windows_exporter:
  autoscrape:
    enabled: true
`
	return ac.addSingletonIntegrationToConfig(winInt, nil, "windows_exporter")

}

func (ac *confighandler) addSingletonIntegrationToConfig(tmp string, data interface{}, integrationName string) (string, error) {
	configNode := &yaml.Node{}
	err := yaml.Unmarshal([]byte(ac.cfgText), configNode)
	if err != nil {
		return "", err
	}
	k, _ := findNode(integrationName, configNode.Content)
	if k != nil {
		return "", fmt.Errorf("%s already exists", integrationName)
	}
	integrationsKey, integrationsValue := findNode("integrations", configNode.Content)
	if integrationsKey == nil {
		integrationsKey, integrationsValue = addIntegrationsNode(configNode)
	}
	t, err := template.New("integration").Parse(tmp)
	if err != nil {
		return "", err
	}
	bb := bytes.Buffer{}
	err = t.Execute(&bb, data)
	if err != nil {
		return "", err
	}
	err = addIntegration(integrationsValue, bb.Bytes())
	if err != nil {
		return "", err
	}

	configText, err := yaml.Marshal(configNode)
	if err != nil {
		return "", err
	}
	return string(configText), nil
}

func addIntegrationsNode(configRoot *yaml.Node) (k *yaml.Node, v *yaml.Node) {
	integrationsKey := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "integrations",
	}
	integrationsValue := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	// using content 0 due to node wierdness with document root
	configRoot.Content[0].Content = append(configRoot.Content[0].Content, integrationsKey)
	configRoot.Content[0].Content = append(configRoot.Content[0].Content, integrationsValue)
	return integrationsKey, integrationsValue
}

func addIntegration(integrationsValue *yaml.Node, bb []byte) error {
	newNode := &yaml.Node{}
	err := yaml.Unmarshal(bb, newNode)
	if err != nil {
		return err
	}
	// Have to use newNode.Content[0] because DocumentNode lives at the top of the hierarchy
	integrationsValue.Content = append(integrationsValue.Content, newNode.Content[0].Content...)
	return nil
}

func findNode(name string, nodes []*yaml.Node) (key *yaml.Node, value *yaml.Node) {
	if len(nodes) == 0 {
		return nil, nil
	}
	if len(nodes) == 1 {
		return findNode(name, nodes[0].Content)
	}
	for i := 0; i < len(nodes); i = i + 2 {
		if nodes[i].Kind == yaml.ScalarNode && nodes[i].Value == name {
			return nodes[i], nodes[i+1]
		}
		k, v := findNode(name, nodes[i+1].Content)
		if k != nil {
			return k, v
		}
	}
	return nil, nil
}

func getDifferenceConfig(configNode *yaml.Node) (string, error) {

	// This whole bit just gives us the default config
	defConfig := &config.Config{}
	err := config.LoadBytes([]byte("server: {}"), false, defConfig)
	if err != nil {
		return "", err
	}
	defConfigString, err := yaml.Marshal(defConfig)
	if err != nil {
		return "", err
	}
	defYamlNode := &yaml.Node{}
	err = yaml.Unmarshal(defConfigString, defYamlNode)
	if err != nil {
		return "", err
	}
	return "", nil
}

func buildNodeMap(parent *NodeMapping, key *yaml.Node, value *yaml.Node, nodeMap []*NodeMapping) []*NodeMapping {
	path := "." + key.Value
	if parent != nil {
		path = parent.path + "." + key.Value
	}
	current := &NodeMapping{
		parent: parent,
		path:   path,
		key:    key,
		value:  value,
	}
	nodeMap = append(nodeMap, current)
	for i := 0; i < len(value.Content); i = i + 2 {
		nodeMap = buildNodeMap(current, value.Content[i], value.Content[i+1], nodeMap)
	}
	return nodeMap
}

type NodeMapping struct {
	parent *NodeMapping
	path   string
	key    *yaml.Node
	value  *yaml.Node
}
