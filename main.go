package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Ti struct {
	Time     string
	TimeFrom string
	TimeTo   string
	Code     uint8
}

func (t *Ti) String() string {
	return t.TimeFrom
}

type Cl struct {
	Code        string
	Name        string
	TeacherFull string
	Teacher     string
	Form        string
}

type Gr struct {
	Code string
	Name string
}

type Rm struct {
	Code uint8
	Name string
}

type Da struct {
	Day       uint8
	DayNumber uint8
	Ti        `json:"Time"`
	Cl        `json:"Class"`
	Gr        `json:"Group"`
	Rm        `json:"Room"`
}

func (d *Da) getStart() string {
	from, _ := time.Parse("2006-01-02T15:04:05", d.Ti.TimeFrom)
	return strconv.Itoa(from.Hour()) + ":" + strconv.Itoa(from.Minute())
}

func (d *Da) getEnd() string {
	to, _ := time.Parse("2006-01-02T15:04:05", d.Ti.TimeTo)
	return strconv.Itoa(to.Hour()) + ":" + strconv.Itoa(to.Minute())
}

func (d *Da) String() string {
	to, _ := time.Parse("2006-01-02T15:04:05", d.Ti.TimeTo)
	from, _ := time.Parse("2006-01-02T15:04:05", d.Ti.TimeFrom)
	fmt.Println("from: ", from)
	fmt.Println("to: ", to)
	return fmt.Sprintf("Para: %v\n\tCabinet: %v\nTimeStart: %v\nTimeEnd: %v\n",
		d.Ti.Time, d.Rm.Name, strconv.Itoa(from.Hour())+":"+strconv.Itoa(from.Minute()), strconv.Itoa(to.Hour())+":"+strconv.Itoa(from.Minute()))
}

type DaAr struct {
	T       []Ti `json:"Times"`
	D       []Da `json:"Data"`
	Semestr string
}

func (d *DaAr) getRasp(day, dayNumber uint8) []*Da {
	arrDat := make([]*Da, 0)
	for _, dat := range d.D {
		if dat.Day == day && dat.DayNumber == dayNumber {
			copyD := dat
			arrDat = append(arrDat, &copyD)
		}
	}
	return arrDat
}

type WeekD int

func toWeekD(day time.Weekday) WeekD {
	if day == 0 {
		return Sunday
	}
	return WeekD(day - 1)
}

func (day WeekD) String() string {
	switch day {
	case Monday:
		return "Monday"
	case Tuesday:
		return "Tuesday"
	case Wednesday:
		return "Wednesday"
	case Thursday:
		return "Thursday"
	case Friday:
		return "Friday"
	case Saturday:
		return "Saturday"
	case Sunday:
		return "Sunday"
	default:
		return ""
	}
}

