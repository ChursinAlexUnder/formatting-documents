// Profile Management

let currentEditingTemplateId = null;
let currentDeletingTemplateId = null;

// Load profile and templates on page load
document.addEventListener('DOMContentLoaded', function() {
    loadProfile();
});

function loadProfile() {
    fetch('/api/profile', {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                renderTemplates(data.templates || [], data.selected_template_id);
            } else {
                window.location.href = '/';
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Ошибка при загрузке профиля', 'error');
        });
}

function renderTemplates(templates, selectedId) {
    const grid = document.getElementById('templatesGrid');
    grid.innerHTML = '';

    // Add existing templates
    if (templates.length > 0) {
        templates.forEach((template) => {
            const card = document.createElement('div');
            card.className = `template-card${selectedId === template.id ? ' selected' : ''}`;

            const buttonsHtml = `
                <div class="template-card-buttons">
                    <button class="card-btn" onclick="viewTemplate(${template.id})">Просмотр</button>
                    <button class="card-btn select-btn" onclick="selectTemplate(${template.id})">Выбрать</button>
                </div>
            `;

            card.innerHTML = `
                <div class="template-card-header">
                    <h3 class="template-card-title">${escapeHtml(template.name)}</h3>
                    <div class="template-card-actions">
                        <button class="icon-btn" title="Редактировать" onclick="editTemplate(${template.id})">✎</button>
                        <button class="icon-btn delete" title="Удалить" onclick="deleteTemplate(${template.id})">🗑</button>
                    </div>
                </div>
                ${buttonsHtml}
            `;
            
            grid.appendChild(card);
        });
    }

    // Add new template card at the end
    const addCard = document.createElement('div');
    addCard.className = 'template-card add-card';
    addCard.onclick = openNewTemplateForm;
    addCard.innerHTML = `
        <div class="add-card-content">
            <div class="add-card-plus">+</div>
            <div class="add-card-text">Добавить</div>
        </div>
    `;
    grid.appendChild(addCard);
}

function openNewTemplateForm() {
    currentEditingTemplateId = null;
    document.getElementById('formTitle').textContent = 'Новый шаблон';
    document.getElementById('templateForm').reset();
    document.getElementById('templateName').focus();
    
    // Set default values
    document.getElementById('templateFont').value = 'Times New Roman';
    document.getElementById('templateFontsize').value = '14';
    document.getElementById('templateAlignment').value = 'По ширине';
    document.getElementById('templateSpacing').value = '1.5';
    document.getElementById('templateBeforeSpacing').value = '0';
    document.getElementById('templateAfterSpacing').value = '0';
    document.getElementById('templateFirstIndentation').value = '1.25';
    document.getElementById('templateListTabulation').value = '2.0';
    document.getElementById('templateHaveTitle').value = 'Есть';
    
    openTemplateModal();
}

