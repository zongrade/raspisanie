package net

import (
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"

	"github.com/zongrade/rasp/internal/myerror"
)

// Хранит данные пользователя
type reqData struct {
	Client  *http.Client
	Cookies []*http.Cookie
}

var rd *reqData = &reqData{
	Client:  &http.Client{},
	Cookies: []*http.Cookie{},
}

// возвращает функцию которая запрашивает данные по группе
func OnesCookieGetDataFromMietUrl() (func(string) ([]byte, error), error) {
	cookies, err := getCookie()
	if err != nil || len(cookies) < 1 {
		return nil, myerror.CreateError("ошибка при получении кук")
	}
	rd.Cookies = cookies
	return func(group string) ([]byte, error) {
		req, errS := http.NewRequest("GET", "https://miet.ru/schedule/data?group="+group, nil)
		if errS != nil {
			return nil, myerror.CreateError("ошибка при создании запроса")
		}

		for _, cookie := range rd.Cookies {
			req.AddCookie(cookie)
		}

		response, errS := rd.Client.Do(req)
		if errS != nil {
			return nil, myerror.CreateError("ошибка при выполнении запроса")
		}
		defer response.Body.Close()

		body, errS := io.ReadAll(response.Body)
		if errS != nil {
			return nil, myerror.CreateError("ошибка при чтении ответа")
		}
		return body, nil
	}, nil
}

// Получает данные с miet за 2 запроса если ReqData пуста и 1 если ReqData полна
func getDataFromMietUrl(ch chan<- struct{}, baseUrl string) ([]byte, error) {
	cookies, err := getCookie()
	if err != nil || len(cookies) < 1 {
		return nil, myerror.CreateError("ошибка при получении кук")
	}
	rd.Cookies = cookies
	req, errS := http.NewRequest("GET", baseUrl, nil)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при создании запроса")
	}

	for _, cookie := range rd.Cookies {
		req.AddCookie(cookie)
	}

	response, errS := rd.Client.Do(req)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при выполнении запроса")
	}
	defer response.Body.Close()

	body, errS := io.ReadAll(response.Body)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при чтении ответа")
	}
	ch <- struct{}{}
	return body, nil
}

// Получает данные с miet за 2 запроса если ReqData пуста и 1 если ReqData полна
func getDataFromMietUrlSt(baseUrl string) ([]byte, error) {
	cookies, err := getCookie()
	if err != nil || len(cookies) < 1 {
		return nil, myerror.CreateError("ошибка при получении кук")
	}
	rd.Cookies = cookies
	req, errS := http.NewRequest("GET", baseUrl, nil)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при создании запроса")
	}

	for _, cookie := range rd.Cookies {
		req.AddCookie(cookie)
	}

	response, errS := rd.Client.Do(req)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при выполнении запроса")
	}
	defer response.Body.Close()

	body, errS := io.ReadAll(response.Body)
	if errS != nil {
		return nil, myerror.CreateError("ошибка при чтении ответа")
	}
	return body, nil
}

// получает расписание в json формате
// Times:[
//
//	Time номер пары текстом
//	Code номер пары числом
//	TimeTo начало пары
//	TimeFrom конец пары
//
// ]
//
// Data:[
//
//	{
//
// Time значение Times
//
//	Class:{
//		Code код пары
//		Name название пары
//		TeacherFull полное имя препода
//		Teacher Сокращение препода
//		Form очное/заочное
//	}
//
// Group:{
// Code код группы
// Name код группы
// }
//
// Room:{
// Code код комнаты
// Name название аудитории
// }
// }
// ]
func GetRaspisanie(ch chan<- struct{}, group string) ([]byte, error) {
	return getDataFromMietUrl(ch, "https://miet.ru/schedule/data?group="+group)
}

func GetRaspisanieSt(group string) ([]byte, error) {
	return getDataFromMietUrlSt("https://miet.ru/schedule/data?group=" + group)
}

func GetAllRaspisanieSt(groups []string) ([]byte, error) {
	RaspisanieChannel := make(chan []byte, len(groups))
	RaspisanieByteGroupSlice := make([]byte, 0, 10924252)
	getData, err := OnesCookieGetDataFromMietUrl()
	if err != nil {
		return nil, err
	}
	getRasp := func(getData func(string) ([]byte, error), group string) {
		bArr, err := getData(group)
		if err != nil {
			RaspisanieChannel <- nil
		} else {
			RaspisanieChannel <- bArr
		}
	}
	for _, group := range groups {
		go getRasp(getData, group)
	}
	err = nil
	errCount := 0
	RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, '[')
	for i := 0; i < cap(RaspisanieChannel); i++ {
		RaspisanieByte := <-RaspisanieChannel
		if RaspisanieByte == nil {
			errCount++
			err = errors.New("erros while request count: " + strconv.Itoa(errCount))
		}
		if i == 0 {
			RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, RaspisanieByte...)
		} else {
			RaspisanieByteGroupSlice = append(append(RaspisanieByteGroupSlice, []byte(",")...), RaspisanieByte...)
		}
	}
	RaspisanieByteGroupSlice = append(RaspisanieByteGroupSlice, ']')
	return RaspisanieByteGroupSlice, err
}

// получает список групп в формате json строкового массива
func GetGroup(ch chan<- struct{}) ([]byte, error) {
	return getDataFromMietUrl(ch, "https://miet.ru/schedule/groups")
}

func GetGroupSt() ([]byte, error) {
	return getDataFromMietUrlSt("https://miet.ru/schedule/groups")
}

// получить куки с миэта
func getCookie() ([]*http.Cookie, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	cookie := []*http.Cookie{}
	getUrl := "https://miet.ru/schedule/"

	re, err := http.Get(getUrl)
	if err != nil {
		return nil, myerror.CreateError("error while get request to get cookie from miet")
	}
	defer re.Body.Close()
	bodyBytes, err := io.ReadAll(re.Body)
	if err != nil {
		return nil, myerror.CreateError("error while read response from miet")
	}
	reg := regexp.MustCompile(`document.cookie="wl=([^;]+);`)
	matches := reg.FindAllStringSubmatch(string(bodyBytes), -1)
	if len(matches) > 0 {
		cookie = append(cookie, &http.Cookie{Name: "wl", Value: matches[0][1]})
	}

	cookieUrl, err := url.Parse(getUrl)

	if err != nil {
		return nil, myerror.CreateError("error while parsing https://miet.ru/schedule/")
	}

	jar.SetCookies(cookieUrl, cookie)

	resp, err := client.Get("https://miet.ru/schedule/")
	if err != nil {
		return nil, myerror.CreateError("error while get request to miet")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, myerror.CreateError("error, status code not 200")
	}

	cookies := client.Jar.Cookies(resp.Request.URL)
	return cookies, nil
}
