package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

// Используется для русского отображения
type RuWeekD int

const (
	RuMonday    RuWeekD = iota //"Понедельник"
	RuTuesday                  //= "Вторник"
	RuWednesday                //= "Среда"
	RuThursday                 //= "Четверг"
	RuFriday                   //= "Пятница"
	RuSaturday                 //= "Суббота"
	RuSunday                   //= "Воскресенье"
)

func (rDay RuWeekD) String() string {
	switch rDay {
	case RuMonday:
		return "Понедельник"
	case RuTuesday:
		return "Вторник"
	case RuWednesday:
		return "Среда"
	case RuThursday:
		return "Четверг"
	case RuFriday:
		return "Пятница"
	case RuSaturday:
		return "Суббота"
	case RuSunday:
		return "Воскресенье"
	default:
		return "Понедельник"
	}
}

func stringToRuWeekD(day string) RuWeekD {
	switch day {
	case "Понедельник":
		return RuMonday
	case "Вторник":
		return RuTuesday
	case "Среда":
		return RuWednesday
	case "Четверг":
		return RuThursday
	case "Пятница":
		return RuFriday
	case "Суббота":
		return RuSaturday
	case "Воскресенье":
		return RuSunday
	default:
		return RuMonday
	}
}

func RuWeekDToList() []RuWeekD {
	return []RuWeekD{RuMonday, RuThursday, RuWednesday,
		RuThursday, RuFriday, RuSaturday, RuSunday}
}

func (rDay RuWeekD) ToWeekD() WeekD {
	switch rDay {
	case RuMonday:
		return Monday
	case RuTuesday:
		return Tuesday
	case RuWednesday:
		return Wednesday
	case RuThursday:
		return Thursday
	case RuFriday:
		return Friday
	case RuSaturday:
		return Saturday
	case RuSunday:
		return Sunday
	default:
		return Monday
	}
}

func isEqualWeekD(EnDay WeekD, RuDay RuWeekD) bool {
	return RuDay.ToWeekD() == EnDay
}

// Используется для английского отображения и работы с датой
type WeekD int

const (
	Monday WeekD = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

func ArrayWeekD() []WeekD {
	return []WeekD{
		Monday,
		Tuesday,
		Wednesday,
		Thursday,
		Friday,
		Saturday,
		Sunday,
	}
}

func ArrayWeekDString(lang string) (slice []string) {
	switch lang {
	case "ru":
		slice = []string{
			Monday.RuString(),
			Tuesday.RuString(),
			Wednesday.RuString(),
			Thursday.RuString(),
			Friday.RuString(),
			Saturday.RuString(),
			Sunday.RuString(),
		}
	default:
		slice = []string{
			Monday.String(),
			Tuesday.String(),
			Wednesday.String(),
			Thursday.String(),
			Friday.String(),
			Saturday.String(),
			Sunday.String(),
		}
	}
	return
}

func toWeekD(day time.Weekday) WeekD {
	if day == 0 {
		return Sunday
	}
	return WeekD(day - 1)
}

func stringToWeekd(day string) WeekD {
	switch day {
	case "Monday":
		return Monday
	case "Tuesday":
		return Tuesday
	case "Wednesday":
		return Wednesday
	case "Thursday":
		return Thursday
	case "Friday":
		return Friday
	case "Saturday":
		return Saturday
	case "Sunday":
		return Sunday
	default:
		return Monday
	}
}

func (day WeekD) RuString() string {
	switch day {
	case Monday:
		return "Понедельник"
	case Tuesday:
		return "Вторник"
	case Wednesday:
		return "Среда"
	case Thursday:
		return "Четверг"
	case Friday:
		return "Пятница"
	case Saturday:
		return "Суббота"
	case Sunday:
		return "Воскресенье"
	default:
		return "Понедельник"
	}
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
		return "Monday"
	}
}

func smth(startDay WeekD, endDay WeekD) (time.Time, time.Time) {
	now := time.Now()
	stTime := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, time.UTC)
	stTime = time.Date(now.Year(), now.Month(), toWeekD(stTime.Weekday()).diffBetwenDays(startDay)+stTime.Day(), 1, 0, 0, 0, time.UTC)
	enTime := stTime.Add(time.Duration(startDay.diffBetwenDays(endDay)) * 24 * time.Hour)
	return stTime, enTime
}

