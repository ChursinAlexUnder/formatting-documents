document.addEventListener('DOMContentLoaded', function () {
    const buttons = document.querySelectorAll('.window-one .window-button'); // указываем и класс первого окна, чтобы сюда не попали кнопки из второго окна

    buttons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.preventDefault(); // Отмена стандартного действия кнопки (перехода по ссылке)

            fetch(this.href) // Отправка AJAX-запроса на сервер
                .then(response => response.json()) // Преобразование ответа сервера в JSON
                .then(data => {
                    const windowTwo = document.querySelector('.window-two'); // Получение второго окна
                    windowTwo.innerHTML = ''; // Очистка второго окна
                    if (data.options && data.options.length > 0) {
                        data.options.forEach(option => {
                            const button = document.createElement('a');
                            button.className = 'window-button';
                            button.href = '#'; // Для демонстрации оставляем пустым
                            button.textContent = option;
                            windowTwo.appendChild(button);
                        });
                    } else {
                        windowTwo.innerHTML = '<p>Ничего нет</p>';
                    }
                })
                .catch(error => {
                    const windowTwo = document.querySelector('.window-two');
                    windowTwo.innerHTML = `<p>Произошла ошибка: ${error.message}</p>`; // Передача текста ошибки в тег <p>
                });
        });
    });
});