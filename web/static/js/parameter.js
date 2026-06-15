document.addEventListener('DOMContentLoaded', function () {
    const buttons = document.querySelectorAll('.window-one .window-one-button');
    const windowTwo = document.querySelector('.window-two');
    const parametersContainer = document.querySelector('.parameters-container');
    let currentController = null;

    buttons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.preventDefault();
            if (currentController) {
                currentController.abort();
            }
            currentController = new AbortController();
            const signal = currentController.signal;
            const url = new URL(this.href);
            const parameterName = url.searchParams.get('parameter');

            windowTwo.innerHTML = '<img id="window-preloader" class="window-preloader" src="../static/pictures/gear-darkblue.svg">';
            document.getElementById('window-preloader').style.display = 'block';

            fetch(this.href, { signal })
                .then(response => response.json())
                .then(data => {
                    document.getElementById('window-preloader').style.display = 'none';
                    if (data.options && data.options.length > 0) {
                        data.options.forEach(option => {
                            const button = document.createElement('a');
                            button.className = 'window-two-button';
                            button.href = '/';
                            button.textContent = option;
                            button.setAttribute('data-target', parameterName);
                            windowTwo.appendChild(button);
                        });
                    } else {
                        windowTwo.innerHTML = '<p>Ничего нет</p>';
                    }
                })
                .catch(error => {
                    if (error.name !== 'AbortError') {
                        document.getElementById('window-preloader').style.display = 'none';
                        windowTwo.innerHTML = `<p>Произошла ошибка: ${error.message}</p>`;
                    }
                })
        });
    });
    windowTwo.addEventListener('click', function (event) {
        if (event.target.classList.contains('window-two-button')) {
            event.preventDefault();
            const targetInput = event.target.getAttribute('data-target');
            const inputElement = parametersContainer.querySelector(`.parameter-input[name="${targetInput}"]`);
            if (inputElement) {
                inputElement.value = event.target.textContent;
            }
            const elementsToAnimate = [
                document.querySelector('.window-one'),
                document.querySelector('.window-two'),
                document.querySelector('.parameters'),
            ];

            elementsToAnimate.forEach(element => {
                element.classList.add('animate-shadow');
                element.addEventListener('animationend', function handleAnimationEnd() {
                    element.classList.remove('animate-shadow');
                    element.removeEventListener('animationend', handleAnimationEnd);
                });
            });
        }
    });
});
