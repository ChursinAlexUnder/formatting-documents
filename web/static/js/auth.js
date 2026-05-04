// Auth Modal Functions

function openRegisterModal() {
    const modal = document.getElementById('registerModal');
    modal.classList.add('active');
}

function closeRegisterModal(event) {
    if (event && event.target.id !== 'registerModal') return;
    const modal = document.getElementById('registerModal');
    modal.classList.remove('active');
}

function openLoginModal() {
    const modal = document.getElementById('loginModal');
    modal.classList.add('active');
}

function closeLoginModal(event) {
    if (event && event.target.id !== 'loginModal') return;
    const modal = document.getElementById('loginModal');
    modal.classList.remove('active');
}

function handleRegister(event) {
    event.preventDefault();
    
    const login = document.getElementById('registerLogin').value;
    const password = document.getElementById('registerPassword').value;

    fetch('/api/auth/register', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ login, password })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Профиль успешно создан!', 'success');
            closeRegisterModal();
            document.getElementById('registerForm').reset();
            updateAuthUI();
            loadSelectedTemplate();
            setTimeout(() => {
                window.location.href = '/profile';
            }, 1000);
        } else {
            showNotification(data.message || 'Ошибка регистрации', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при регистрации', 'error');
    });
}

function handleLogin(event) {
    event.preventDefault();
    
    const login = document.getElementById('loginLogin').value;
    const password = document.getElementById('loginPassword').value;

    fetch('/api/auth/login', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ login, password })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Вход выполнен успешно!', 'success');
            closeLoginModal();
            document.getElementById('loginForm').reset();
            updateAuthUI();
            loadSelectedTemplate();
            setTimeout(() => {
                window.location.href = '/profile';
            }, 1000);
        } else {
            showNotification(data.message || 'Ошибка входа', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при входе', 'error');
    });
}

function logout() {
    fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'same-origin',
        cache: 'no-store'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Вы вышли из профиля', 'info');
            updateAuthUI();
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при выходе', 'error');
    });
}

function goToProfile() {
    window.location.assign('/profile');
}

function updateAuthUI() {
    fetch('/api/profile', {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            const authButtons = document.getElementById('authButtons');
            const profileButtons = document.getElementById('profileButtons');
            const resetBtn = document.getElementById('headerResetTemplateBtn');
            const loginLabel = document.getElementById('headerProfileLogin');
            
            if (Boolean(data.success) === true) {
                if (authButtons) authButtons.style.display = 'none';
                if (profileButtons) profileButtons.style.display = 'flex';
                if (loginLabel) {
                    const fullLogin = String(data.login || '');
                    loginLabel.textContent = truncateLogin(fullLogin, 20);
                    loginLabel.title = fullLogin;
                    loginLabel.style.display = 'inline-flex';
                }
                
                // Load selected template
                if (data.selected_template_id) {
                    if (resetBtn) resetBtn.style.display = 'inline-flex';
                    loadTemplateInfo(data.selected_template_id);
                } else {
                    if (resetBtn) resetBtn.style.display = 'none';
                    const selector = document.getElementById('templateSelector');
                    if (selector) selector.style.display = 'none';
                }
            } else {
                if (authButtons) authButtons.style.display = 'flex';
                if (profileButtons) profileButtons.style.display = 'none';
                if (resetBtn) resetBtn.style.display = 'none';
                if (loginLabel) {
                    loginLabel.style.display = 'none';
                    loginLabel.textContent = '';
                    loginLabel.title = '';
                }
                
                // Hide template selector
                const selector = document.getElementById('templateSelector');
                if (selector) selector.style.display = 'none';
            }
        })
        .catch(error => {
            console.error('Error:', error);
            const authButtons = document.getElementById('authButtons');
            const profileButtons = document.getElementById('profileButtons');
            const resetBtn = document.getElementById('headerResetTemplateBtn');
            const loginLabel = document.getElementById('headerProfileLogin');
            if (authButtons) authButtons.style.display = 'flex';
            if (profileButtons) profileButtons.style.display = 'none';
            if (resetBtn) resetBtn.style.display = 'none';
            if (loginLabel) loginLabel.style.display = 'none';
        });
}

