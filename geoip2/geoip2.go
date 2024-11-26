package geoip2

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/database"
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
	"gitlab.miliantech.com/infrastructure/log"
	"gitlab.miliantech.com/infrastructure/trace"
	"gitlab.miliantech.com/risk/base/risk_common/utils"
	"go.uber.org/zap"
)

type City geoip2.City

func ConvertToLocalCity(city *geoip2.City) (localCity *City) {
	if city == nil {
		return &City{}
	}
	return (*City)(unsafe.Pointer(city))
}

func (record City) GetTimeZone() string {
	return record.Location.TimeZone
}

func (record City) GetCountryName() string {
	if utils.ArrayContains([]string{"Asia/Shanghai", "Asia/Chongqing"}, record.GetTimeZone()) {
		return record.Country.Names["zh-CN"]
	}
	return record.Country.Names["en"]
}

func (record City) GetProvinceName() string {
	if len(record.Subdivisions) > 0 {
		if utils.ArrayContains([]string{"Asia/Shanghai", "Asia/Chongqing"}, record.GetTimeZone()) {
			proviceName := record.Subdivisions[0].Names["zh-CN"]
			return strings.ReplaceAll(strings.ReplaceAll(proviceName, "市", ""), "省", "")
		}
		return record.Subdivisions[0].Names["en"]
	}
	return ""
}

func (record City) GetCityName() string {
	if utils.ArrayContains([]string{"Asia/Shanghai", "Asia/Chongqing"}, record.GetTimeZone()) {
		cityName := record.City.Names["zh-CN"]
		return strings.ReplaceAll(cityName, "市", "")
	}
	return record.City.Names["en"]
}

// ======================================
//
//	geoipupdate databasae
//
// ======================================
const (
	configFileName   string = "GeoIP.conf"         // 下载离线 IP 地址库时的配置文件
	cityDatabaseName string = "GeoLite2-City.mmdb" //根据下面的 EditionIDs 定，可以从 downloadCityDatabse 函数中看出来
)

// 全局 离线 IP 地址库
var db *geoip2.Reader

// 关闭文件
func CloseDB() {
	db.Close()
}

// 配置文件内容
var configFileContent = ``

func Init() {
	PrepareConfigFile()
	OpenDB()
}

func SetConfigFileContent(ctt string) {
	configFileContent = ctt
}

// 准备配置文件
func PrepareConfigFile() error {
	return ioutil.WriteFile(configFileName, []byte(configFileContent), 0666)
}

func OpenDB() {
	var err error
	db, err = geoip2.Open("./GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println("open geo lite file failed", err)
		return
	}
	fmt.Println("open geo lite file success")
}

// 同步 geolite 数据
func UpdateCityDatabase() {
	ctx := trace.Context(context.Background())
	config, err := geoipupdate.NewConfig(configFileName, "", "./", true)
	if err != nil {
		log.YError(ctx, "updateCityDatabase loading configuration file failed", 0, 0, zap.Error(err))
		return
	}

	client := geoipupdate.NewClient(config)

	if err = downloadCityDatabse(client, config); err != nil {
		log.YError(ctx, "updateCityDatabase retrieving updates", 0, 0, zap.Error(err))
		return
	}
	log.YInfo(ctx, "updateCityDatabase download finish", 0, 0)
	newDB, err := geoip2.Open(cityDatabaseName)
	if err != nil {
		log.YError(ctx, "updateCityDatabase open new db failed", 0, 0, zap.Error(err))
		return
	}
	// 替换新的db
	if db != nil {
		if db.Metadata().BuildEpoch == newDB.Metadata().BuildEpoch {
			// 文件内容打包时间相同，认为是同一个文件，为了避免影响业务
			log.YInfo(ctx, "updateCityDatabase build epoch time is same, just return", 0, 0, zap.Any("data", db.Metadata()), zap.Any("go_text", newDB.Metadata()))
			return
		}
		db.Close()
	}
	db = newDB
	log.YInfo(ctx, "updateCityDatabase use new db finish", 0, 0, zap.Any("data", db.Metadata()))
}

func downloadCityDatabse(client *http.Client, config *geoipupdate.Config) error {
	dbReader := database.NewHTTPDatabaseReader(client, config)

	for _, editionID := range config.EditionIDs {
		filename, err := geoipupdate.GetFilename(config, editionID, client)
		if err != nil {
			return errors.Wrapf(err, "error retrieving filename for %s", editionID)
		}
		filePath := filepath.Join(config.DatabaseDirectory, filename)
		dbWriter, err := database.NewLocalFileDatabaseWriter(filePath, config.LockFile, config.Verbose)
		if err != nil {
			return errors.Wrapf(err, "error creating database writer for %s", editionID)
		}
		if err := dbReader.Get(dbWriter, editionID); err != nil {
			return errors.WithMessagef(err, "error while getting database for %s", editionID)
		}
	}
	return nil
}
