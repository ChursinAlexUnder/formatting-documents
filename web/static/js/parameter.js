document.addEventListener('DOMContentLoaded', function () {
    const buttons = document.querySelectorAll('.window-one .window-one-button'); // указываем и класс первого окна, чтобы сюда не попали кнопки из второго окна
    const windowTwo = document.querySelector('.window-two'); // Получение второго окна


    buttons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.preventDefault(); // Отмена стандартного действия кнопки (перехода по ссылке)

            windowTwo.innerHTML = '<img id="window-preloader" class="window-preloader" src="../static/pictures/gear-darkblue.svg">'; // Очистка второго окна
            // Показываем прелоадер
            document.getElementById('window-preloader').style.display = 'block';
            
            fetch(this.href) // Отправка AJAX-запроса на сервер
                .then(response => response.json()) // Преобразование ответа сервера в JSON
                .then(data => {
                    document.getElementById('window-preloader').style.display = 'none';
                    if (data.options && data.options.length > 0) {
                        data.options.forEach(option => {
                            const button = document.createElement('a');
                            button.className = 'window-two-button';
                            button.href = '#form'; // Для демонстрации оставляем пустым
                            button.textContent = option;
                            windowTwo.appendChild(button);
                        });
                    } else {
                        windowTwo.innerHTML = '<p>Ничего нет</p>';
                    }
                })
                .catch(error => {
                    document.getElementById('window-preloader').style.display = 'none';
                    windowTwo.innerHTML = `<p>Произошла ошибка: ${error.message}</p>`; // Передача текста ошибки в тег <p>
                })
        });
    });
});