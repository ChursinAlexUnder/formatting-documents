// Глобальные переменные (для каждого пользователя они)
let eventSource; 
let = sliderItems = [];
const breakpoint = 768;
let isMobile = window.innerWidth < breakpoint;
let currentSlide = 0;
let isInit = false;
let prevCount = 0;
let newHighlightCount = 0;

function buttonControl() {
    const prevButton = document.querySelector('.slider-prev');
    const nextButton = document.querySelector('.slider-next');

    if (currentSlide !== 0 && sliderItems.length !== 0) {
        prevButton.classList.remove('slider-prev-next-off');
    } else if (!prevButton.classList.contains('slider-prev-next-off')) {
        prevButton.classList.add('slider-prev-next-off');
    }
    if (currentSlide !== sliderItems.length - 1 && sliderItems.length !== 0) {
        nextButton.classList.remove('slider-prev-next-off');
    } else if (!nextButton.classList.contains('slider-prev-next-off')) {
        nextButton.classList.add('slider-prev-next-off');
    }
}

// Инициализация слайдера
const initSlider = (items) => {
    const dotsContainer = document.querySelector('.slider-dots');

    // обновление текущей позиции в соответствии с измененными данными в слайдере
    if (currentSlide >= sliderItems.length) {
        currentSlide = Math.max(0, sliderItems.length - 1);
        goToSlide(currentSlide)
    } else {
        currentSlide = Math.min(currentSlide + newHighlightCount, sliderItems.length - 1)
        goToSlide(currentSlide)
    }

    buttonControl()

    dotsContainer.innerHTML = Array.from({length: items.length}, (_, i) => 
        `<div class="slider-dot ${i === currentSlide ? 'active' : ''}" data-index="${i}"></div>`
    ).join('');
    
    // Оптимизация: делегирование событий
    dotsContainer.addEventListener('click', (e) => {
        const dot = e.target.closest('.slider-dot');
        if (dot) goToSlide(parseInt(dot.dataset.index));
    });
};

// Обновление слайдов
const updateSlider = (items) => {
    const slider = document.getElementById('slider');
    let realSlides;
    sliderItems = items;
    
    // Фиктивный слайд (левый)
    const dummyLeft = `
        <div class="slider-item-dummy">
            <p><strong>Время форматирования:</strong> 00:00</p>
            <p><strong>Шрифт:</strong> Times New Roman</p>
            <p><strong>Размер шрифта:</strong> 20</p>
            <p><strong>Выравнивание:</strong> По правому краю</p>
            <p><strong>Интервал:</strong> 3.0</p>
            <p><strong>Интервал перед абзацем:</strong> 3.0</p>
            <p><strong>Интервал после абзаца:</strong> 3.0</p>
            <p><strong>Отступ первой строки:</strong> 1.75</p>
            <p><strong>Табуляция в списках:</strong> 3.75</p>
        </div>
    `;
    // Формируем HTML для реальных слайдов
    if (!items || items.length === 0) {
        realSlides = `<div class="slider-empty-text">Пока ничего нет</div>`;
    } else {
        realSlides = items.map((item, index) => {
            // Если элемент входит в первые newHighlightCount, добавляем класс "new-highlight и slider-item-animation (для плавного появления)"
            const highlightClass = index < newHighlightCount ? " new-highlight slider-item-animation" : "";
            const animationClass = isInit === false && highlightClass === "" ? " slider-item-animation" : "";
            return `
                <div class="slider-item${highlightClass}${animationClass}" style="--index: ${index}">
                    <p><strong>Время форматирования:</strong> ${item.time}</p>
                    <p><strong>Шрифт:</strong> ${item.font}</p>
                    <p><strong>Размер шрифта:</strong> ${item.fontsize}</p>
                    <p><strong>Выравнивание:</strong> ${item.alignment}</p>
                    <p><strong>Интервал:</strong> ${item.spacing}</p>
                    <p><strong>Интервал перед абзацем:</strong> ${item.beforeSpacing}</p>
                    <p><strong>Интервал после абзаца:</strong> ${item.afterSpacing}</p>
                    <p><strong>Отступ первой строки:</strong> ${item.firstIndentation}</p>
                    <p><strong>Табуляция в списках:</strong> ${item.listTabulation}</p>
                </div>
            `;
        }).join('');
    }
    // Фиктивный слайд (правый)
    const dummyRight = `<div class="slider-item-dummy"></div>`;
    
    // Объединяем: левый dummy, реальные слайды, правый dummy
    slider.innerHTML = dummyLeft + realSlides + dummyRight;

    // Для каждого нового слайда, помеченного классом "new-highlight", устанавливаем таймер и событие наведения
    const highlightedItems = slider.querySelectorAll('.slider-item.new-highlight');
    highlightedItems.forEach(item => {
        // Удаляем выделение по наведению
        item.addEventListener('mouseenter', () => {
            item.classList.remove('new-highlight');
        });
        // Таймер: через 15 секунд удаляем выделение, если оно осталось
        setTimeout(() => {
            item.classList.remove('new-highlight');
        }, 15000);
    });
};

