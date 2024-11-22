document.addEventListener('DOMContentLoaded', function () {
    const buttons = document.querySelectorAll('.window-one .window-one-button'); // указываем и класс первого окна, чтобы сюда не попали кнопки из второго окна
    const windowTwo = document.querySelector('.window-two'); // Получение второго окна
    const parametersContainer = document.querySelector('.parameters-container'); // Контейнер параметров


    buttons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.preventDefault(); // Отмена стандартного действия кнопки (перехода по ссылке)

            // Получаем значение параметра из ссылки
            const url = new URL(this.href);
            const parameterName = url.searchParams.get('parameter');

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
                            button.href = '/'; // Оставляем пустым
                            button.textContent = option;
                            button.setAttribute('data-target', parameterName); // Добавляем параметр из ссылки
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
    // Обработчик нажатия кнопок во втором окне
    windowTwo.addEventListener('click', function (event) {
        if (event.target.classList.contains('window-two-button')) {
            event.preventDefault();
            const targetInput = event.target.getAttribute('data-target'); // Получаем целевой input
            const inputElement = parametersContainer.querySelector(`.parameter-input[name="${targetInput}"]`); // Ищем input по name
            if (inputElement) {
                inputElement.value = event.target.textContent; // Меняем value на текст кнопки
            }
            //
            // Анимация
            //
            const elementsToAnimate = [
                document.querySelector('.window-one'), // Первое окно
                document.querySelector('.window-two'), // Второе окно
                document.querySelector('.parameters'), // Контейнер параметров
            ];

            elementsToAnimate.forEach(element => {
                // Добавляем класс анимации
                element.classList.add('animate-shadow');

                // Удаляем класс после завершения анимации
                element.addEventListener('animationend', function handleAnimationEnd() {
                    element.classList.remove('animate-shadow');
                    element.removeEventListener('animationend', handleAnimationEnd); // Удаляем обработчик
                });
            });
        }
    });
});