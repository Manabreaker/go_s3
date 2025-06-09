package filestore

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

type FileStore struct {
	Files     *sql.DB
	StorePath string
}

func New(files *sql.DB, storePath string) *FileStore {
	return &FileStore{
		Files:     files,
		StorePath: storePath,
	}
}

func (f *FileStore) FindFilesWithDates(id int) ([]string, []int, error) {
	rows, err := f.Files.Query("SELECT filename, uploaded_at FROM files WHERE userid = $1;", id)
	if err != nil {
		return nil, nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var fileNames []string
	var fileUploadTimes []int

	for rows.Next() {
		var name string
		var uploadedAt time.Time
		if err := rows.Scan(&name, &uploadedAt); err != nil {
			return nil, nil, err
		}
		fileNames = append(fileNames, name)
		fileUploadTimes = append(fileUploadTimes, int(uploadedAt.Unix()-10800)) // Преобразование времени в Unix timestamp с учетом часового пояса
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка после итерации: %v", err)
		return nil, nil, err
	}

	return fileNames, fileUploadTimes, nil
}

func (f *FileStore) FindFiles(id int) ([]string, error) {
	rows, err := f.Files.Query("SELECT filename FROM files WHERE userid = $1;", id)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var fileNames []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		fileNames = append(fileNames, name)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка после итерации: %v", err)
		return nil, err
	}

	return fileNames, nil
}

func (f *FileStore) Save(userID int, filename string, fileBytes []byte) error {
	tx, err := f.Files.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO files (userid, filename) VALUES ($1, $2)",
		userID,
		filename,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	dirPath := fmt.Sprintf("%s/%d", f.StorePath, userID)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(fmt.Sprintf("%s/%s", dirPath, filename), fileBytes, 0644); err != nil {
		return err
	}
	return nil
}

func (f *FileStore) GetFullFileName(userID int, filename string) (string, error) {
	_, err := os.ReadFile(fmt.Sprintf("%s/%d/%s", f.StorePath, userID, filename))
	if err != nil {
		return "", err
	}
	return string(fmt.Sprintf("%s/%d/%s", f.StorePath, userID, filename)), nil
}

func (f *FileStore) GetFileBytes(userID int, filename string) ([]byte, error) {
	file, err := os.ReadFile(fmt.Sprintf("%s/%d/%s", f.StorePath, userID, filename))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileStore) Delete(userID int, filename string) error {
	tx, err := f.Files.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = f.Files.Exec("DELETE FROM files WHERE userid = $1 AND filename = $2", userID, filename)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if _, err := os.Stat(fmt.Sprintf("%d/%s", userID, filename)); err != nil {
			return err
		}
		if err := os.Remove(fmt.Sprintf("%d/%s", userID, filename)); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStore) FindByUUID(uuid string) (int, string, error) {
	var (
		userID   int
		filename string
	)
	if err := f.Files.QueryRow("SELECT userid, filename FROM files WHERE uuid = $1 and public = true LIMIT 1;", uuid).
		Scan(&userID, &filename); err != nil {
		return 0, "", err
	}

	return userID, filename, nil
}

func (f *FileStore) Share(id int, filename string) (string, error) {
	var Uuid string
	err := f.Files.QueryRow("UPDATE files SET public = true WHERE userid = $1 AND filename = $2 RETURNING uuid", id, filename).Scan(&Uuid)
	if err != nil {
		return "", err
	}
	return Uuid, nil
}