// func (startDay WeekD) diffBetwenDays(endDay WeekD) (int, time.Time) {
func (startDay WeekD) diffBetwenDays(endDay WeekD) int {
	diff := int(endDay - startDay)
	if startDay == Sunday && endDay == Sunday {
		diff = int(endDay + 1)
	}
	return diff
	//TODO: написать функцию высчитывающую дату endDay
	//now:=time.Now()
	//stTime :=  time.Date(2023,now.Month(),now.Day(),1,0,0,0,time.UTC())
	//return diff,obj.startTime.Add(24 * time.Duration(diff) * time.Hour).Weekday())
}

// TODO: функция использует stringToWeekd
//func (we WeekD) getRaspByDay(w *fyne.Window) *Da {
//	getDa(w, we.String())
//}

func main() {
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

func getDa(w *fyne.Window, dat string) (Day, DayNumber uint8) {
	specificDate := time.Date(2023, time.October, 23, 1, 0, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		info := dialog.NewInformation("Error load Time", fmt.Sprintf("error load Moscow UTC Time occured: %v", err), *w)
		info.Show()
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
	} else {
		specificDate = specificDate.In(loc)
	}
	fmt.Println(specificDate)
	currentData := time.Now()
	if dat == "Завтра" {
		currentData = currentData.Add(time.Hour * 24)
	} else if dat == Monday.String() {
		//buff := time.Now().Weekday().String()
		//currentData = toWeekD(buff).String()
	} else if dat == Tuesday.String() {
	} else if dat == Wednesday.String() {
	} else if dat == Thursday.String() {
	} else if dat == Friday.String() {
	} else if dat == Saturday.String() {
	} else if dat == Sunday.String() {
	}
	//TODO: можно принять в dat дни недели и в else это обработать
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

func isStorageExist2(a *fyne.App) (isExist bool, dataDir string) {
	dataDir = (*a).Storage().RootURI().Path()
	if dataDir == "" {
		return false, ""
	}
	return true, dataDir
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

// Существует ли data.json
// возвращает да/нет, путь до data, ошибку
func isDataJsonExist2(a *fyne.App) (bool, string, error) {
	if ok, dataDir := isStorageExist2(a); ok {
		filePath := filepath.Join(dataDir, "data.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false, filePath, nil
		} else if err != nil {
			return false, filePath, errors.New("error while get data.json")
		} else {
			return true, filePath, nil
		}
	}
	return false, "", nil
}

func isDataJsonExist(a *fyne.App, w *fyne.Window) (bool, string, error) {
	if ok, dataDir := isStorageExist(a, w); ok {
		filePath := filepath.Join(dataDir, "data.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("json not exist")
			return false, filePath, nil
		} else if err != nil {
			return false, filePath, err
		} else {
			fmt.Println("json exist")
			return true, filePath, nil
		}
	}
	return false, "", nil
}

func deleteJson(jsonPath string) error {
	err := os.Remove(jsonPath)
	return err
}

func decodeJsonData2(fi *os.File, data *DaAr) error {
	dec := json.NewDecoder(fi)
	if err := dec.Decode(&data); err != nil {
		return err
	} else {
		return nil
	}
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

// Открывает data.json по пути и декодирует в структуру
// возвращает ошибку
func openData2(jsonPath string, data *DaAr) error {
	fi, err := os.Open(jsonPath)
	if err != nil {
		return errors.New("Open json occured error")
	}
	defer fi.Close()
	err = decodeJsonData2(fi, data)
	if err != nil {
		err = errors.New("err decoding json open")
		if deleteJson(jsonPath) != nil {
			return errors.New("Error delete json")
		}
	}
	return err
}

func openData(w *fyne.Window, jsonPath string, data *DaAr) error {
	fi, err := os.Open(jsonPath)
	fmt.Println("jsonPath:", jsonPath)
	fmt.Println("json open")
	if err != nil {
		fmt.Println("err json open")
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
		fmt.Println("err decoding json open")
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
	if len(numb) > 14 {
		return "no class number"
	}
	return numb
}

func normalizeTime(t string) string {
	smh := strings.Split(t, ":")
	if len(smh[1]) < 2 {
		buff := strings.Split(t, ":")
		buff[1] += "0"
		return strings.Join(buff, ":")
	}
	return t
}

func isAfter(first, sec *Da) bool {
	numF, _ := strconv.Atoi(strings.Split(first.Time, " ")[0])
	numS, _ := strconv.Atoi(strings.Split(sec.Time, " ")[0])
	return numF > numS
}

func normalizeRasp(dA []*Da) {
	c := 0
	n := len(dA)
	swapped := true
	for swapped {
		swapped = false
		for i := 1; i < n; i++ {
			c++
			if isAfter(dA[i-1], dA[i]) {
				// Обмен значений
				dA[i-1], dA[i] = dA[i], dA[i-1]
				swapped = true
			}
		}
	}
	fmt.Println("c:", c)
}

func normalizeParName(s string) string {
	newS := strings.Split(s, " ")
	str := ""
	numb := 3
	for len(newS) > numb {
		buff := append(append(append([]string(nil), newS[0:numb]...), "\n"), newS[numb:]...)
		str += strings.Join(buff[0:numb+1], " ")
		newS = newS[numb:]
	}
	return str
}

func createListOfPars(w *fyne.Window, value string, data *DaAr, li *fyne.Container) *fyne.Container {
	Day, DayNumber := getDa(w, value)
	dA := data.getRasp(Day+1, DayNumber)
	normalizeRasp(dA)
	for _, it := range dA {
		//classNum - номер класса, Time - номер пары, Name - название пары,
		//Teacher - Имя препода, getStart()/getEnd() - время начала/конца пары
		card := widget.NewCard(it.Ti.Time, createClassNum(it.Rm.Name), container.NewVBox(
			widget.NewLabel(normalizeParName(it.Cl.Name)),
			widget.NewLabel(it.Cl.Teacher),
			widget.NewLabel(normalizeTime(it.getStart())),
			widget.NewLabel(normalizeTime(it.getEnd())),
		))
		line := canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113})
		li.Add(container.NewVBox(card, line))
	}
	return li
}

// сбрасывает текстовое поле, предполагается использовать VT.version
// в качестве base
func resetText(ok *widget.Label, base string) {
	ok.SetText(base)
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
		return fmt.Errorf("ошибка при распарсивании JSON:%v", err)
	}
	ch <- struct{}{}
	return nil
}

// создаётся data.json и в него записывается расписание
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

// заполняет виртуальное дерево и структуру содержащую расписание
// устанавливает textState
func initializeData(VT *VirtualTree, data *DaAr, textState *widget.Label) error {
	defer resetText(textState, VT.version)
	var globalError error
	if ext, jsonPath, globalError := isDataJsonExist2(VT.App); globalError != nil {
		return globalError
	} else if ext {
		//если существует то заполняется структура расписанием
		if openData2(jsonPath, data) == nil {
			return nil
		} else {
			return errors.New("Cannot open Data.json")
		}
	} else {
		//если нет то запрашиваем json с miet/schedule
		ch := make(chan struct{})
		go setLoading(textState, ch)
		bArr, globalError := getJson(ch)
		if globalError != nil {
			return globalError
		} else {
			go setParsing(textState, ch)
			globalError = parsingData(ch, bArr, data)
			if globalError != nil {
				return globalError
			} else if globalError = createAndSaveData(bArr, jsonPath); globalError != nil {
				return globalError
			}
		}
	}
	return globalError
}

func initData(a *fyne.App, w *fyne.Window, content *fyne.Container, textState *widget.Label, data *DaAr, value string, li *fyne.Container) error {
	var globalError error
	if ext, jsonPath, err := isDataJsonExist(a, w); err != nil {
		info := dialog.NewInformation("Error data.json", "error while get data.json", *w)
		go func() {
			<-time.After(1 * time.Second)
			info.Hide()
		}()
		globalError = err
		return globalError
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
			globalError = err
			return globalError
		} else {
			go setParsing(textState, ch)
			err = parsingData(ch, bArr, data)
			if err != nil {
				info := dialog.NewInformation("Error parsing data", err.Error(), *w)
				go func() {
					<-time.After(1 * time.Second)
					info.Hide()
				}()
				globalError = err
				return globalError
			} else if createAndSaveData(bArr, jsonPath) != nil {
				info := dialog.NewInformation("Error while save and write to json", err.Error(), *w)
				go func() {
					<-time.After(1 * time.Second)
					info.Hide()
				}()
				globalError = err
				return globalError
			} else {
				initData(a, w, content, textState, data, value, li)
			}
		}
	}
	return globalError
}

func createWeekdayButton(w *fyne.Window) {
	weekBtns := container.NewVBox()
	week := []WeekD{
		Monday, Tuesday,
		Wednesday, Thursday,
		Friday, Saturday,
	}
	for _, s := range week {
		weekBtns.Add(widget.NewButton(s.String(), func() {
			////TODO: сделать расписание по определённому дню через дату и название дня
			getDa(w, s.String())
		}))
	}
}

type VirtualTree struct {
	App     *fyne.App
	Head    *fyne.Container
	Body    *fyne.Container
	version string
}

type Settings struct {
	Version string
	Group   string
	IsLast  bool
}

func (s Settings) VersionString() string {
	if s.Version == "error" {
		return "error version"
	} else if s.Version != "" {
		if s.IsLast {
			return s.Version + "\nlast version"
		}
		return s.Version + "\nneed update"
	}
	return "no info"
}

// Версия которая вызывает функцию больше чем в параметре
func (s Settings) isVersionGreater(anS Settings) bool {
	anotherSettingsVerison := strings.Split(anS.Version, ".")
	for i, baseVersion := range strings.Split(s.Version, ".") {
		num1, err := strconv.Atoi(anotherSettingsVerison[i])
		if err != nil {
			return false
		}
		num2, err := strconv.Atoi(baseVersion)
		if err != nil {
			return false
		}
		if num1 > num2 {
			return false
		}
	}
	return true
}

func try() {
	var data DaAr
	a := app.NewWithID("com.raspisanie.app")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("internet")
	//li := container.NewVBox()
	w.Resize(fyne.Size{Width: 500, Height: 500})
	var content *fyne.Container
	var menu *fyne.Container
	var VT VirtualTree
	//TODO: виртуальное дерево для рендеринга
	//currValue := "Сегодня"
	ok := widget.NewLabel("none data")
	VT.App = &a
	VT.Head = container.NewHBox(
		ok,
		canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113}),
		widget.NewRadioGroup([]string{"Сегодня", "Завтра"}, func(value string) {
			//TODO: переделать
			fmt.Println(value)
			content.RemoveAll()
			VT.Body = fillBody(&data, idenDate(value))
			content.Add(VT.Body)
			w.Content().Refresh()
		}),
		canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113}),
		//TODO: here
		//widget.NewButton()
	)
	settings, err := getSettings2(VT.App)
	if err != nil {
		fmt.Println(err.Error())
		VT.version = "error"
		resetText(ok, VT.version)
	} else {
		VT.version = settings.Version
	}
	errs := initializeData(&VT, &data, ok)
	if errs != nil {
		w.SetContent(
			widget.NewLabel("errors occured: " + errs.Error()),
		)
	} else {
		VT.Body = fillBody(&data, idenDate(""))
		menu = container.NewVBox(
			VT.Head,
			canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113}),
		)
		content = VT.Body
		scr := container.NewVScroll(container.NewVBox(menu, content))
		w.SetContent(scr)
	}
	isUp, upV := checkUpdate()
	if (!isUp && upV == nil) || VT.version == "error" {
		resetText(ok, "error version")
	} else {
		if upV != nil && settings != nil {
			if upV.isVersionGreater(*settings) {
				settings.IsLast = false
			} else {
				settings.IsLast = true
			}
		}
	}
	if settings != nil && settings.IsLast {
		resetText(ok, VT.version+"\nlast version")
	} else {
		resetText(ok, VT.version+"\nneed update")
	}
	VT.Head.Add(container.NewGridWithColumns(3, createArrButtonRaspByDay(&data, content, &VT, &w)...))
	w.ShowAndRun()
}