const (
	Monday WeekD = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

func updateTime(clock *widget.Label) {
	formatted := time.Now().Format("Time: 03:04:05")
	clock.SetText(formatted)
}

func main() {
	//saveJson(getJson())
	//osnova()
	//netTest()
	try()
}

type s int

func (num *s) next() {
	if *(num) < 2 {
		*(num) += 1
	} else {
		*(num) = 0
	}
}

func (num s) toInt() int {
	return int(num)
}

func (num *s) reset() {
	*(num) = 0
}

func (num s) String() string {
	return strconv.Itoa(num.toInt())
}

func getDa(w fyne.Window, dat string) (Day, DayNumber uint8) {
	specificDate := time.Date(2023, time.October, 23, 1, 0, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		dialog.ShowError(fmt.Errorf("error load Moscow UTC Time"), w)
	} else {
		specificDate = specificDate.In(loc)
	}
	fmt.Println(specificDate)
	currentData := time.Now()
	if dat == "Завтра" {
		currentData = currentData.Add(time.Hour * 24)
	}
	//currentData := time.Now()
	//currentData := time.Date(2023, time.October, 23, 1, 0, 0, 0, time.UTC)
	diff := currentData.Sub(specificDate).Hours()
	dayOfStart := math.Floor(diff / 24)
	fmt.Println(dayOfStart)
	DayNumber = uint8(math.Floor(dayOfStart/7)) % 4
	Day = uint8(dayOfStart) % 7
	fmt.Println("Day: ", Day)
	fmt.Println("DayNumber: ", DayNumber)
	currentDay := toWeekD(currentData.Weekday())
	fmt.Println(currentDay)
	return Day, DayNumber
}

func isStorageExist(a *fyne.App, w *fyne.Window) (isExist bool, dataDir string) {
	dataDir = (*a).Storage().RootURI().Path()
	if dataDir == "" {
		info := dialog.NewInformation("Storage error", "hmm, no storage", *w)
		info.Show()
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
		return false, ""
	}
	return true, dataDir
}

func isDataJsonExist(a *fyne.App, w *fyne.Window) (bool, string, error) {
	if ok, dataDir := isStorageExist(a, w); ok {
		filePath := filepath.Join(dataDir, "data.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false, filePath, nil
		} else if err != nil {
			return false, filePath, err
		} else {
			return true, filePath, nil
		}
	}
	return false, "", nil
}

func deleteJson(jsonPath string) error {
	err := os.Remove(jsonPath)
	return err
}

func decodeJsonData(fi *os.File, data *DaAr, w *fyne.Window) error {
	dec := json.NewDecoder(fi)
	if err := dec.Decode(&data); err != nil {
		info := dialog.NewInformation("Error decoding json", "while decoding data.json occured error", *w)
		info.Show()
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
		return err
	} else {
		return nil
	}
}

func openData(w *fyne.Window, jsonPath string, data *DaAr) error {
	fi, err := os.Open(jsonPath)
	if err != nil {
		info := dialog.NewInformation("Error while open json", fmt.Sprintf("Open json occured: %v", err), *w)
		info.Show()
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
		return err
	}
	defer fi.Close()
	err = decodeJsonData(fi, data, w)
	if err != nil {
		if deleteJson(jsonPath) != nil {
			info := dialog.NewInformation("Error delete json", fmt.Sprintf("Delete json occured: %v", err), *w)
			info.Show()
			go func() {
				<-time.After(1 * time.Second)
				info.Hide()
			}()
			return err
		}
	}
	return err
}

func createClassNum(numb string) string {
	if len(numb) > 10 {
		return "no class number"
	}
	return numb
}

func createListOfPars(w *fyne.Window, value string, data *DaAr, li *fyne.Container) *fyne.Container {
	Day, DayNumber := getDa(*w, value)
	dA := data.getRasp(Day+1, DayNumber)
	for _, it := range dA {
		card := widget.NewCard(it.Ti.Time, it.Rm.Name, container.NewVBox(
			widget.NewLabel(it.Cl.Name),
			widget.NewLabel(it.Cl.Teacher),
			widget.NewLabel(it.getStart()),
			widget.NewLabel(it.getEnd()),
		))
		line := canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113})
		li.Add(container.NewVBox(card, line))
	}
	return li
}

// Парралельное использование
func setLoading(ok *widget.Label, c <-chan struct{}) {
	loading := []string{"Получение данных\\...", "Получение данных.|..", "Получение данных../."}
	var num s
	for {
		select {
		case <-c:
			num.reset()
			ok.SetText("Данные получены")
			return
		default:
			ok.SetText(loading[num])
			num.next()

			time.Sleep(1 * time.Millisecond * 150)
		}
	}
}

// Парралельное использование
func setParsing(ok *widget.Label, c <-chan struct{}) {
	parsing := []string{"Парсинг\\...", "Парсинг.|..", "Парсинг../."}
	var num s
	for {
		select {
		case <-c:
			ok.SetText("Данные распаршены")
			return
		default:
			ok.SetText(parsing[num])
			(&num).next()

			time.Sleep(1 * time.Millisecond * 150)
		}
	}
}

func parsingData(ch chan<- struct{}, bArr []byte, data *DaAr) error {
	if err := json.Unmarshal(bArr, &data); err != nil {
		return fmt.Errorf("Ошибка при распарсивании JSON:%v", err)
	}
	ch <- struct{}{}
	return nil
}

func createAndSaveData(bArr []byte, jsonPath string) error {
	file, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании json:%v", err)
	}
	defer file.Close()
	_, err = file.Write(bArr)
	if err != nil {
		return fmt.Errorf("ошибка при записи в json:%v", err)
	}
	return nil
}

func initData(a *fyne.App, w *fyne.Window, content *fyne.Container, textState *widget.Label, data *DaAr, value string, li *fyne.Container) {
	if ext, jsonPath, err := isDataJsonExist(a, w); err != nil {
		info := dialog.NewInformation("Error data.json", "error while get data.json", *w)
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
	} else if ext {
		if openData(w, jsonPath, data) == nil {
			li = createListOfPars(w, value, data, li)
			content.Add(li)
			(*w).Content().Refresh()
		}
	} else {
		ch := make(chan struct{})
		go setLoading(textState, ch)
		bArr, err := getJson(ch)
		if err != nil {
			info := dialog.NewInformation("Error while send net request", err.Error(), *w)
			go func() {
				<-time.After(1 * time.Second)
				info.Hide()
			}()
		} else {
			go setParsing(textState, ch)
			parsingData(ch, bArr, data)
			if createAndSaveData(bArr, jsonPath) != nil {
				info := dialog.NewInformation("Error while save and write to json", err.Error(), *w)
				go func() {
					<-time.After(1 * time.Second)
					info.Hide()
				}()
			} else {
				initData(a, w, content, textState, data, value, li)
			}
		}
		//TODO: не существует json -> запросить net  и создать
	}
}