function editTemplate(templateId) {
    fetch(`/api/templates/get?id=${templateId}`, {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success && data.template) {
                const template = data.template;
                currentEditingTemplateId = templateId;
                
                document.getElementById('formTitle').textContent = 'Редактировать шаблон';
                document.getElementById('templateName').value = template.name;
                document.getElementById('templateFont').value = template.font;
                document.getElementById('templateFontsize').value = template.fontsize;
                document.getElementById('templateAlignment').value = template.alignment;
                document.getElementById('templateSpacing').value = normalizeSelectValue(template.spacing, ['1.0', '1.5', '2.0', '2.5', '3.0']);
                document.getElementById('templateBeforeSpacing').value = normalizeSelectValue(template.beforeSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0']);
                document.getElementById('templateAfterSpacing').value = normalizeSelectValue(template.afterSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0']);
                document.getElementById('templateFirstIndentation').value = normalizeSelectValue(template.firstIndentation, ['0', '0.5', '1.0', '1.25', '1.5', '1.75', '2.0', '2.5', '3.0']);
                document.getElementById('templateListTabulation').value = normalizeSelectValue(template.listTabulation, ['0', '0.25', '0.5', '0.75', '1.0', '1.25', '1.5', '1.75', '2.0', '2.25', '2.5', '2.75', '3.0', '3.25', '3.5', '3.75', '4.0']);
                document.getElementById('templateHaveTitle').value = template.haveTitle;
                
                openTemplateModal();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Ошибка при загрузке шаблона', 'error');
        });
}

function viewTemplate(templateId) {
    fetch(`/api/templates/get?id=${templateId}`, {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success && data.template) {
                const template = data.template;
                const content = document.getElementById('templateViewContent');
                
                content.innerHTML = `
                    <div class="template-view">
                        <div class="view-param">
                            <span class="view-param-label">Название</span>
                            <span class="view-param-value">${escapeHtml(template.name)}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Шрифт</span>
                            <span class="view-param-value">${template.font}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Размер шрифта</span>
                            <span class="view-param-value">${template.fontsize}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Выравнивание</span>
                            <span class="view-param-value">${template.alignment}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Междустрочный интервал</span>
                            <span class="view-param-value">${normalizeSelectValue(template.spacing, ['1.0', '1.5', '2.0', '2.5', '3.0'])}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Интервал перед абзацем</span>
                            <span class="view-param-value">${normalizeSelectValue(template.beforeSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0'])}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Интервал после абзаца</span>
                            <span class="view-param-value">${normalizeSelectValue(template.afterSpacing, ['0', '1.0', '1.5', '2.0', '2.5', '3.0'])}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Отступ первой строки</span>
                            <span class="view-param-value">${normalizeSelectValue(template.firstIndentation, ['0', '0.5', '1.0', '1.25', '1.5', '1.75', '2.0', '2.5', '3.0'])}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Табуляция в списках</span>
                            <span class="view-param-value">${normalizeSelectValue(template.listTabulation, ['0', '0.25', '0.5', '0.75', '1.0', '1.25', '1.5', '1.75', '2.0', '2.25', '2.5', '2.75', '3.0', '3.25', '3.5', '3.75', '4.0'])}</span>
                        </div>
                        <div class="view-param">
                            <span class="view-param-label">Титульный лист</span>
                            <span class="view-param-value">${template.haveTitle}</span>
                        </div>
                    </div>
                `;
                
                openViewModal();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Ошибка при загрузке шаблона', 'error');
        });
}

function handleSaveTemplate(event) {
    event.preventDefault();
    
    const templateData = {
        id: currentEditingTemplateId,
        name: document.getElementById('templateName').value,
        font: document.getElementById('templateFont').value,
        fontsize: parseInt(document.getElementById('templateFontsize').value),
        alignment: document.getElementById('templateAlignment').value,
        spacing: parseFloat(document.getElementById('templateSpacing').value),
        beforeSpacing: parseFloat(document.getElementById('templateBeforeSpacing').value),
        afterSpacing: parseFloat(document.getElementById('templateAfterSpacing').value),
        firstIndentation: parseFloat(document.getElementById('templateFirstIndentation').value),
        listTabulation: parseFloat(document.getElementById('templateListTabulation').value),
        haveTitle: document.getElementById('templateHaveTitle').value
    };

    const endpoint = currentEditingTemplateId 
        ? '/api/templates/update' 
        : '/api/templates/create';

    fetch(endpoint, {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(templateData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            const message = currentEditingTemplateId 
                ? 'Шаблон обновлен' 
                : 'Шаблон создан';
            showNotification(message, 'success');
            closeTemplateModal();
            loadProfile();
        } else {
            showNotification(data.message || 'Ошибка при сохранении', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при сохранении шаблона', 'error');
    });
}

function deleteTemplate(templateId) {
    fetch(`/api/templates/get?id=${templateId}`, {
        credentials: 'same-origin',
        cache: 'no-store'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success && data.template) {
                currentDeletingTemplateId = templateId;
                document.getElementById('deleteTemplateName').textContent = escapeHtml(data.template.name);
                openDeleteModal();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Ошибка при загрузке шаблона', 'error');
        });
}

function confirmDelete() {
    if (!currentDeletingTemplateId) return;

    fetch(`/api/templates/delete?id=${currentDeletingTemplateId}`, {
        method: 'POST',
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Шаблон удален', 'success');
            closeDeleteModal();
            loadProfile();
        } else {
            showNotification(data.message || 'Ошибка при удалении', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при удалении шаблона', 'error');
    });
}

function selectTemplate(templateId) {
    fetch(`/api/templates/select?id=${templateId}`, {
        method: 'POST',
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification(`Шаблон "${data.template.name}" выбран`, 'success');
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        } else {
            showNotification(data.message || 'Ошибка при выборе шаблона', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Ошибка при выборе шаблона', 'error');
    });
}

// Modal Management

function openTemplateModal() {
    const modal = document.getElementById('templateModal');
    modal.classList.add('active');
}

function closeTemplateModal(event) {
    if (event && event.target.id !== 'templateModal') return;
    const modal = document.getElementById('templateModal');
    modal.classList.remove('active');
}

function openViewModal() {
    const modal = document.getElementById('viewModal');
    modal.classList.add('active');
}

function closeViewModal(event) {
    if (event && event.target.id !== 'viewModal') return;
    const modal = document.getElementById('viewModal');
    modal.classList.remove('active');
}

function openDeleteModal() {
    const modal = document.getElementById('deleteModal');
    modal.classList.add('active');
}

function closeDeleteModal(event) {
    if (event && event.target.id !== 'deleteModal') return;
    const modal = document.getElementById('deleteModal');
    modal.classList.remove('active');
    currentDeletingTemplateId = null;
}

// Utility function
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function normalizeSelectValue(rawValue, allowedValues) {
    const numericValue = Number(rawValue);
    const match = allowedValues.find((value) => Math.abs(Number(value) - numericValue) < 0.00001);
    return match || String(rawValue);
}