// Навигация
const goToSlide = (index) => {
    const slider = document.getElementById('slider');

    currentSlide = Math.max(0, Math.min(index, sliderItems.length - 1));

    buttonControl()
    
    const sliderWidth = slider.offsetWidth;
    let slideWidth, leftPos;
    let gap = sliderWidth * 0.05;

    // Определяем параметры в зависимости от устройства
    if (isMobile) {
        // Для мобильных: 1 слайд = 100% ширины
        slideWidth = sliderWidth;
        leftPos = (currentSlide + 1) * (slideWidth + gap);
    } else {
        // Для десктопа: 3 слайда (30% + 5% gap)
        slideWidth = sliderWidth * 0.30;
        leftPos = (currentSlide + 1) * (slideWidth + gap) + (slideWidth / 2) - (sliderWidth / 2);
    }
    
    slider.scrollTo({
        left: leftPos,
        behavior: 'smooth'
    });
    
    document.querySelectorAll('.slider-dot').forEach((dot, i) => 
        dot.classList.toggle('active', i === currentSlide)
    );
};

// SSE обработчик
const connectSSE = () => {
    eventSource = new EventSource("/events");

    eventSource.onmessage = (e) => {
        try {
            const data = JSON.parse(e.data);
            const counterElement = document.getElementById('counter');

            // Сравниваем новый счетчик с предыдущим
            const newCount = parseInt(data.count);
            if (newCount > prevCount && isInit === true) {
                newHighlightCount = newCount - prevCount;
            } else {
                newHighlightCount = 0;
            }
            prevCount = newCount;

            counterElement.textContent = `${data.count}`;

            // Анимация для счетчика
            counterElement.style.color = 'rgb(0, 113, 187)'
            
            setTimeout(() => {
                counterElement.style.transition = 'color 1.5s ease';
                counterElement.style.color = 'black';
            }, 300);
            counterElement.style.transition = 'none';
            

            updateSlider(data.last_formatting);
            initSlider(data.last_formatting);
            if (isInit === false) {
                goToSlide(0);
                const prevButton = document.querySelector('.slider-prev');
                const nextButton = document.querySelector('.slider-next');
                prevButton.style.visibility = 'visible';
                nextButton.style.visibility = 'visible';
                isInit = true
            }
            
        } catch (err) {
            console.error('Ошибка обработки данных:', err);
        }
    };
    eventSource.onerror = () => {
        eventSource.close();
        setTimeout(connectSSE, 3000);
    };
};

// Обработчики событий
document.querySelector('.slider-prev').addEventListener('click', () => goToSlide(currentSlide - 1));
document.querySelector('.slider-next').addEventListener('click', () => goToSlide(currentSlide + 1));

window.addEventListener('resize', () => {
    isMobile = window.innerWidth < breakpoint;
    goToSlide(currentSlide);
});

window.addEventListener('load', connectSSE);
window.addEventListener('beforeunload', () => eventSource?.close());
