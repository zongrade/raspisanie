package settings

import (
	"encoding/json"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"github.com/zongrade/rasp/internal/net"
	"github.com/zongrade/rasp/internal/utils"
)

type ConfigFile struct {
	FileName string
	IsExist  bool
	Filepath string
}

type ConfigFilepath struct {
	PersonSettings  ConfigFile
	PersonGroupData ConfigFile
	AllGroupsData   ConfigFile
	Groups          ConfigFile
}

func (c *ConfigFilepath) ToSlice() []*ConfigFile {
	return []*ConfigFile{&c.PersonSettings, &c.AllGroupsData, &c.Groups, &c.PersonGroupData}
}

func (c *ConfigFile) SetExist() {
	c.IsExist = true
}

func (c *ConfigFile) OpenFile() (*os.File, error) {
	file, err := os.OpenFile(c.Filepath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return file, err
}

var CurrentConfig = &ConfigFilepath{
	PersonSettings:  ConfigFile{FileName: "settings.json", IsExist: false, Filepath: ""},
	PersonGroupData: ConfigFile{FileName: "data.json", IsExist: false, Filepath: ""},
	AllGroupsData:   ConfigFile{FileName: "allData.json", IsExist: false, Filepath: ""},
	Groups:          ConfigFile{FileName: "groups.json", IsExist: false, Filepath: ""},
}

func InitConfig(a *fyne.App) error {
	checkConfig(a)
	if err := initGroups(a); err != nil {
		return err
	}
	if err := initAllGroupData(a); err != nil {
		return err
	}
	return nil
}

func initGroups(a *fyne.App) error {
	if !CurrentConfig.Groups.IsExist {
		if data, err := net.GetGroupSt(); err != nil {
			//error while request groups
			return err
		} else {
			if saved, filepath, err := utils.Save(a, data, CurrentConfig.Groups.FileName); err != nil {
				//error while save groups.json
				return err
			} else if !saved {
				CurrentConfig.Groups.Filepath = filepath
				fmt.Println("already exist")
			} else {
				CurrentConfig.Groups.Filepath = filepath
				fmt.Println("saved groups file")
			}
		}
	}
	return nil
}

// TODO: сделать вместо ссылки на файл, функцию возвращающую файл по пути
func initAllGroupData(a *fyne.App) error {
	if !CurrentConfig.AllGroupsData.IsExist && CurrentConfig.Groups.IsExist {
		file, err := CurrentConfig.Groups.OpenFile()
		defer file.Close()
		if err != nil {
			return err
		}
		bData := []byte{}
		if _, err := file.Read(bData); err != nil {
			return err
		}
		groups := []string{}
		if err := json.Unmarshal(bData, &groups); err != nil {
			return err
		}
		if data, err := net.GetAllRaspisanieSt(groups); err != nil {
			//error while request groups
			return err
		} else {
			if saved, filepath, err := utils.Save(a, data, CurrentConfig.AllGroupsData.FileName); err != nil {
				//error while save groups.json
				return err
			} else if !saved {
				CurrentConfig.AllGroupsData.Filepath = filepath
				fmt.Println("такой файл уже существует")
			} else {
				CurrentConfig.AllGroupsData.Filepath = filepath
				fmt.Println("saved groups file")
			}
		}
	}
	return nil
}

func initPersonSettings(a *fyne.App) error {
	if !CurrentConfig.PersonSettings.IsExist {
		file, err := CurrentConfig.Groups.OpenFile()
		defer file.Close()
		if err != nil {
			return err
		}
		bData := []byte{}
		if _, err := file.Read(bData); err != nil {
			return err
		}
		groups := []string{}
		if err := json.Unmarshal(bData, &groups); err != nil {
			return err
		}
		bData = []byte{}
		if CurrentConfig.AllGroupsData.Filepath != "" {
			CurrentConfig.Groups.File.Seek(0, 0)
			CurrentConfig.AllGroupsData.File.Read(bData)
		}
		//TODO: добавить взятие данных из AllGroupsData
	}
}

func checkConfig(a *fyne.App) {
	for _, fileSettings := range CurrentConfig.ToSlice() {
		if exist, filepath, err := utils.IsFileExist(a, fileSettings.FileName); err != nil {
			fmt.Printf("ошибка попытки найти %s: %v\n", fileSettings.FileName, err)
		} else if !exist {
			fmt.Printf("не существует %s\n", fileSettings.FileName)
		} else {
			fileSettings.SetExist()
			fmt.Printf("существует %s\n", filepath)
		}
	}
}
