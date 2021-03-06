package gowb

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xiaolin8/gowb/pkg/config"
	"github.com/xiaolin8/gowb/pkg/db"
	gowbLog "github.com/xiaolin8/gowb/pkg/log"
	"github.com/xiaolin8/gowb/pkg/utils"
	"github.com/xiaolin8/gowb/pkg/web"
	"os"
	"runtime"
	"unsafe"
)

const logo = `
 _____   _____   _          __  _____  
/  ___| /  _  \ | |        / / |  _  \ 
| |     | | | | | |  __   / /  | |_| | 
| |  _  | | | | | | /  | / /   |  _  { 
| |_| | | |_| | | |/   |/ /    | |_| | 
\_____/ \_____/ |___/|___/     |_____/ 
`

type Gowb struct {
	ConfigName       string
	ConfigType       string
	Config           config.Config
	Routers          []web.Router
	RouterConfigs    []web.RouterConfigs
	AutoCreateTables []interface{}
	Middleware       []gin.HandlerFunc
}

func Bootstrap(g Gowb) (err error) {
	fmt.Println(logo)
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	//if !reflect.DeepEqual(g, Gowb{}) {
	if g.ConfigName != "" && g.ConfigType != "" {
		cu, err := utils.NewConfig(g.ConfigName, g.ConfigType)
		if err != nil {
			return err
		}
		// 解析并处理yaml
		var _config config.Config
		if err := cu.Unmarshal(&_config); err != nil {
			return err
		} else {
			if err := doBootstrap(g, _config); err != nil {
				return err
			}
		}
	} else if unsafe.Sizeof(g.Config) > 0 {
		if err := doBootstrap(g, g.Config); err != nil {
			return err
		}
	} else {
		return errors.New("configName and configType is empty")
	}
	return nil
}

func doBootstrap(g Gowb, config config.Config) error {
	c := context.WithValue(context.Background(), "routers", g.Routers)
	c = context.WithValue(c, "router_configs", g.RouterConfigs)
	c = context.WithValue(c, "config", config)
	c = context.WithValue(c, "middleware", g.Middleware)
	//初始化mysql
	if config.Mysql.Enabled {
		err := initMysql(c, g)
		if err != nil {
			return err
		}
	}

	//初始化日志
	err := gowbLog.InitLogger(c)
	if err != nil {
		return err
	}

	//初始化gin
	web.Bootstrap(c)
	return nil
}

func initMysql(c context.Context, g Gowb) error {
	err := db.InitMysql(c)
	if err != nil {
		return err
	}
	//建表
	for _, t := range g.AutoCreateTables {
		db.DB.AutoMigrate(t)
	}
	return nil
}
