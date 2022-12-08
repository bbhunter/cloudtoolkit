package cache

import (
	"encoding/json"
	"log"

	"github.com/404tk/cloudtoolkit/utils"
)

type Credential struct {
	UUID      string
	User      string
	AccessKey string
	Provider  string
	JsonData  string
}

func (cfg *InitCfg) CredInsert(user string, data map[string]string) {
	provider, _ := data[utils.Provider]
	accessKey, _ := data[utils.AccessKey]
	uuid := utils.Md5Encode(accessKey + provider)
	if Cfg.CredSelect(uuid) != "" {
		return
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Println("[-] Map to json failed:", err.Error())
		return
	}

	cfg.Creds = append(cfg.Creds, Credential{
		UUID:      uuid,
		User:      user,
		AccessKey: accessKey,
		Provider:  provider,
		JsonData:  string(b),
	})
}

func (cfg *InitCfg) CredSelect(uuid string) string {
	for _, v := range cfg.Creds {
		if v.UUID == uuid {
			return v.JsonData
		}
	}
	return ""
}

func (cfg *InitCfg) CredSelectAll() (creds []string) {
	for _, v := range cfg.Creds {
		creds = append(creds, v.JsonData)
	}
	return
}

func (cfg *InitCfg) CredDelete(uuid string) {
	for index, v := range cfg.Creds {
		if v.UUID == uuid {
			if index == len(cfg.Creds) {
				cfg.Creds = cfg.Creds[:index]
			} else {
				cfg.Creds = append(cfg.Creds[:index], cfg.Creds[index+1:]...)
			}
		}
	}
}
