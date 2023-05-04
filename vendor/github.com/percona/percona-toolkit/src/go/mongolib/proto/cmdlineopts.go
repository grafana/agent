package proto

type ProcessManagement struct {
	Fork bool `bson:"fork"`
}

type Replication struct {
	ReplSet string `bson:"replSet"`
}

type Sharding struct {
	ClusterRole string `bson:"clusterRole"`
}

type CloStorage struct {
	DbPath string `bson:"dbPath"`
	Engine string `bson:"engine"`
}

type CloSystemLog struct {
	Destination string `bson:"destination"`
	Path        string `bson:"path"`
}

type Parsed struct {
	Sharding          Sharding          `bson:"sharding"`
	Storage           CloStorage        `bson:"storage"`
	SystemLog         CloSystemLog      `bson:"systemLog"`
	Net               Net               `bson:"net"`
	ProcessManagement ProcessManagement `bson:"processManagement"`
	Replication       Replication       `bson:"replication"`
	Security          Security          `bson:"security"`
}

// Security is a struct to hold security related configs
type Security struct {
	KeyFile           string `bson:"keyFile"`
	ClusterAuthMode   string `bson:"clusterAuthMode"`
	Authorization     string `bson:"authorization"`
	JavascriptEnabled bool   `bson:javascriptEnabled"`
	Sasl              struct {
		HostName            string `bson:"hostName"`
		ServiceName         string `bson:"serverName"`
		SaslauthdSocketPath string `bson:"saslauthdSocketPath"`
	} `bson:"sasl"`
	EnableEncryption     bool   `bson:"enableEncryption"`
	EncryptionCipherMode string `bson:"encryptionCipherMode"`
	EncryptionKeyFile    string `bson:"encryptionKeyFile"`
	Kmip                 struct {
		KeyIdentifier             string `bson:"keyIdentifier"`
		RotateMasterKey           bool   `bson:"rotateMasterKey"`
		ServerName                string `bson:"serverName"`
		Port                      string `bson:"port"`
		ClientCertificateFile     string `bson:"clientCertificateFile"`
		ClientCertificatePassword string `bson:"clientCertificatePassword"`
		ServerCAFile              string `bson:"serverCAFile"`
	} `bson:"kmip"`
}

// NET config options. See https://docs.mongodb.com/manual/reference/configuration-options/#net-options
type Net struct {
	HTTP                   HTTP   `bson:"http"`
	SSL                    SSL    `bson:"ssl"`
	Port                   int64  `bson:"port"`
	BindIP                 string `bson:"bindIp"`
	MaxIncomingConnections int    `bson:"maxIncomingConnections"`
	WireObjectCheck        bool   `bson:"wireObjectCheck"`
	IPv6                   bool   `bson:"ipv6"`
	UnixDomainSocket       struct {
		Enabled         bool   `bson:"enabled"`
		PathPrefix      string `bson:"pathPrefix"`
		FilePermissions int64  `bson:"filePermissions"`
	} `bson:"unixDomainSocket"`
}

type HTTP struct {
	Enabled              bool    `bson:"enabled"`
	Port                 float64 `bson:"port"`
	JSONPEnabled         bool    `bson:"JSONPEnabled"`
	RESTInterfaceEnabled bool    `bson:"RESTInterfaceEnabled"`
}

// SSL config options. See https://docs.mongodb.com/manual/reference/configuration-options/#net-ssl-options
type SSL struct {
	SSLOnNormalPorts                    bool   `bson:"sslOnNormalPorts"` // deprecated since 2.6
	Mode                                string `bson:"mode"`             // disabled, allowSSL, preferSSL, requireSSL
	PEMKeyFile                          string `bson:"PEMKeyFile"`
	PEMKeyPassword                      string `bson:"PEMKeyPassword"`
	ClusterFile                         string `bson:"clusterFile"`
	ClusterPassword                     string `bson:"clusterPassword"`
	CAFile                              string `bson:"CAFile"`
	CRLFile                             string `bson:"CRLFile"`
	AllowConnectionsWithoutCertificates bool   `bson:"allowConnectionsWithoutCertificates"`
	AllowInvalidCertificates            bool   `bson:"allowInvalidCertificates"`
	AllowInvalidHostnames               bool   `bson:"allowInvalidHostnames"`
	DisabledProtocols                   string `bson:"disabledProtocols"`
	FIPSMode                            bool   `bson:"FIPSMode"`
}

type CommandLineOptions struct {
	Argv     []string `bson:"argv"`
	Ok       float64  `bson:"ok"`
	Parsed   Parsed   `bson:"parsed"`
	Security Security `bson:"security"`
}