func createArrButtonRaspByDay(data *DaAr, content *fyne.Container, VT *VirtualTree, w *fyne.Window) (slice []fyne.CanvasObject) {
	weekDs := ArrayWeekD()
	for i, dayName := range ArrayWeekDString("ru") {
		buffI := i
		buffDay := dayName
		slice = append(slice, widget.NewButton(buffDay, func() {
			fmt.Println(buffDay)
			content.RemoveAll()
			VT.Body = fillBody(data, weekDs[buffI])
			content.Add(VT.Body)
			(*w).Content().Refresh()
		}))
	}
	return
}

// Есть ли апдейт и его версию
func checkUpdate() (bool, *Settings) {
	url := "https://github.com/zongrade/raspisanie/tags"
	responce, err := http.Get(url)
	if err != nil {
		return false, nil
	}
	defer responce.Body.Close()
	body, err := io.ReadAll(responce.Body)
	if err != nil {
		return false, nil
	}
	re := regexp.MustCompile(`class="Link--primary Link">([^<]+)</a>`)
	anRe := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	type version struct {
		max uint64
		med uint64
		min uint64
	}
	max := version{}
	for _, st := range matches {
		bufAr := strings.Split(anRe.FindString(st[0]), ".")
		b := []uint64{}
		for _, st := range bufAr {
			ui, err := strconv.ParseUint(st, 10, 8)
			if err != nil {
				b = append(b, 0)
			} else {
				b = append(b, ui)
			}
		}
		if len(bufAr) == 3 {
			if max.max <= b[0] {
				max.max = b[0]
			}
			if max.med <= b[1] {
				max.med = b[1]
			}
			if max.min <= b[2] {
				max.min = b[2]
			}
		}
	}
	return true, &Settings{Version: fmt.Sprintf("%d.%d.%d", max.max, max.med, max.min)}
}