function loadSelectedTemplate() {
    const selectedCookie = document.cookie
        .split('; ')
        .find(row => row.startsWith('selected_template='));
    
    if (selectedCookie) {
        const templateId = selectedCookie.split('=')[1];
        loadTemplateInfo(parseInt(templateId));
    }
}

function loadTemplateInfo(templateId) {
    fetch(`/api/templates/get?id=${templateId}`, {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success && data.template) {
                const selector = document.getElementById('templateSelector');
                const templateName = document.getElementById('templateName');
                
                if (selector && templateName) {
                    selector.style.display = 'flex';
                    templateName.textContent = data.template.name;
                    
                    // Apply template parameters to form
                    applyTemplateToForm(data.template);
                }
            }
        })
        .catch(error => console.error('Error:', error));
}

function applyTemplateToForm(template) {
    if (document.querySelector('input[name="havetitle"]')) {
        document.querySelector('input[name="havetitle"]').value = template.haveTitle;
    }
    if (document.querySelector('input[name="font"]')) {
        document.querySelector('input[name="font"]').value = template.font;
    }
    if (document.querySelector('input[name="fontsize"]')) {
        document.querySelector('input[name="fontsize"]').value = template.fontsize;
    }
    if (document.querySelector('input[name="alignment"]')) {
        document.querySelector('input[name="alignment"]').value = template.alignment;
    }
    if (document.querySelector('input[name="spacing"]')) {
        document.querySelector('input[name="spacing"]').value = normalizeTemplateValue(template.spacing, ['1.0', '1.5', '2.0', '2.5', '3.0']);
    }
    if (document.querySelector('input[name="beforespacing"]')) {
        document.querySelector('input[name="beforespacing"]').value = normalizeTemplateValue(template.beforeSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0']);
    }
    if (document.querySelector('input[name="afterspacing"]')) {
        document.querySelector('input[name="afterspacing"]').value = normalizeTemplateValue(template.afterSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0']);
    }
    if (document.querySelector('input[name="firstindentation"]')) {
        document.querySelector('input[name="firstindentation"]').value = normalizeTemplateValue(template.firstIndentation, ['0', '0.5', '1.0', '1.25', '1.5', '1.75', '2.0', '2.5', '3.0']);
    }
    if (document.querySelector('input[name="listtabulation"]')) {
        document.querySelector('input[name="listtabulation"]').value = normalizeTemplateValue(template.listTabulation, ['0', '0.25', '0.5', '0.75', '1.0', '1.25', '1.5', '1.75', '2.0', '2.25', '2.5', '2.75', '3.0', '3.25', '3.5', '3.75', '4.0']);
    }
}

function normalizeTemplateValue(rawValue, allowedValues) {
    const numericValue = Number(rawValue);
    const match = allowedValues.find((value) => Math.abs(Number(value) - numericValue) < 0.00001);
    return match || String(rawValue);
}

function truncateLogin(login, maxLength) {
    if (login.length <= maxLength) {
        return login;
    }
    return `${login.slice(0, maxLength)}...`;
}

function resetTemplate() {
    fetch('/api/templates/reset', {
        method: 'POST',
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Шаблон сброшен', 'info');
            const selector = document.getElementById('templateSelector');
            const resetBtn = document.getElementById('headerResetTemplateBtn');
            if (selector) selector.style.display = 'none';
            if (resetBtn) resetBtn.style.display = 'none';
            
            // Reset form to default values
            resetFormToDefaults();
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при сбросе шаблона', 'error');
    });
}

function resetFormToDefaults() {
    const defaults = {
        'havetitle': 'Есть',
        'font': 'Times New Roman',
        'fontsize': '14',
        'alignment': 'По ширине',
        'spacing': '1.5',
        'beforespacing': '0',
        'afterspacing': '0',
        'firstindentation': '1.25',
        'listtabulation': '2.0'
    };
    
    for (const [name, value] of Object.entries(defaults)) {
        const input = document.querySelector(`input[name="${name}"]`);
        if (input) input.value = value;
    }
}

// Initialize auth UI on page load
document.addEventListener('DOMContentLoaded', function() {
    updateAuthUI();
});
