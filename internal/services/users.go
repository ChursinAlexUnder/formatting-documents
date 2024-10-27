package services

import "formatting-documents/internal/domain"

// добавление номера нового активного пользователя и устанока текущего
func AddUserNumber() {
	for i := 0; i < len(domain.Users); i++ {
		if domain.Users[i] == 0 {
			domain.Users[i] = i + 1
			domain.User = i + 1
			break
		}
	}
	if domain.User == 0 {
		domain.Users = append(domain.Users, len(domain.Users)+1)
		domain.User = len(domain.Users)
	}
}

// удаление пользователя из активных
func DeleteUser() {
	domain.Users[domain.User-1] = 0
	domain.User = 0
}
