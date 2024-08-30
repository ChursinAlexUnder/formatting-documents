// Получение ссылок на элементы DOM с помощью их id и классов
const documentInput = document.getElementById("documentInput");
const documentLabel = document.querySelector(".document-label");
const documentIcon = document.querySelector(".document-icon");
const documentText = document.querySelector(".document-text");
const documentNameSpan = document.querySelector(".document-name");

// Добавление слушателя события 'change' к элементу documentInput, который срабатывает при выборе файла
documentInput.addEventListener("change", function () {
  // Проверка, выбран ли хотя бы один файл
  if (documentInput.files.length > 0) {
    // Если файл выбран, установите текст элемента documentNameSpan на имя выбранного файла
    documentNameSpan.textContent = documentInput.files[0].name;
    // Скройте элементы documentText и documentIcon, если они используются
    documentText.style.display = "none";
    documentIcon.style.display = "none";
  } else {
    // Если файл не выбран, очистите текст элемента documentNameSpan
    documentNameSpan.textContent = "";
    // Верните видимость элементам documentText, если они используются
    documentText.style.display = "inline";
  }
});