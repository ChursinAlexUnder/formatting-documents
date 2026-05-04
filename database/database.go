package database

import (
	"context"
	"database/sql"
	"fmt"
	"formatting-documents/internal/domain"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

const queryTimeout = 3 * time.Second

// InitDB инициализирует подключение к БД
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		_ = DB.Close()
		DB = nil
		return fmt.Errorf("failed to connect to database: %w", err)
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

// CreateUser создаёт нового пользователя
func CreateUser(login, password string) (*domain.User, error) {
	if len(login) > 100 || len(password) > 100 {
		return nil, fmt.Errorf("login or password too long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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
		if err.Error() == "pq: duplicate key value violates unique constraint \"profiles_login_key\"" {
			return nil, fmt.Errorf("login already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &domain.User{ID: userID, Login: login}, nil
}

// GetUserByLogin получает пользователя по логину
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// VerifyPassword проверяет пароль
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
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to verify password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return 0, fmt.Errorf("invalid password")
	}

	return userID, nil
}

// CreateTemplate создаёт новый шаблон
func CreateTemplate(template *domain.FormattingTemplate) (*domain.FormattingTemplate, error) {
	// Проверка лимита шаблонов
	var count int
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM formatting_templates WHERE profile_id = $1",
		template.ProfileID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check template count: %w", err)
	}
	if count >= 50 {
		return nil, fmt.Errorf("template limit exceeded")
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
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// GetTemplatesByUserID получает все шаблоны пользователя
func GetTemplatesByUserID(userID int64) ([]domain.FormattingTemplate, error) {
	ctx, cancel := withTimeout()
	defer cancel()

	rows, err := DB.QueryContext(
		ctx,
		`SELECT id, profile_id, name, font, fontsize, alignment, spacing, before_spacing, after_spacing, first_indentation, list_tabulation, have_title
		FROM formatting_templates WHERE profile_id = $1 ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
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
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// GetTemplateByID получает шаблон по ID
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
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &t, nil
}

// UpdateTemplate обновляет шаблон
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
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return template, nil
}

// DeleteTemplate удаляет шаблон
func DeleteTemplate(templateID, userID int64) error {
	ctx, cancel := withTimeout()
	defer cancel()

	result, err := DB.ExecContext(
		ctx,
		"DELETE FROM formatting_templates WHERE id = $1 AND profile_id = $2",
		templateID, userID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// SaveSelectedTemplate сохраняет выбранный шаблон в сессию (через таблицу)
func SaveSelectedTemplate(userID, templateID int64) error {
	// Проверяем, что шаблон принадлежит пользователю
	var count int
	ctx, cancel := withTimeout()
	defer cancel()

	err := DB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM formatting_templates WHERE id = $1 AND profile_id = $2",
		templateID, userID,
	).Scan(&count)

	if err != nil || count == 0 {
		return fmt.Errorf("template not found or doesn't belong to user")
	}

	// Для сейчас сохраняем в памяти или можно добавить таблицу для сессий
	return nil
}