func try() {
	var data DaAr
	a := app.NewWithID("com.raspisanie.app")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("internet")
	li := container.NewVBox()
	w.Resize(fyne.Size{Width: 500})
	var content *fyne.Container
	var menu *fyne.Container
	currValue := "Сегодня"
	ok := widget.NewLabel("none data")
	menu = container.NewHBox(
		ok,
		widget.NewRadioGroup([]string{"Сегодня", "Завтра"}, func(value string) {
			if currValue != value {
				fmt.Println("currValue:", currValue)
				fmt.Println("value:", value)
				currValue = value
				li.RemoveAll()
				content.Remove(li)
				initData(&a, &w, content, ok, &data, value, li)
			}
		}),
	)
	content = container.NewVBox(
		menu,
	)
	initData(&a, &w, content, ok, &data, "Сегодня", li)
	scr := container.NewVScroll(content)
	w.SetContent(scr)
	w.ShowAndRun()
}

func netTest() {
	var data DaAr
	a := app.NewWithID("com.raspisanie.app")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("internet")
	w.Resize(fyne.Size{Width: 600})
	var content *fyne.Container
	ok := widget.NewLabel("none data")
	Day, DayNumber := getDa(w, "Сегодня")
	getData := widget.NewButton("Get Data", func() {
		loading := []string{"Получение данных\\...", "Получение данных.|..", "Получение данных../."}
		parsing := []string{"Парсинг\\...", "Парсинг.|..", "Парсинг../."}

		ch := make(chan struct{})

		dataDir := a.Storage().RootURI().Path()
		if dataDir == "" {
			dialog.ShowError(fmt.Errorf("error path"), w)
		}
		filePath := filepath.Join(dataDir, "data.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			info := dialog.NewInformation("No data.json", "Trying to create", w)
			info.Show()
			go func() {
				<-time.After(1 * time.Second)
				info.Hide()
			}()
			go func(c <-chan struct{}) {
				var num s
				for {
					select {
					case <-c:
						num.reset()
						ok.SetText("Данные получены")
						return
					default:
						ok.SetText(loading[num])
						num.next()

						time.Sleep(1 * time.Millisecond * 150)
					}
				}
			}(ch)

			bArr, _ := getJson(ch)
			go func(c <-chan struct{}) {
				var num s
				for {
					select {
					case <-c:
						ok.SetText("Данные распаршены")
						return
					default:
						ok.SetText(parsing[num])
						(&num).next()

						time.Sleep(1 * time.Millisecond * 150)
					}
				}
			}(ch)

			//// Распарсиваем JSON-данные
			if err := json.Unmarshal(bArr, &data); err != nil {
				fmt.Println("Ошибка при распарсивании JSON:", err)
			}
			ch <- struct{}{}
			fmt.Println(filePath)
			file, err := os.Create(filePath)
			if err != nil {
				dialog.ShowError(err, w)
			}
			defer file.Close()
			_, err = file.Write(bArr)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			ok.SetText("Data saved")
		} else if err != nil {
			dialog.ShowError(err, w)
		} else {
			fi, err := os.Open(filePath)
			if err != nil {
				dialog.ShowError(err, w)
			}
			defer fi.Close()
			dec := json.NewDecoder(fi)
			if err := dec.Decode(&data); err != nil {
				dialog.ShowError(err, w)
			} else {
				dialog.ShowInformation("data", data.Semestr, w)
				//var bA []string
				//for _, st := range data.getRasp(Day+1, DayNumber) {
				//bA = append(bA, st.String())
				//}
				//hmm := container.NewVBox(
				//	widget.NewLabel("Predmet"),
				//	widget.NewLabel("Prepod"),
				//	widget.NewLabel("Start"),
				//	widget.NewLabel("End"),
				//)
				dA := data.getRasp(Day+1, DayNumber)
				li := container.NewVBox()
				for _, it := range dA {
					card := widget.NewCard(it.Ti.Time, it.Rm.Name, container.NewVBox(
						widget.NewLabel(it.Cl.Name),
						widget.NewLabel(it.Cl.Teacher),
						widget.NewLabel(it.getStart()),
						widget.NewLabel(it.getEnd()),
					))
					line := canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113})
					li.Add(container.NewVBox(card, line))
				}
				content.Add(li)
				w.Content().Refresh()
				//fmt.Println("here: ", data.getRasp(Day+1, DayNumber))
				//fmt.Println(fmt.Sprintf("current date: %v\n%v", strings.Join(bA, "")))
				//ok.SetText(strings.Join(bA, ""))
			}
		}

		//ok.SetText(data.Semestr)
		//ok.SetText("data ok")
	})
	content = container.NewVBox(
		container.NewHBox(
			ok,
			widget.NewRadioGroup([]string{"Сегодня", "Завтра"}, func(value string) {
				Day, DayNumber = getDa(w, value)
				getData.OnTapped()
			}),
		),
		getData,
	)
	scr := container.NewVScroll(content)
	w.SetContent(scr)
	w.ShowAndRun()
}

