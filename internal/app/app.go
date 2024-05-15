package app

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/zongrade/rasp/internal/net"
	raspisanietypes "github.com/zongrade/rasp/internal/raspisanieTypes"
	"github.com/zongrade/rasp/internal/utils"
)

func Main() {
	a := app.NewWithID("com.raspisanie.app")
	a.Settings().SetTheme(theme.DarkTheme())
}

func TestFindPrepod() {
	a := app.NewWithID("com.raspisanie.app")
	return
	//TODO: сделать нормально эту функцию
	//TODO: сделать только один раз запрос кук
	GroupNameSlice := make([]string, 262)
	RaspisanieByteGroupSlice := make([]byte, 0, 10924252)
	//RaspisanieGroupSlice := make([]*raspisanietypes.RaspisanieGroup, 0, 262)
	RaspisanieGroupSlice := make([]raspisanietypes.RaspisanieGroup, 0, 262)
	var countGorutines = 10
	SyncChannel := make(chan struct{}, countGorutines*30)
	//RaspisanieChannel := make(chan *raspisanietypes.DaAr, countGorutines*30)
	RaspisanieChannel := make(chan []byte, countGorutines*30)
	startTime := time.Now()
	ByteListGroup, err := net.GetGroup(SyncChannel)
	<-SyncChannel
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = json.Unmarshal(ByteListGroup, &GroupNameSlice); err != nil {
		fmt.Println(err)
	}
	endTimeGroup := time.Now()
	elapsedTime := endTimeGroup.Sub(startTime)
	fmt.Printf("запрос названия групп завершился за: %s\n", elapsedTime)
	getData, err := net.OnesCookieGetDataFromMietUrl()
	if err != nil {
		fmt.Println(err)
		return
	}
	startTime = time.Now()
	for _, group := range GroupNameSlice {
		SyncChannel <- struct{}{}
		//go getRasp(getData, group, RaspisanieChannel, SyncChannel)
		go getRasp2(getData, group, RaspisanieChannel, SyncChannel)
	}
	RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, '[')
	for i := 0; i < len(GroupNameSlice); i++ {
		//RaspisanieGroupSlice[i] = <-RaspisanieChannel
		if i == 0 {
			RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, <-RaspisanieChannel...)
		} else {
			RaspisanieByteGroupSlice = append(append(RaspisanieByteGroupSlice, []byte(",")...), <-RaspisanieChannel...)
		}
	}
	RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, ']')
	endTimeGroup = time.Now()
	elapsedTime = endTimeGroup.Sub(startTime)
	fmt.Printf("запрос расписания 262 групп завершился за: %s\n", elapsedTime)
	startTime = time.Now()
	if err := json.Unmarshal(RaspisanieByteGroupSlice, &RaspisanieGroupSlice); err != nil {
		fmt.Println(err)
	}
	close(RaspisanieChannel)
	endTimeGroup = time.Now()
	elapsedTime = endTimeGroup.Sub(startTime)
	fmt.Printf("конвертация 262 групп завершилась за: %s\n", elapsedTime)
	if ok, err := utils.Save(&a, RaspisanieByteGroupSlice, "allData.json"); err != nil {
		fmt.Printf("сохранить не удалось %v\n", err)
	} else if !ok {
		fmt.Println("сохранить не удалось, нормальный файл уже существует")
	} else {
		fmt.Println("сохранить удалось")
	}
}

func getRasp2(getData func(string) ([]byte, error), group string, RaspisanieChannel chan<- []byte, SyncChannel <-chan struct{}) {
	bArr, err := getData(group)
	if err != nil {
		<-SyncChannel
		RaspisanieChannel <- nil
	} else {
		<-SyncChannel
		RaspisanieChannel <- bArr
	}
}
func getRasp(getData func(string) ([]byte, error), group string, RaspisanieChannel chan<- *raspisanietypes.RaspisanieGroup, SyncChannel <-chan struct{}) {
	bArr, err := getData(group)
	if err != nil {
		<-SyncChannel
		RaspisanieChannel <- nil
	} else {
		<-SyncChannel
		datA := &raspisanietypes.RaspisanieGroup{}
		json.Unmarshal(bArr, datA)
		RaspisanieChannel <- datA
	}
}