func tryLoadSettings() (*Settings, error) {
	bySettings := resourceSettingsJson.StaticContent
	settings := &Settings{}
	if err := json.Unmarshal(bySettings, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// получение файла settings.json, он должен включаться в билд
func getSettings(a *fyne.App) (*Settings, error) {
	res, err := fyne.LoadResourceFromPath("testdata/settings.json")
	if err != nil {
		return nil, err
	}
	settings := &Settings{}
	if err := json.Unmarshal(res.Content(), settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func getSettings2(a *fyne.App) (*Settings, error) {
	settings, err := tryLoadSettings()
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func fillBody(data *DaAr, wk WeekD) *fyne.Container {
	day, dayNumber := identificateDate(wk)
	dA := data.getRasp(day+1, dayNumber)
	normalizeRasp(dA)
	var bufB []fyne.CanvasObject
	//VT.Body = container.NewVBox()
	for _, it := range dA {
		card := widget.NewCard(it.Ti.Time, createClassNum(it.Rm.Name), container.NewVBox(
			widget.NewLabel(normalizeParName(it.Cl.Name)),
			widget.NewLabel(it.Cl.Teacher),
			widget.NewLabel(normalizeTime(it.getStart())),
			widget.NewLabel(normalizeTime(it.getEnd())),
		))
		line := canvas.NewLine(color.NRGBA{R: 24, G: 65, B: 196, A: 113})
		bufB = append(bufB, container.NewVBox(card, line))
		//VT.Body.Add(container.NewVBox(card, line))
	}
	return container.NewVBox(bufB...)
}

// после этой функции вызвать identificateDate
func idenDate(dt string) WeekD {
	switch dt {
	case "Сегодня":
		return toWeekD(time.Now().Weekday())
	case "Завтра":
		return toWeekD(time.Now().Add(time.Hour * 24).Weekday())
	default:
		return toWeekD(time.Now().Weekday())
	}
}

// Определяет дату
// TODO: доделать, совместить дни недели и сегодня/завтра
func identificateDate(dat WeekD) (day, dayNumber uint8) {
	specificDate := time.Date(2023, time.October, 23, 1, 0, 0, 0, time.UTC)
	currentData := time.Now()
	currDay := toWeekD(currentData.Weekday())
	tt := time.Duration(currDay.diffBetwenDays(dat))
	newDate := currentData.Add(tt * 24 * time.Hour)
	//TODO: day это день недели, dayNumber 1/2 числитель/знаменатель
	//dayNumber вычисляется относительно 23 октября
	fmt.Println(newDate)
	day = uint8(toWeekD(newDate.Weekday()))
	dayNumber = uint8(math.Floor(math.Abs(specificDate.Sub(newDate).Hours())/24/7)) % 4
	return
}

func getJson(ch chan<- struct{}) ([]byte, error) {

	fmt.Println("here get json")
	client := &http.Client{}

	cookies := getCookie()
	if len(cookies) < 1 {
		return nil, fmt.Errorf("ошибка при получении кук")
	}

	url := "https://miet.ru/schedule/data?group=%D0%9F-21%D0%9C"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса:%v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса:%v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа:%v", err)
	}
	ch <- struct{}{}
	return body, nil
}

// получение валидных кук
func getCookie() []*http.Cookie {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	cookie := []*http.Cookie{}
	getUrl := "https://miet.ru/schedule/"

	re, err := http.Get(getUrl)
	if err != nil {
		return nil
	}
	defer re.Body.Close()
	bodyBytes, err := io.ReadAll(re.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении тела ответа:", err)
		return nil
	}
	reg := regexp.MustCompile(`document.cookie="wl=([^;]+);`)
	matches := reg.FindAllStringSubmatch(string(bodyBytes), -1)
	if len(matches) > 0 {
		cookie = append(cookie, &http.Cookie{Name: "wl", Value: matches[0][1]})
	}
	fmt.Println("cookie first:", cookie)

	cookieUrl, _ := url.Parse(getUrl)

	jar.SetCookies(cookieUrl, cookie)

	resp, err := client.Get("https://miet.ru/schedule/")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	cookies := client.Jar.Cookies(resp.Request.URL)
	return cookies
}
