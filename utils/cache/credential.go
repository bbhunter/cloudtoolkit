package cache

import (
	"encoding/base64"
	"encoding/json"

	"github.com/404tk/cloudtoolkit/utils"
	"github.com/404tk/cloudtoolkit/utils/logger"
)

type Credential struct {
	UUID      string
	User      string
	AccessKey string
	Provider  string
	JsonData  string
	Note      string
}

func (cfg *InitCfg) CredInsert(user string, data map[string]string) {
	provider := data[utils.Provider]
	accessKey := data[utils.AccessKey]
	switch provider {
	case "azure":
		accessKey = data[utils.AzureClientId]
	case "gcp":
		tojson, _ := base64.StdEncoding.DecodeString(data[utils.GCPserviceAccountJSON])
		accessKey = utils.Md5Encode(string(tojson))
	}
	uuid := utils.Md5Encode(accessKey + provider)

	b, err := json.Marshal(data)
	if err != nil {
		logger.Error("Map to json failed:", err.Error())
		return
	}

	if Cfg.CredSelect(uuid) != "" {
		Cfg.CredUpdate(uuid, string(b))
	} else {
		cfg.Creds = append(cfg.Creds, Credential{
			UUID:      uuid,
			User:      truncateString(user, 20),
			AccessKey: truncateString(accessKey, 35),
			Provider:  provider,
			JsonData:  string(b),
		})
	}
}

func (cfg *InitCfg) CredSelect(uuid string) string {
	for _, v := range cfg.Creds {
		if v.UUID == uuid {
			return v.JsonData
		}
	}
	return ""
}

func (cfg *InitCfg) CredUpdate(uuid, data string) {
	for k, v := range cfg.Creds {
		if v.UUID == uuid {
			cfg.Creds[k].JsonData = data
			return
		}
	}
}

func (cfg *InitCfg) CredNote(uuid, data string) {
	for k, v := range cfg.Creds {
		if v.UUID == uuid {
			cfg.Creds[k].Note = data
			return
		}
	}
}

func (cfg *InitCfg) CredDelete(uuid string) {
	for index, v := range cfg.Creds {
		if v.UUID == uuid {
			if index == len(cfg.Creds) {
				cfg.Creds = cfg.Creds[:index]
			} else {
				cfg.Creds = append(cfg.Creds[:index], cfg.Creds[index+1:]...)
			}
			return
		}
	}
}

func truncateString(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
