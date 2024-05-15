package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// save file function
func Save(a *fyne.App, data []byte, fileName string) (bool, string, error) {
	if exist, filePath, err := IsFileExist(a, fileName); err != nil {
		//Ошибка поиска файла
		return false, "", err
	} else if exist {
		fileContent, err := readFile(filePath)
		if err != nil {
			//Ошибка чтения файла
			return false, "", err
		}
		if calculateHash(data) != calculateHash(fileContent) {
			err = writeFile(filePath, data)
			if err != nil {
				fmt.Println("Ошибка при перезаписи файла:", err)
				// ошибка при записи
				return false, "", err
			}
			//хэши не совпали, файл перезаписан
			return true, filePath, nil
		}
		// если хэши совпадают, обновление не нужно
		return false, filePath, nil
	} else {
		createAndFillFile(filePath, data)
		//файла не существовало, он создан и заполнен
		return true, filePath, nil
	}
}

// Запись в файл
// TODO: сделать компрессию через gzip
func writeFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0644)
}

// Создаёт и заполняет файл содержимым
func createAndFillFile(filePath string, content []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		//ощибка при создании файла
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		//ошибка при записи в файл
		return err
	}
	return nil

}

// считает хэш у байтового среза
func calculateHash(data []byte) string {
	// Создаем новый объект хеша SHA-256
	hash := sha256.New()

	// Записываем данные в хеш
	hash.Write(data)

	// Получаем байты хеша
	hashBytes := hash.Sum(nil)

	// Преобразуем байты хеша в строку
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}

// чтение файла по указанному пути
func readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func GetFile(a *fyne.App, fileName string) (*os.File, error) {
	if exist, path, err := IsFileExist(a, fileName); err != nil {
		return nil, err
	} else if exist {
		return os.OpenFile(path, os.O_RDWR, 0666)
	}
	return nil, nil
}

// Существует ли filename
// формат filename data.json
// возвращает да/нет, путь до data, ошибку
func IsFileExist(a *fyne.App, fileName string) (bool, string, error) {
	if ok, dataDir := isStorageExist(a); ok {
		filePath := filepath.Join(dataDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// не существует файла
			return false, filePath, nil
		} else if err != nil {
			// ошибка получения файла
			return false, filePath, errors.New("error while get " + fileName)
		} else {
			// файл существует
			return true, filePath, nil
		}
	}
	// не существует хранилища
	return false, "", errors.New("storage not exist")
}

// Проверка возможности писать в хранилище
func isStorageExist(a *fyne.App) (isExist bool, dataDir string) {
	dataDir = (*a).Storage().RootURI().Path()
	if dataDir == "" {
		return false, ""
	}
	return true, dataDir
}
