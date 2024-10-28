// JavaScript для отображения прелоадера при отправке формы
document.getElementById("document-download-form").addEventListener("submit", function() {
    document.getElementById("preloader").style.display = "block"; // Показываем прелоадер
});

// Скрываем прелоадер при загрузке страницы, даже если страница была загружена из кеша
window.addEventListener("pageshow", function() {
    document.getElementById("preloader").style.display = "none"; // Скрываем прелоадер
});