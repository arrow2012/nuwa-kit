package options

import "fmt"

// AuthOptions contains authentication-specific configuration
type CloudProviderOptions struct {
	AliyunUID             string `json:"ALIYUN_UID" mapstructure:"ALIYUN_UID"`
	AliyunAccessKeyID     string `json:"ALIBABA_CLOUD_ACCESS_KEY_ID" mapstructure:"ALIBABA_CLOUD_ACCESS_KEY_ID"`
	AliyunAccessKeySecret string `json:"ALIBABA_CLOUD_ACCESS_KEY_SECRET" mapstructure:"ALIBABA_CLOUD_ACCESS_KEY_SECRET"`
}

// NewServerOptions create a `zero` value instance.
func NewCloudProviderOptions() *CloudProviderOptions {
	return &CloudProviderOptions{
		AliyunUID:             "1570127113148550",
		AliyunAccessKeyID:     "",
		AliyunAccessKeySecret: "",
	}
}

func (o *CloudProviderOptions) Validate() []error {
	errs := []error{}
	if o.AliyunUID == "" {
		errs = append(errs, fmt.Errorf("aliyun uid cannot be empty"))
	}
	if o.AliyunAccessKeyID == "" {
		errs = append(errs, fmt.Errorf("aliyun access key id cannot be empty"))
	}
	if o.AliyunAccessKeySecret == "" {
		errs = append(errs, fmt.Errorf("aliyun access key secret cannot be empty"))
	}
	return errs
}
