const textarea = document.getElementById("text-input-id");
  // Добавить обработчик события keydown
  textarea.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      textarea.blur();  // Убирает фокус с textarea
    }
  });