package services

import (
	"encoding/json"
	"formatting-documents/internal/domain"
	"os"
)

// добавление номера нового активного пользователя и устанока текущего
func AddUserNumber() error {
	var (
		usersByte   []byte
		usersStruct domain.Users
		err         error
	)
	usersByte, err = os.ReadFile("../buffer/database.json")
	if err != nil {
		return err
	}

	if err = json.Unmarshal(usersByte, &usersStruct); err != nil {
		return err
	}

	for i := 0; i < len(usersStruct.Numbers); i++ {
		if usersStruct.Numbers[i] == 0 {
			usersStruct.Numbers[i] = i + 1
			domain.User = i + 1
			break
		}
	}
	if domain.User == 0 {
		usersStruct.Numbers = append(usersStruct.Numbers, len(usersStruct.Numbers)+1)
		domain.User = len(usersStruct.Numbers)
	}

	usersByte, err = json.Marshal(usersStruct)
	if err != nil {
		return err
	}

	if err = os.WriteFile("../buffer/database.json", usersByte, 0600); err != nil {
		return err
	}
	return nil
}

// удаление пользователя из активных
func DeleteUserNumber() error {
	var (
		usersByte   []byte
		usersStruct domain.Users
		err         error
	)
	usersByte, err = os.ReadFile("../buffer/database.json")
	if err != nil {
		return err
	}

	if err = json.Unmarshal(usersByte, &usersStruct); err != nil {
		return err
	}

	// удаление пользователя из активных
	usersStruct.Numbers[domain.User-1] = 0
	domain.User = 0

	usersByte, err = json.Marshal(usersStruct)
	if err != nil {
		return err
	}

	if err = os.WriteFile("../buffer/database.json", usersByte, 0600); err != nil {
		return err
	}
	return nil
}