func osnova() {
	t := Ti{
		Time:     "1 пара",
		Code:     1,
		TimeFrom: "0001-01-01T09:00:00",
		TimeTo:   "0001-01-01T10:20:00",
	}
	a := app.NewWithID("raspisanie.preferences")
	w := a.NewWindow("clock")
	clock := widget.NewLabel("")
	updateTime(clock)
	ok := widget.NewLabel("unknown")
	saveJs := widget.NewButton("Save Json", func() {

		// Получаем директорию внутреннего хранилища приложения
		if a.Preferences().StringWithFallback("dataDir", a.Preferences().String("fyne_storage")) == "" {
			var ChJs fyne.URI
			ChJs, err := storage.Child(a.Storage().RootURI(), "data.json")
			if err != nil {
				write, _ := storage.Writer(ChJs)
				dJs, _ := json.Marshal(t)
				write.Write(dJs)
				defer write.Close()
				ok.SetText("CreateJson")
			}
			closer, err := storage.Reader(ChJs)
			ok := widget.NewLabel("here")
			if err != nil {
				ok.SetText("error reader")
			} else {
				var t Ti
				c := []byte{}
				closer.Read(c)
				json.Unmarshal(c, &t)
				//ok.SetText(t.String())
				ok.SetText("Json exists")
				//ok.SetText(ChJs.Path())
			}
			//defer closer.Close()
		} else {
			ok.SetText("No dir")
		}
	})
	content := container.NewVBox(
		clock,
		ok,
		saveJs,
	)
	//settings := app.SettingsSchema{
	//	Theme: theme.DarkTheme(),
	//}
	a.Settings().SetTheme(theme.DarkTheme())
	//background := canvas.NewRectangle(color.RGBA{R: 24, G: 65, B: 196, A: 148})
	//background.Resize(content.MinSize())
	//contentWithBackground := container.New(layout.NewBorderLayout(nil, nil, nil, nil), background, content)
	w.SetContent(content)
	go func() {
		for range time.Tick(time.Second) {
			updateTime(clock)
		}
	}()
	w.ShowAndRun()
}

func getJson(ch chan<- struct{}) ([]byte, error) {
	client := &http.Client{}

	cookies := []*http.Cookie{
		{Name: "_ym_uid", Value: "164440448297613284"},
		{Name: "BITRIX_SM_UIDL", Value: "lmuff%40mail.ru"},
		{Name: "BITRIX_SM_SALE_UID", Value: "0"},
		{Name: "BITRIX_SM_LOGIN", Value: "lmuff%40mail.ru"},
		{Name: "__utma", Value: "236244190.1120334790.1644404482.1646161857.1646490205.6"},
		{Name: "top100_id", Value: "t1.-1.461407572.1655304532011"},
		{Name: "adtech_uid", Value: "40c4f10a-3ede-4682-ac97-f455f1bf7654%3Amiet.ru"},
		{Name: "t3_sid_NaN", Value: "s1.1090775154.1667383698943.1667384961953.3.3"},
		{Name: "BITRIX_SM_user_group_referrer", Value: "YTowOnt9"},
		{Name: "BX_USER_ID", Value: "df1586cd776de69c436b6613d96d2f79"},
		{Name: "BITRIX_SM_user_group_links", Value: "YToxOntpOjIxO2k6MTt9"},
		{Name: "_ym_d", Value: "1693490168"},
		{Name: "_ym_isad", Value: "1"},
		{Name: "wl", Value: "8149fe91664ed3fbcca4f0758882cb42"},
		{Name: "SL_G_WPT_TO", Value: "ru"},
		{Name: "SL_GWPT_Show_Hide_tmp", Value: "1"},
		{Name: "SL_wptGlobTipTmp", Value: "1"},
		{Name: "MIET_PHPSESSID", Value: "kCtzZOH1WmMtWouwYXjPz33sub0lG8xk"},
		{Name: "last_visit", Value: "1698205812607%3A%3A1698216612607"},
		{Name: "t3_sid_512157", Value: "s1.1952763010.1698216587168.1698216612613.209.3"},
	}

	url := "https://miet.ru/schedule/data?group=%D0%9F-21%D0%9C"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании запроса:%v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при выполнении запроса:%v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при чтении ответа:%v", err)
	}
	ch <- struct{}{}
	return body, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveJson(d []byte) {
	err := os.WriteFile("./data.json", d, 0666)
	check(err)
}
