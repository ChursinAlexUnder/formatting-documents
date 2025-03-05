// 
// ИЗУЧИТЬ ЭТОТ КОД!!!
// 
let eventSource;  // Делаем переменную глобальной для управления соединением

function connectSSE() {
    // Создаем новое соединение
    eventSource = new EventSource("/events");

    // Обработчик входящих сообщений
    eventSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            document.getElementById('counter').textContent = "Счётчик: " + data.count;
            
            const slider = document.getElementById('slider');
            slider.innerHTML = "";
            
            data.last_formatting.forEach(item => {
                const div = document.createElement("div");
                div.classList.add("slider-item");
                div.innerHTML = `
                    <strong>Font:</strong> ${item.font}<br>
                    <strong>Fontsize:</strong> ${item.fontsize}<br>
                    <strong>Alignment:</strong> ${item.alignment}<br>
                    <strong>Spacing:</strong> ${item.spacing}<br>
                    <strong>BeforeSpacing:</strong> ${item.beforeSpacing}<br>
                    <strong>AfterSpacing:</strong> ${item.afterSpacing}<br>
                    <strong>FirstIndentation:</strong> ${item.firstIndentation}<br>
                    <strong>ListTabulation:</strong> ${item.listTabulation}
                `;
                slider.appendChild(div);
            });
        } catch (error) {
            console.error("Ошибка при обработке данных:", error);
        }
    };

    // Обработчик ошибок с автоматическим переподключением
    eventSource.onerror = function(err) {
        console.log("Ошибка соединения. Пытаемся переподключиться...");
        eventSource.close();
        setTimeout(connectSSE, 3000);  // Повторная попытка через 3 секунды
    };
}

// Запускаем соединение при загрузке страницы
window.addEventListener('load', function() {
    connectSSE();
});

// Закрываем соединение при закрытии/обновлении страницы
window.addEventListener('beforeunload', function() {
    if (eventSource) {
        eventSource.close();
        console.log("Соединение закрыто");
    }
});