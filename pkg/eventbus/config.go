package eventbus

type AuthenticationType string

const (
	AuthenticationNone      AuthenticationType = ""
	AuthenticationPlain     AuthenticationType = "PLAINTEXT"
	AuthenticationAWSMskIAM AuthenticationType = "AWS_MSK_IAM"
)

type Config struct {
	Region         string             `yaml:"region"`
	Brokers        []string           `yaml:"brokers"`
	Topic          string             `yaml:"topic"`
	Authentication AuthenticationType `yaml:"authentication,omitempty"`
}
