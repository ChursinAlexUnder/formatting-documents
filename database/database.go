package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"formatting-documents/internal/domain"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

const queryTimeout = 3 * time.Second

//go:embed schema.sql
var schemaSQL string

func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		_ = DB.Close()
		DB = nil
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if _, err = DB.Exec(schemaSQL); err != nil {
		_ = DB.Close()
		DB = nil
		return fmt.Errorf("не удалось применить схему базы данных: %w", err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(10 * time.Minute)

	return nil
}

func IsAvailable() bool {
	return DB != nil
}

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}
func CreateUser(login, password string) (*domain.User, error) {
	if len(login) > 100 || len(password) > 100 {
		return nil, fmt.Errorf("логин или пароль слишком длинный")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("не удалось зашифровать пароль: %w", err)
	}

	var userID int64
	ctx, cancel := withTimeout()
	defer cancel()

	err = DB.QueryRowContext(
		ctx,
		"INSERT INTO profiles (login, password_hash) VALUES ($1, $2) RETURNING id",
		login, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) && pqError.Code == "23505" {
			return nil, fmt.Errorf("пользователь с таким логином уже существует")
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return &domain.User{ID: userID, Login: login}, nil
}
func GetUserByLogin(login string) (*domain.User, error) {
	var user domain.User
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT id, login FROM profiles WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return &user, nil
}
func GetUserByID(id int64) (*domain.User, error) {
	var user domain.User
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT id, login FROM profiles WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Login)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return &user, nil
}
func VerifyPassword(login, password string) (int64, error) {
	var userID int64
	var hashedPassword string

	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT id, password_hash FROM profiles WHERE login = $1",
		login,
	).Scan(&userID, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("пользователь не найден")
		}
		return 0, fmt.Errorf("не удалось проверить пароль: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return 0, fmt.Errorf("неверный пароль")
	}

	return userID, nil
}
func CreateTemplate(template *domain.FormattingTemplate) (*domain.FormattingTemplate, error) {
	var count int
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM formatting_templates WHERE profile_id = $1",
		template.ProfileID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить количество шаблонов: %w", err)
	}
	if count >= 50 {
		return nil, fmt.Errorf("достигнут лимит шаблонов")
	}

	err = DB.QueryRowContext(
		ctx,
		`INSERT INTO formatting_templates
		(profile_id, name, font, fontsize, alignment, spacing, before_spacing, after_spacing, first_indentation, list_tabulation, have_title)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`,
		template.ProfileID, template.Name, template.Font, template.Fontsize, template.Alignment,
		template.Spacing, template.BeforeSpacing, template.AfterSpacing, template.FirstIndentation,
		template.ListTabulation, template.HaveTitle,
	).Scan(&template.ID)

	if err != nil {
		return nil, fmt.Errorf("не удалось создать шаблон: %w", err)
	}

	return template, nil
}
func GetTemplatesByUserID(userID int64) ([]domain.FormattingTemplate, error) {
	ctx, cancel := withTimeout()
	defer cancel()

	rows, err := DB.QueryContext(
		ctx,
		`SELECT id, profile_id, name, font, fontsize, alignment, spacing, before_spacing, after_spacing, first_indentation, list_tabulation, have_title
		FROM formatting_templates WHERE profile_id = $1 ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить шаблоны: %w", err)
	}
	defer rows.Close()

	var templates []domain.FormattingTemplate
	for rows.Next() {
		var t domain.FormattingTemplate
		err := rows.Scan(
			&t.ID, &t.ProfileID, &t.Name, &t.Font, &t.Fontsize, &t.Alignment,
			&t.Spacing, &t.BeforeSpacing, &t.AfterSpacing, &t.FirstIndentation,
			&t.ListTabulation, &t.HaveTitle,
		)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать шаблон: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, nil
}
func GetTemplateByID(templateID int64) (*domain.FormattingTemplate, error) {
	var t domain.FormattingTemplate
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		`SELECT id, profile_id, name, font, fontsize, alignment, spacing, before_spacing, after_spacing, first_indentation, list_tabulation, have_title
		FROM formatting_templates WHERE id = $1`,
		templateID,
	).Scan(
		&t.ID, &t.ProfileID, &t.Name, &t.Font, &t.Fontsize, &t.Alignment,
		&t.Spacing, &t.BeforeSpacing, &t.AfterSpacing, &t.FirstIndentation,
		&t.ListTabulation, &t.HaveTitle,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("шаблон не найден")
		}
		return nil, fmt.Errorf("не удалось получить шаблон: %w", err)
	}

	return &t, nil
}
func UpdateTemplate(template *domain.FormattingTemplate) (*domain.FormattingTemplate, error) {
	ctx, cancel := withTimeout()
	defer cancel()

	_, err := DB.ExecContext(
		ctx,
		`UPDATE formatting_templates
		SET name = $1, font = $2, fontsize = $3, alignment = $4, spacing = $5,
		before_spacing = $6, after_spacing = $7, first_indentation = $8, list_tabulation = $9, have_title = $10
		WHERE id = $11 AND profile_id = $12`,
		template.Name, template.Font, template.Fontsize, template.Alignment, template.Spacing,
		template.BeforeSpacing, template.AfterSpacing, template.FirstIndentation,
		template.ListTabulation, template.HaveTitle, template.ID, template.ProfileID,
	)

	if err != nil {
		return nil, fmt.Errorf("не удалось обновить шаблон: %w", err)
	}

	return template, nil
}
func DeleteTemplate(templateID, userID int64) error {
	ctx, cancel := withTimeout()
	defer cancel()

	result, err := DB.ExecContext(
		ctx,
		"DELETE FROM formatting_templates WHERE id = $1 AND profile_id = $2",
		templateID, userID,
	)

	if err != nil {
		return fmt.Errorf("не удалось удалить шаблон: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось проверить результат удаления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("шаблон не найден")
	}

	return nil
}
