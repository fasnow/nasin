package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const loginURL = "/nacos/v1/auth/users/login"

// ?&accessToken=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJuYWNvcyIsImV4cCI6MTY4NDUyNDgzNX0.m2pZysTFyqRJfCWYFRKXisqOtAg008ozUjcphoKFQNs&namespaceId=
const namespaceURL = "/nacos/v1/console/namespaces"

// ?dataId=&group=&appName=&config_tags=&pageNo=1&pageSize=10&tenant=&search=accurate
const configURL = "/nacos/v1/cs/configs"

type Auth struct {
	AccessToken string `json:"accessToken"`
	TokenTtl    int    `json:"tokenTtl"`
	GlobalAdmin bool   `json:"globalAdmin"`
}

type Namespace struct {
	Code    int             `json:"code"`
	Message interface{}     `json:"message"`
	Data    []NamespaceItem `json:"data"`
}

type NamespaceItem struct {
	Namespace         string `json:"namespace"`
	NamespaceShowName string `json:"namespaceShowName"`
	Quota             int    `json:"quota"`
	ConfigCount       int    `json:"configCount"`
	Type              int    `json:"type"`
}

type Config struct {
	TotalCount     int `json:"totalCount"`
	PageNumber     int `json:"pageNumber"`
	PagesAvailable int `json:"pagesAvailable"`
	PageItems      []struct {
		ID      interface{} `json:"id"`
		DataID  string      `json:"dataId"`
		Group   string      `json:"group"`
		Content string      `json:"content"`
		Md5     interface{} `json:"md5"`
		Tenant  string      `json:"tenant"`
		AppName string      `json:"appName"`
		Type    string      `json:"type"`
	} `json:"pageItems"`
}

type ConfigItem struct {
	name    string
	content string
}

type Nasin struct {
	*http.Client
	auth    Auth
	baseURL string
}

func NewNasinClient(url, username, password string) (*Nasin, error) {
	var nasin Nasin
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	nasin.baseURL = url
	nasin.Client = NewHttpClient()
	auth, err := nasin.login(username, password)
	if err != nil {
		return nil, err
	}
	nasin.auth = *auth
	return &nasin, nil
}

func (n *Nasin) GetProject() ([]NamespaceItem, error) {
	params := url.Values{}
	// ?&accessToken=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJuYWNvcyIsImV4cCI6MTY4NDUyNDgzNX0.m2pZysTFyqRJfCWYFRKXisqOtAg008ozUjcphoKFQN
	//s&namespaceId=

	params.Set("accessToken", n.auth.AccessToken)
	params.Set("namespaceId", "")
	request, err := http.NewRequest("GET", n.baseURL+namespaceURL, nil)
	if err != nil {
		return nil, err
	}
	authInfo, err := json.Marshal(n.auth)
	if err != nil {
		return nil, err
	}
	request.Header.Set("user-agent", "")
	request.Header.Set("Authorization", string(authInfo))
	resp, err := n.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("获取项目列表时返回非预期响应: %s\n响应体:\n%s", strconv.Itoa(resp.StatusCode), string(body)))
	}
	var proj Namespace
	err = json.Unmarshal(body, &proj)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("获取项目列表时返回非预期响应: %s\n响应体:\n%s\n程序错误信息:%s\n", strconv.Itoa(resp.StatusCode), string(body), err))
	}
	if proj.Code != 200 {
		return nil, errors.New(fmt.Sprintf("获取项目列表失败: %s", proj.Message))
	}
	return proj.Data, nil
}

func (n *Nasin) GetProjectConfig(data NamespaceItem) ([]ConfigItem, error) {
	params := url.Values{}
	params.Set("dataId", "")
	params.Set("group", "")
	params.Set("appName", "")
	params.Set("config_tags", "")
	params.Set("pageNo", strconv.Itoa(1))
	params.Set("pageSize", strconv.Itoa(data.ConfigCount))
	params.Set("tenant", data.Namespace)
	params.Set("search", "accurate")
	params.Set("accessToken", n.auth.AccessToken)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", n.baseURL+configURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	authInfo, _ := json.Marshal(n.auth)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "")
	request.Header.Set("Authorization", string(authInfo))
	response, err := n.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("获取项目配置列表失败: %s\n响应体:\n%s", strconv.Itoa(response.StatusCode), string(body)))
	}
	var config Config
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("获取项目配置列表时返回非预期响应: %s\n响应体:\n%s\n程序错误信息:%s\n", strconv.Itoa(response.StatusCode), string(body), err))
	}
	var list []ConfigItem
	for _, v := range config.PageItems {
		var tmp ConfigItem
		tmp.name = v.DataID
		tmp.content = v.Content
		list = append(list, tmp)
	}
	return list, nil
}

func (n *Nasin) LoginInfo() Auth {
	return n.auth
}

func (n *Nasin) login(username, password string) (*Auth, error) {
	data := strings.NewReader(fmt.Sprintf("username=%v&password=%v", username, password))
	request, err := http.NewRequest("POST", n.baseURL+loginURL, data)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := n.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("登陆时返回非预期响应: %s\n响应体:\n%s\n程序错误信息:%s\n", strconv.Itoa(response.StatusCode), string(body), err))
	}
	var authorization Auth
	err = json.Unmarshal(body, &authorization)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("登陆时返回非预期响应: %s\n响应体:\n%s\n程序错误信息:%s\n", strconv.Itoa(response.StatusCode), string(body), err))
	}
	return &authorization, nil
}
