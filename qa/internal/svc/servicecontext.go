package svc

import (
	"encoding/json"
	"os"

	"github.com/luyb177/XiaoAnBackend/qa/internal/config"
)

type ServiceContext struct {
	Config config.Config
	QAData []QAItem
}

func NewServiceContext(c config.Config) *ServiceContext {
	data, _ := os.ReadFile(c.DataPath)
	//fmt.Println(c.DataPath)
	var qaList []QAItem
	_ = json.Unmarshal(data, &qaList)
	//fmt.Println(qaList)
	return &ServiceContext{
		Config: c,
		QAData: qaList,
	}
}

type QAItem struct {
	ID       int      `json:"id"`
	Keywords []string `json:"keywords"`
	Answer   string   `json:"answer"`
}
