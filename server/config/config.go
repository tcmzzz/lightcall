package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pocketbase/pocketbase/core"
)

type Provider interface {
	Dial() (*Dial, error)
	Privacy() (*Privacy, error)
	Cloud() (*Cloud, error)
	IceServers() ([]IceServer, error)
	ClearCache()
}

// 拨号配置 (name="dial")
type Dial struct {
	Caller struct {
		Affinity bool `json:"affinity"` // 主叫亲和性配置
	} `json:"caller"`
}

// 隐私配置 (name="privacy")
type Privacy struct {
	HideNumber bool `json:"hideNumber"` // 是否隐藏号码
}

// 云服务配置 (name="cloud")
type Cloud struct {
	Addr      string `json:"addr"`   // 服务地址
	AppID     string `json:"appid"`  // 应用ID
	Secret    string `json:"secret"` // 密钥
	Lifecycle struct {
		PreCall struct {
			Blacklist bool `json:"blacklist"` // 黑名单检查
			FlashCard bool `json:"flashCard"` // 闪卡提示
		} `json:"precall"`
	} `json:"lifecycle"`
}

// ICE服务器配置 (name="ice_servers")
type IceServer struct {
	URLs       string `json:"urls"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}

type instance struct {
	app   core.App
	cache *cache.Cache
}

var _ Provider = &instance{}

func New(app core.App) Provider {
	return &instance{
		app:   app,
		cache: cache.New(10*time.Second, 10*time.Second),
	}
}

func (i *instance) ClearCache() {
	i.cache.Flush()
}

func (i *instance) getConfig(section string) (string, error) {
	if val, found := i.cache.Get(section); found {
		return val.(string), nil
	}

	c, err := i.app.FindCollectionByNameOrId("config")
	if err != nil {
		return "", err
	}
	record, err := i.app.FindFirstRecordByFilter(c, fmt.Sprintf("name = '%s'", section))
	if err != nil {
		return "", err
	}
	value := record.GetString("value")

	i.cache.Set(section, value, 10*time.Second)
	return value, nil
}

func (i *instance) Dial() (*Dial, error) {
	str, err := i.getConfig("dial")
	if err != nil {
		return nil, err
	}
	ret := &Dial{}
	if err := json.Unmarshal([]byte(str), ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *instance) Privacy() (*Privacy, error) {
	str, err := i.getConfig("privacy")
	if err != nil {
		return nil, err
	}
	ret := &Privacy{}
	if err := json.Unmarshal([]byte(str), ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *instance) Cloud() (*Cloud, error) {
	str, err := i.getConfig("cloud")
	if err != nil {
		return nil, err
	}
	ret := &Cloud{}
	if err := json.Unmarshal([]byte(str), ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *instance) IceServers() ([]IceServer, error) {
	str, err := i.getConfig("ice_servers")
	if err != nil {
		return nil, err
	}
	var ret []IceServer
	if err := json.Unmarshal([]byte(str), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
