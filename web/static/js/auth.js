
let registerTurnstileWidgetId = null;
let loginTurnstileWidgetId = null;

let registerTurnstileToken = '';
let loginTurnstileToken = '';

let turnstileConfigPromise = null;
const MODAL_CLOSE_DURATION = 260;

function openAnimatedModal(modal) {
    if (!modal) return;
    if (modal._closeTimer) {
        clearTimeout(modal._closeTimer);
        modal._closeTimer = null;
    }
    modal.classList.remove('closing');
    modal.classList.add('active');
}

function closeAnimatedModal(modal, afterClose) {
    if (!modal || !modal.classList.contains('active') || modal.classList.contains('closing')) {
        return;
    }

    modal.classList.add('closing');
    modal._closeTimer = setTimeout(() => {
        modal.classList.remove('active', 'closing');
        modal._closeTimer = null;
        if (afterClose) afterClose();
    }, MODAL_CLOSE_DURATION);
}

function loadTurnstileConfig() {
    if (!turnstileConfigPromise) {
        turnstileConfigPromise = fetch('/api/config/turnstile', {
            credentials: 'same-origin',
            cache: 'no-store'
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Настройка капчи недоступна');
                }
                return response.json();
            })
            .then(data => {
                if (!data.success || !data.site_key) {
                    throw new Error('Не найден публичный ключ капчи');
                }
                return data;
            })
            .catch(error => {
                turnstileConfigPromise = null;
                throw error;
            });
    }
    return turnstileConfigPromise;
}

function waitForTurnstile(timeoutMs = 10000) {
    return new Promise((resolve, reject) => {
        const startedAt = Date.now();
        const check = () => {
            if (window.turnstile) {
                resolve(window.turnstile);
                return;
            }
            if (Date.now() - startedAt >= timeoutMs) {
                reject(new Error('Скрипт капчи не загрузился'));
                return;
            }
            setTimeout(check, 50);
        };
        check();
    });
}

async function renderRegisterTurnstile() {
    if (registerTurnstileWidgetId !== null) return;

    const [config, turnstileApi] = await Promise.all([
        loadTurnstileConfig(),
        waitForTurnstile()
    ]);

    if (registerTurnstileWidgetId !== null) return;
    registerTurnstileWidgetId = turnstileApi.render('#registerTurnstile', {
        sitekey: config.site_key,
        action: 'register',
        callback: function(token) {
            registerTurnstileToken = token;
        },
        'expired-callback': function() {
            registerTurnstileToken = '';
        },
        'error-callback': function() {
            registerTurnstileToken = '';
        }
    });
}

async function renderLoginTurnstile() {
    if (loginTurnstileWidgetId !== null) return;

    const [config, turnstileApi] = await Promise.all([
        loadTurnstileConfig(),
        waitForTurnstile()
    ]);

    if (loginTurnstileWidgetId !== null) return;
    loginTurnstileWidgetId = turnstileApi.render('#loginTurnstile', {
        sitekey: config.site_key,
        action: 'login',
        callback: function(token) {
            loginTurnstileToken = token;
        },
        'expired-callback': function() {
            loginTurnstileToken = '';
        },
        'error-callback': function() {
            loginTurnstileToken = '';
        }
    });
}

function openRegisterModal() {
    const modal = document.getElementById('registerModal');
    openAnimatedModal(modal);
    renderRegisterTurnstile().catch(error => {
        console.error('Ошибка Turnstile:', error);
        showNotification('Не удалось загрузить капчу', 'error');
    });
}

function closeRegisterModal(event) {
    if (event && event.target.id !== 'registerModal') return;
    const modal = document.getElementById('registerModal');

    registerTurnstileToken = '';
    if (registerTurnstileWidgetId !== null && window.turnstile) {
        window.turnstile.reset(registerTurnstileWidgetId);
    }
    closeAnimatedModal(modal);
}

function openLoginModal() {
    const modal = document.getElementById('loginModal');
    openAnimatedModal(modal);
    renderLoginTurnstile().catch(error => {
        console.error('Ошибка Turnstile:', error);
        showNotification('Не удалось загрузить капчу', 'error');
    });
}

function closeLoginModal(event) {
    if (event && event.target.id !== 'loginModal') return;
    const modal = document.getElementById('loginModal');

    loginTurnstileToken = '';
    if (loginTurnstileWidgetId !== null && window.turnstile) {
        window.turnstile.reset(loginTurnstileWidgetId);
    }
    closeAnimatedModal(modal);
}

function handleRegister(event) {
    event.preventDefault();

    const login = document.getElementById('registerLogin').value;
    const password = document.getElementById('registerPassword').value;

    if (!registerTurnstileToken) {
        showNotification('Подтвердите капчу', 'error');
        return;
    }

    fetch('/api/auth/register', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            login,
            password,
            'cf-turnstile-response': registerTurnstileToken
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Профиль успешно создан!', 'success');
            closeRegisterModal();
            document.getElementById('registerForm').reset();
            registerTurnstileToken = '';

            if (registerTurnstileWidgetId !== null && window.turnstile) {
                window.turnstile.reset(registerTurnstileWidgetId);
            }

            updateAuthUI();
            loadSelectedTemplate();
            setTimeout(() => {
                window.location.href = '/profile';
            }, 1000);
        } else {
            showNotification(data.message || 'Ошибка регистрации', 'error');

            registerTurnstileToken = '';
            if (registerTurnstileWidgetId !== null && window.turnstile) {
                window.turnstile.reset(registerTurnstileWidgetId);
            }
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showNotification('Ошибка при регистрации', 'error');

        registerTurnstileToken = '';
        if (registerTurnstileWidgetId !== null && window.turnstile) {
            window.turnstile.reset(registerTurnstileWidgetId);
        }
    });
}

function handleLogin(event) {
    event.preventDefault();

    const login = document.getElementById('loginLogin').value;
    const password = document.getElementById('loginPassword').value;

    if (!loginTurnstileToken) {
        showNotification('Подтвердите капчу', 'error');
        return;
    }

    fetch('/api/auth/login', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            login,
            password,
            'cf-turnstile-response': loginTurnstileToken
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Вход выполнен успешно!', 'success');
            closeLoginModal();
            document.getElementById('loginForm').reset();
            loginTurnstileToken = '';

            if (loginTurnstileWidgetId !== null && window.turnstile) {
                window.turnstile.reset(loginTurnstileWidgetId);
            }

            updateAuthUI();
            loadSelectedTemplate();
            setTimeout(() => {
                window.location.href = '/profile';
            }, 1000);
        } else {
            showNotification(data.message || 'Ошибка входа', 'error');

            loginTurnstileToken = '';
            if (loginTurnstileWidgetId !== null && window.turnstile) {
                window.turnstile.reset(loginTurnstileWidgetId);
            }
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showNotification('Ошибка при входе', 'error');

        loginTurnstileToken = '';
        if (loginTurnstileWidgetId !== null && window.turnstile) {
            window.turnstile.reset(loginTurnstileWidgetId);
        }
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
        console.error('Ошибка:', error);
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
                const selector = document.getElementById('templateSelector');
                if (selector) selector.style.display = 'none';
            }
        })
        .catch(error => {
            console.error('Ошибка:', error);
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
                    applyTemplateToForm(data.template);
                }
            }
        })
        .catch(error => console.error('Ошибка:', error));
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
            resetFormToDefaults();
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
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
document.addEventListener('DOMContentLoaded', function() {
    updateAuthUI();
});
